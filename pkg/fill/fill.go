// Package fill fills and seals goweb chunks marked ai:open|filled|sealed.
// Tangle never calls this package.
package fill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xyzzyapps/goweb/pkg/parser"
	"github.com/xyzzyapps/goweb/pkg/preproc"
)

type Options struct {
	Mock     bool
	MockDir  string
	Model    string
	Name     string // only this chunk name; empty = all
	Refill   bool
	DryRun   bool
	Strict   bool
	EmptyOK  bool
	Context  bool
	Send     Completer // nil → HTTP unless Mock
}

type Completer func(user string) (string, error)

type hole struct {
	chunk  *parser.Chunk
	source string
}

func collect(entry string, vars map[string]string) ([]hole, error) {
	res, err := preproc.Process(entry, vars)
	if err != nil {
		return nil, err
	}
	var holes []hole
	for src := range res.Sources {
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		lines := splitKeep(string(data))
		doc, err := parser.ParseLines(lines, src)
		if err != nil {
			return nil, err
		}
		for _, c := range doc.Chunks {
			if c.AI == "" {
				continue
			}
			holes = append(holes, hole{chunk: c, source: src})
		}
	}
	return holes, nil
}

func splitKeep(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func Fill(entry string, vars map[string]string, opt Options) (int, error) {
	holes, err := collect(entry, vars)
	if err != nil {
		return 0, err
	}
	var work []hole
	for _, h := range holes {
		if opt.Name != "" && h.chunk.Name != opt.Name {
			continue
		}
		switch h.chunk.AI {
		case "open":
			work = append(work, h)
		case "filled":
			if opt.Refill {
				work = append(work, h)
			}
		}
	}
	if opt.DryRun {
		for _, h := range work {
			fmt.Fprintf(os.Stderr, "%s  <<%s>>= ai:%s\n", h.source, h.chunk.Name, h.chunk.AI)
		}
		return len(work), nil
	}
	mockDir := opt.MockDir
	if mockDir == "" {
		mockDir = filepath.Join(filepath.Dir(entry), "mock")
	}
	send := opt.Send
	if send == nil && !opt.Mock {
		ep, err := ResolveEndpoint(opt.Model)
		if err != nil {
			return 0, err
		}
		send = func(user string) (string, error) { return complete(ep, user) }
	}
	n := 0
	for _, h := range work {
		user := prompt(h, opt.Context)
		var body string
		if opt.Mock {
			b, err := os.ReadFile(filepath.Join(mockDir, h.chunk.Name+".txt"))
			if err != nil {
				return n, fmt.Errorf("mock %s: %w", h.chunk.Name, err)
			}
			body = string(b)
		} else {
			body, err = send(user)
			if err != nil {
				return n, err
			}
		}
		body = strings.TrimRight(body, "\n")
		if body == "" && opt.Strict {
			return n, fmt.Errorf("empty model output for %s", h.chunk.Name)
		}
		if body != "" && !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		if err := rewrite(h.source, h.chunk.Name, h.chunk.AI, "filled", body); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func Seal(entry string, vars map[string]string, opt Options) (int, error) {
	holes, err := collect(entry, vars)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, h := range holes {
		if opt.Name != "" && h.chunk.Name != opt.Name {
			continue
		}
		switch h.chunk.AI {
		case "open":
			if !opt.EmptyOK {
				return n, fmt.Errorf("cannot seal ai:open %q without --empty-ok", h.chunk.Name)
			}
			if !opt.DryRun {
				if err := rewrite(h.source, h.chunk.Name, "open", "sealed", ""); err != nil {
					return n, err
				}
			}
			n++
		case "filled":
			if !opt.DryRun {
				if err := rewrite(h.source, h.chunk.Name, "filled", "sealed", ""); err != nil {
					return n, err
				}
			}
			n++
		}
	}
	return n, nil
}

func prompt(h hole, ctx bool) string {
	var b strings.Builder
	b.WriteString("Fill the code hole. Return only the chunk body, no markdown fences, no explanation.\n")
	b.WriteString("chunk: " + h.chunk.Name + "\n")
	if h.chunk.Language != "" {
		b.WriteString("lang: " + h.chunk.Language + "\n")
	}
	if h.chunk.File != "" {
		b.WriteString("file: " + h.chunk.File + "\n")
	}
	b.WriteString("stub:\n")
	b.WriteString(h.chunk.Body)
	if ctx {
		data, err := os.ReadFile(h.source)
		if err == nil {
			s := string(data)
			if len(s) > 24000 {
				s = s[:24000]
			}
			b.WriteString("\n\nmodule context (truncated):\n")
			b.WriteString(s)
		}
	}
	return b.String()
}

// rewrite swaps ai:oldFlag → ai:newFlag on the header of name.
// If newBody is non-empty, the body (lines after header until >>) is replaced.
func rewrite(path, name, oldFlag, newFlag, newBody string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	nl := "\n"
	raw := string(data)
	if strings.Contains(raw, "\r\n") {
		nl = "\r\n"
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
	}
	lines := strings.Split(raw, "\n")
	// Drop trailing empty from Split on file ending with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	oldTok := "ai:" + oldFlag
	newTok := "ai:" + newFlag
	want := "<<" + name + ">>="
	changed := false
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, want) {
			continue
		}
		if !strings.Contains(line, oldTok) {
			continue
		}
		lines[i] = strings.Replace(line, oldTok, newTok, 1)
		if newBody != "" {
			end := i + 1
			for end < len(lines) && strings.TrimSpace(lines[end]) != ">>" {
				end++
			}
			body := strings.TrimSuffix(newBody, "\n")
			var bodyLines []string
			if body != "" {
				bodyLines = strings.Split(body, "\n")
			}
			out := append([]string{}, lines[:i+1]...)
			out = append(out, bodyLines...)
			if end < len(lines) {
				out = append(out, lines[end:]...)
			}
			lines = out
		}
		changed = true
		break
	}
	if !changed {
		return fmt.Errorf("could not rewrite <<%s>>= in %s", name, path)
	}
	var b strings.Builder
	for i, l := range lines {
		if i > 0 {
			b.WriteString(nl)
		}
		b.WriteString(l)
	}
	b.WriteString(nl)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
