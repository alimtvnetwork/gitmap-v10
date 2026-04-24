package cmd

import (
	"path/filepath"
	"sort"
	"testing"
)

// TestNormalizeExtList pins the public contract of `--ext` parsing:
// trim spaces, lowercase, prepend a dot when missing, drop empty
// entries, deduplicate. Empty input yields nil so the walker can
// short-circuit the no-filter path.
func TestNormalizeExtList(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single dotted", ".go", []string{".go"}},
		{"single bare", "go", []string{".go"}},
		{"mixed case", ".Go,MD", []string{".go", ".md"}},
		{"with spaces", "  .go , md ", []string{".go", ".md"}},
		{"dedup", ".go,go,.GO", []string{".go"}},
		{"drops empties and lone dot", ".,,.md, ", []string{".md"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeExtList(tc.in)
			if !equalStringSlice(got, tc.want) {
				t.Fatalf("normalizeExtList(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

// TestMatchesExtFilter covers the four combinations the walker cares
// about: no filter, matching ext (case-insensitive), non-matching ext,
// and a file with no extension at all.
func TestMatchesExtFilter(t *testing.T) {
	if !matchesExtFilter("/x/foo.go", nil) {
		t.Error("nil filter must match every file")
	}
	if !matchesExtFilter("/x/foo.GO", []string{".go"}) {
		t.Error("uppercase extension must match lowercase allow-list")
	}
	if matchesExtFilter("/x/foo.txt", []string{".go", ".md"}) {
		t.Error(".txt should not match {.go,.md}")
	}
	if matchesExtFilter("/x/Makefile", []string{".go"}) {
		t.Error("file with no extension must not match a non-empty filter")
	}
}

// TestParseReplaceFlagsExt drives the flag parser end-to-end with the
// `--ext` flag in both space-separated and `=`-joined forms, plus a
// positional sandwich, to prove splitReplaceFlagsAndArgs hands the
// value through to flag.Parse instead of treating it as positional.
func TestParseReplaceFlagsExt(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"space form", []string{"--ext", ".go,.md", "old", "new"}, []string{".go", ".md"}},
		{"equals form", []string{"--ext=.go,.md", "old", "new"}, []string{".go", ".md"}},
		{"interleaved", []string{"old", "--ext", "go", "new"}, []string{".go"}},
		{"missing flag", []string{"old", "new"}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, pos, err := parseReplaceFlags(tc.args)
			if err != nil {
				t.Fatalf("parseReplaceFlags: %v", err)
			}
			if !equalStringSlice(opts.exts, tc.want) {
				t.Errorf("opts.exts = %v, want %v", opts.exts, tc.want)
			}
			if len(pos) != 2 || pos[0] != "old" || pos[1] != "new" {
				t.Errorf("positional = %v, want [old new]", pos)
			}
		})
	}
}

// TestWalkRepoFilesHonorsExtFilter seeds a repo with multiple text
// extensions and verifies the walker only returns the allow-listed
// ones. Excluded directories and binary files must still be skipped
// regardless of their extension.
func TestWalkRepoFilesHonorsExtFilter(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "README.md"), []byte("doc\n"))
	mustWriteFile(t, filepath.Join(root, "src", "app.go"), []byte("package app\n"))
	mustWriteFile(t, filepath.Join(root, "src", "notes.txt"), []byte("ignored by filter\n"))
	mustWriteFile(t, filepath.Join(root, "CHANGELOG.MD"), []byte("upper-case ext\n"))
	mustWriteFile(t, filepath.Join(root, ".git", "HEAD"), []byte("ref:\n"))

	got, err := walkRepoFiles(root, []string{".go", ".md"})
	if err != nil {
		t.Fatalf("walkRepoFiles: %v", err)
	}

	rels := relativizeAll(t, root, got)
	sort.Strings(rels)

	want := []string{"CHANGELOG.MD", "README.md", "src/app.go"}
	if !equalStringSlice(rels, want) {
		t.Fatalf("walkRepoFiles returned %v, want %v", rels, want)
	}
}
