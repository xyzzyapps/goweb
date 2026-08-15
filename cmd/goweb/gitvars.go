package main

import (
	"os/exec"
	"strings"
)

// applyGitConfigDefaults fills AUTHOR, EMAIL, and REPO from git config
// when those keys were not set via --var.
func applyGitConfigDefaults(vars map[string]string) {
	if vars["AUTHOR"] == "" {
		if name := gitConfig("user.name"); name != "" {
			vars["AUTHOR"] = name
		}
	}
	if vars["EMAIL"] == "" {
		if email := gitConfig("user.email"); email != "" {
			vars["EMAIL"] = email
		}
	}
	if vars["REPO"] == "" {
		if url := gitConfig("remote.origin.url"); url != "" {
			vars["REPO"] = normalizeRemoteURL(url)
		}
	}
}

func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func pagesURL(repo string) string {
	repo = strings.TrimSuffix(repo, "/")
	const pfx = "https://github.com/"
	if !strings.HasPrefix(repo, pfx) {
		return repo
	}
	rest := strings.TrimPrefix(repo, pfx)
	owner, name, ok := strings.Cut(rest, "/")
	if !ok || owner == "" || name == "" {
		return repo
	}
	return "https://" + owner + ".github.io/" + name + "/"
}

func normalizeRemoteURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	if strings.HasPrefix(url, "git@") {
		// git@github.com:owner/repo → https://github.com/owner/repo
		rest := strings.TrimPrefix(url, "git@")
		if host, path, ok := strings.Cut(rest, ":"); ok {
			return "https://" + host + "/" + path
		}
	}
	if strings.HasPrefix(url, "ssh://git@") {
		rest := strings.TrimPrefix(url, "ssh://git@")
		return "https://" + rest
	}
	return url
}
