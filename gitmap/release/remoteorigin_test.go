package release_test

import (
	"testing"

	"github.com/alimtvnetwork/gitmap-v10/gitmap/release"
)

// TestParseHTTPSURL verifies HTTPS remote URLs are parsed correctly.
func TestParseHTTPSURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		owner string
		repo  string
	}{
		{"standard", "https://github.com/octocat/hello-world.git", "octocat", "hello-world"},
		{"no .git suffix", "https://github.com/octocat/hello-world", "octocat", "hello-world"},
		{"deep host", "https://gitlab.example.com/org/project.git", "org", "project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, err := release.ParseGitURLExported(tc.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tc.owner {
				t.Errorf("owner: expected %q, got %q", tc.owner, owner)
			}
			if repo != tc.repo {
				t.Errorf("repo: expected %q, got %q", tc.repo, repo)
			}
		})
	}
}

// TestParseSSHURL verifies SSH remote URLs are parsed correctly.
func TestParseSSHURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		owner string
		repo  string
	}{
		{"standard", "git@github.com:octocat/hello-world.git", "octocat", "hello-world"},
		{"no .git suffix", "git@github.com:octocat/hello-world", "octocat", "hello-world"},
		{"custom host", "git@gitlab.example.com:org/project.git", "org", "project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, err := release.ParseGitURLExported(tc.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tc.owner {
				t.Errorf("owner: expected %q, got %q", tc.owner, owner)
			}
			if repo != tc.repo {
				t.Errorf("repo: expected %q, got %q", tc.repo, repo)
			}
		})
	}
}

// TestParseInvalidURL verifies invalid URLs return errors.
func TestParseInvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty string", ""},
		{"plain text", "not-a-url"},
		{"ftp protocol", "ftp://github.com/owner/repo"},
		{"https too few parts", "https://github.com"},
		{"ssh no colon", "git@github.com"},
		{"ssh no slash", "git@github.com:noslash"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := release.ParseGitURLExported(tc.url)
			if err == nil {
				t.Error("expected error for invalid URL, got nil")
			}
		})
	}
}
