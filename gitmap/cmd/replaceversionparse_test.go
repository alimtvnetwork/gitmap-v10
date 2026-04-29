package cmd

import "testing"

// TestSlugFromRemote covers every remote-URL shape we care about: HTTPS,
// SSH (git@host:path), bare slug, with and without trailing `.git`. The
// expected result is always the last path segment with `.git` trimmed.
func TestSlugFromRemote(t *testing.T) {
	cases := map[string]string{
		"https://github.com/alimtvnetwork/gitmap-v9.git": "gitmap-v9",
		"https://github.com/alimtvnetwork/gitmap-v9":     "gitmap-v9",
		"git@github.com:alimtvnetwork/gitmap-v9.git":     "gitmap-v9",
		"git@github.com:alimtvnetwork/gitmap-v9":         "gitmap-v9",
		"ssh://git@host.example/foo/bar/gitmap-v12.git":  "gitmap-v12",
		"gitmap-v3":     "gitmap-v3",
		"gitmap-v3.git": "gitmap-v3",
	}

	for in, want := range cases {
		got := slugFromRemote(in)
		if got != want {
			t.Errorf("slugFromRemote(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestRemoteSlugRegex documents what the version-suffix regex accepts
// and rejects. A failed match must return nil so detectVersion can
// emit the spec's clear "expected suffix -vN" error.
func TestRemoteSlugRegex(t *testing.T) {
	type want struct {
		matches bool
		base    string
		num     string
	}
	cases := map[string]want{
		"gitmap-v9":          {true, "gitmap", "7"},
		"my-tool-v123":       {true, "my-tool", "123"},
		"some-app-prefix-v0": {true, "some-app-prefix", "0"},
		"gitmap":             {false, "", ""},
		"gitmap-v":           {false, "", ""},
		"gitmap-vX":          {false, "", ""},
		"gitmap-v9-extra":    {false, "", ""},
	}
	for in, w := range cases {
		m := remoteSlugRe.FindStringSubmatch(in)
		if (m != nil) != w.matches {
			t.Fatalf("regex match for %q = %v, want %v", in, m != nil, w.matches)
		}
		if !w.matches {
			continue
		}
		if m[1] != w.base || m[2] != w.num {
			t.Errorf("regex %q -> base=%q num=%q, want base=%q num=%q",
				in, m[1], m[2], w.base, w.num)
		}
	}
}

// TestPairsForTarget locks the dual-form contract: every target produces
// both a `-vN` and a `/vN` replacement so Go module import paths and
// repo URLs are bumped in the same pass.
func TestPairsForTarget(t *testing.T) {
	got := pairsForTarget("gitmap", 4, 7)
	if len(got) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(got))
	}
	if got[0].old != "gitmap-v4" || got[0].new != "gitmap-v9" {
		t.Errorf("dash form wrong: %+v", got[0])
	}
	if got[1].old != "gitmap/v4" || got[1].new != "gitmap/v7" {
		t.Errorf("slash form wrong: %+v", got[1])
	}
}
