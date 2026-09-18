---
name: goweb
description: >
  Use goweb, a noweb-style literate programming tool on GitHub Flavored Markdown.
  Load when writing or editing <<chunk>>= sources, running tangle/weave/render/sync,
  scaffolding a literate program, or the user mentions goweb, literate markdown,
  noweb chunks, or /goweb.
---

# goweb

Binary: `goweb` (this repo: `go build -o goweb.exe ./cmd/goweb`). Language contract: `SPEC.md` at the repo root — do not invent syntax.

## When to use which command

| Goal | Command |
|------|---------|
| Extract source files | `goweb tangle [--var k=v ...] source.md` |
| One chunk to stdout | `goweb tangle source.md chunkname` |
| Clean markdown | `goweb weave source.md [-o out.md]` |
| HTML (lit-lang.org theme) | `goweb render source.md [-o out.html]` |
| Multi-page HTML | `goweb render --site dir/ --output-dir out/` |
| Edit tangled files → `.md` | `goweb tangle --line-directives` then `goweb sync source.md` |
| Scaffold | `goweb init [dir]` |
| Xref / DOT | `goweb index source.md` / `goweb graph source.md` |
| Wrap existing files | `goweb reverse a.go b.js > program.md` |
| Wrap a folder | `goweb reverse src/ -o program.md` |
| Fill `ai:open` holes | `goweb fill [--mock] source.md` |
| Freeze filled holes | `goweb seal source.md` |

Always pass `--var` the same way for tangle, weave, render, and sync.

If unset, these are filled from git config: `AUTHOR` ← `user.name`, `EMAIL` ← `user.email`, `REPO` ← `remote.origin.url` (normalized to https, no `.git`). `--var` wins.

## Write sources

```
<<name>>= file: path/out.js pipe: cmd exec: cmd session: s tags: a,b override ai:open
body may contain <<other>> and {{VAR}}
>>
```

- `>>` alone ends a chunk. Literal `>>` in a body is `\>>`.
- `<<name>>` is a reference (expanded at tangle; in weave/render it links to that chunk). `<<name>>=` is a definition.
- Same name: bodies concatenate unless `override` or `<<override "file.md">>`.
- `<<import "rel.md">>` inlines another file (paths relative to the importing file).
- `<<if expr>>` / `<<elif expr>>` / `<<else>>` / `<<end>>` — expr is `name`, `name==val`, or `name!=val`. Truthy unless unset, empty, `false`, `0`, or `no`. Bare `--var debug` ⇒ `true`.
- Unknown `{{VAR}}` is left unchanged.
- Only chunks with `file:` are written by a full tangle.

Split a program across `.md` files; one entry file imports the rest (see `examples/preact-todo/main.md`).

## Agent workflow

1. Edit the `.md` sources, not only the tangled output (unless using `sync`).
2. `goweb tangle --var ... entry.md` from the directory that should own output paths.
3. Run/tests on tangled files.
4. `goweb render --var ... entry.md -o docs.html` when docs must match.
5. Do not invent a second example tree; the shipped example is `examples/preact-todo/`.

Browser-served apps must tangle **`.js` (or other JS MIME) modules**, not `.tsx`/`.ts`. Static servers send those as `text/plain` and `type="module"` fails. Use an import map + `htm` (or a bundle) as in the example.

Default HTML theme is [lit-lang.org](https://lit-lang.org/) (amber `#ffcd42`, offset shadows, no RTD sidebar, no auto dark mode). Do not restyle unless asked.

## Pitfalls

- Cycles in `<<ref>>` graphs fail tangle.
- `sync` without `--line-directives` cannot map edits back.
- `--match tag` skips untagged `file:` chunks. Use `index`/`graph` as compact LLM context.
- `<<override>>` is a directive, never a `#chunk-override` link.
- `pipe:` / `exec:` need the command on `PATH` (`--pipe-dir` sets cwd).
- Opening `index.html` as `file://` breaks ES modules; use a static server.
- Do not commit `draft/` or `.todo/`. License is CC BY-SA 4.0.
- `goweb fill` only rewrites `ai:open` (and `ai:filled` with `--refill`). Never auto-seal. Tangle does not call a model.
