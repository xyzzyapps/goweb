package main

import "testing"

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"https://github.com/xyzzyapps/goweb.git", "https://github.com/xyzzyapps/goweb"},
		{"https://github.com/xyzzyapps/goweb", "https://github.com/xyzzyapps/goweb"},
		{"git@github.com:xyzzyapps/goweb.git", "https://github.com/xyzzyapps/goweb"},
		{"ssh://git@github.com/xyzzyapps/goweb.git", "https://github.com/xyzzyapps/goweb"},
	}
	for _, tt := range tests {
		if got := normalizeRemoteURL(tt.in); got != tt.want {
			t.Errorf("normalizeRemoteURL(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestPagesURL(t *testing.T) {
	got := pagesURL("https://github.com/xyzzyapps/goweb")
	want := "https://xyzzyapps.github.io/goweb/"
	if got != want {
		t.Errorf("pagesURL = %q; want %q", got, want)
	}
}

func TestApplyGitConfigDefaultsDoesNotOverride(t *testing.T) {
	vars := map[string]string{
		"AUTHOR": "KeepMe",
		"EMAIL":  "keep@example.com",
		"REPO":   "https://example.com/repo",
	}
	applyGitConfigDefaults(vars)
	if vars["AUTHOR"] != "KeepMe" || vars["EMAIL"] != "keep@example.com" || vars["REPO"] != "https://example.com/repo" {
		t.Fatalf("overrode explicit vars: %#v", vars)
	}
}
