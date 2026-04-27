package cmd

// Cross-platform tests for parseStartupRemoveFlags. Validates that
// --backend and --dry-run parse independently, in any order, and
// that the positional name comes through cleanly. These tests are
// parser-only — they don't exercise startup.RemoveWithOptions, so
// they pass on every OS without touching the registry or
// filesystem.

import (
	"testing"
)

// TestParseStartupRemoveFlags_PositionalOnly is the baseline:
// no flags, just a name. Should return name + zero dryRun + empty
// backend (= unspecified / dual-backend fallback on Windows).
func TestParseStartupRemoveFlags_PositionalOnly(t *testing.T) {
	name, dryRun, backend, err := parseStartupRemoveFlags([]string{"foo"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != "foo" || dryRun || backend != "" {
		t.Errorf("got name=%q dryRun=%v backend=%q, want foo/false/\"\"",
			name, dryRun, backend)
	}
}

// TestParseStartupRemoveFlags_Backend confirms --backend=registry
// is captured into the returned string. ParseBackend is
// responsible for validating the value — this test only proves
// the flag wiring routes the value through.
func TestParseStartupRemoveFlags_Backend(t *testing.T) {
	_, _, backend, err := parseStartupRemoveFlags(
		[]string{"--backend=registry", "foo"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if backend != "registry" {
		t.Errorf("backend = %q, want registry", backend)
	}
}

// TestParseStartupRemoveFlags_BackendStartupFolder confirms the
// other valid backend value also routes through. Together with
// _Backend above, this proves the flag accepts both Windows
// backends without per-value special-casing.
func TestParseStartupRemoveFlags_BackendStartupFolder(t *testing.T) {
	_, _, backend, err := parseStartupRemoveFlags(
		[]string{"--backend=startup-folder", "foo"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if backend != "startup-folder" {
		t.Errorf("backend = %q, want startup-folder", backend)
	}
}

// TestParseStartupRemoveFlags_BackendAndDryRun confirms the two
// flags compose without interfering. Order of flags on the command
// line should not matter — flag.Parse handles arbitrary ordering
// of named flags, but positional args must come last.
func TestParseStartupRemoveFlags_BackendAndDryRun(t *testing.T) {
	name, dryRun, backend, err := parseStartupRemoveFlags(
		[]string{"--dry-run", "--backend=startup-folder", "myapp"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != "myapp" || !dryRun || backend != "startup-folder" {
		t.Errorf("got name=%q dryRun=%v backend=%q, want myapp/true/startup-folder",
			name, dryRun, backend)
	}
}

// TestParseStartupRemoveFlags_MissingName confirms a missing
// positional name surfaces as an error so the caller exits 2 with
// the usage message rather than silently no-op'ing.
func TestParseStartupRemoveFlags_MissingName(t *testing.T) {
	if _, _, _, err := parseStartupRemoveFlags([]string{"--dry-run"}); err == nil {
		t.Fatal("expected error for missing positional name, got nil")
	}
}
