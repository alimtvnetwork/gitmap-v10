package cmd

// shouldRunBatch decides whether `gitmap cn` should fan out across multiple
// repos. Three triggers, evaluated in priority order:
//
//  1. Explicit `--csv <path>` flag.
//  2. Explicit `--all` flag.
//  3. Implicit: cwd has no `.git` entry but at least one git-repo subdirectory.
//
// Trigger #3 is intentionally cheap — it does NOT walk the tree here; it
// just checks the cwd's own .git presence. The walker re-runs in
// runCloneNextBatch, where ErrBatchEmpty will surface if the implicit
// trigger fired but no repos were found one level down.
func shouldRunBatch(flags CloneNextFlags) bool {
	if len(flags.CSVPath) > 0 || flags.All {
		return true
	}

	return false
}
