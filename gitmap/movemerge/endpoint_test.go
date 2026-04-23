package movemerge

import "testing"

func TestClassifyEndpoint_FolderPaths(t *testing.T) {
	cases := []string{"./local", "/abs/path", "..\\rel", "plain-folder"}
	for _, raw := range cases {
		kind, _, _, _ := ClassifyEndpoint(raw)
		if kind != EndpointFolder {
			t.Errorf("ClassifyEndpoint(%q) kind = %v, want Folder", raw, kind)
		}
	}
}

func TestClassifyEndpoint_HTTPSWithBranch(t *testing.T) {
	kind, url, branch, _ := ClassifyEndpoint("https://github.com/owner/repo:develop")
	if kind != EndpointURL {
		t.Fatalf("kind = %v, want URL", kind)
	}
	if url != "https://github.com/owner/repo" {
		t.Errorf("url = %q", url)
	}
	if branch != "develop" {
		t.Errorf("branch = %q", branch)
	}
}

func TestClassifyEndpoint_HTTPSNoBranch(t *testing.T) {
	_, url, branch, _ := ClassifyEndpoint("https://github.com/owner/repo.git")
	if url != "https://github.com/owner/repo.git" || branch != "" {
		t.Errorf("got url=%q branch=%q", url, branch)
	}
}

func TestClassifyEndpoint_SCPGitAtForm(t *testing.T) {
	// git@host:user/repo has a colon but it is not a branch.
	_, url, branch, _ := ClassifyEndpoint("git@github.com:owner/repo.git")
	if url != "git@github.com:owner/repo.git" || branch != "" {
		t.Errorf("scp form: url=%q branch=%q", url, branch)
	}
}

func TestMapURLToFolder(t *testing.T) {
	got := MapURLToFolder("/tmp", "https://github.com/owner/my-repo.git")
	if got != "/tmp/my-repo" {
		t.Errorf("MapURLToFolder = %q", got)
	}
}
