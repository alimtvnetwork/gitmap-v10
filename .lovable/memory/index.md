# Memory: index.md
Updated: now

# Project Memory

## Core
NO-QUESTIONS MODE ACTIVE (40-task budget): never ask user clarifying questions. Log every ambiguity to `.lovable/question-and-ambiguity/xx-brief-title.md` (sequential numbering, update `00-index.md` table). Each entry: original task + spec ref, ambiguity, all options w/ pros/cons, recommendation, decision taken. Proceed with best inference. Resume questions only when user says "ask question".
Strict code style: <200 lines/file, <15 lines/func, positive logic, pascal case constants, 'is/has' boolean prefixes.
Organize constants by owning package/domain; do not force artificial prefix-only naming rules.
Zero-swallow error policy. Explicitly log errors to os.Stderr using standardized format. Use `errors.Is`.
NEVER manually create, modify, or delete files within `.gitmap/release/` or `.gitmap/release-assets/`.
No magic strings. Centralize in constants. All CLI IDs must be exclusively in `constants_cli.go`.
Windows-first platform development strategy. Scripts must handle Windows encoding (UTF-8 BOM).
Go v1.24.13. golangci-lint pinned to v1.64.8, govulncheck pinned to v1.1.4.
SQLite connection pooling restricted to `SetMaxOpenConns(1)`.
Database schema uses strict PascalCase, INTEGER PRIMARY KEY AUTOINCREMENT.
Unified `.gitmap/` directory structure at repository root for all artifacts.
Clone-next flattens by default (v2.75.0+): clones into base name folder, tracks versions in RepoVersionHistory.
Clone-next `-f` / `--force` (v3.50.0+): chdir-to-parent before remove when cwd IS target folder; refuses versioned-folder fallback.
Completion generator uses marker-comment opt-in (v3.0.0+): `// gitmap:cmd top-level` on const block, `// gitmap:cmd skip` per spec. CI `generate-check` enforces drift.
VS Code Project Manager sync: resolve user-data root per OS first, then append `User/globalStorage/alefragnani.project-manager/projects.json` — never hardcode the full path.
Current version: v3.152.0.
Consumer-facing JSON outputs use `gitmap/stablejson` (key-by-key, no struct reflection) so field order cannot drift across Go versions or encoding/json/v2.
`gitmap cn` accepts folder-arg forms (v3.117.0+): `cn vX <folder>`, `cn v+1 <folder>`, `cn <folder>` (defaults v++). Dispatcher in `clonenextfolderdispatch.go` runs BEFORE alias dispatcher; uses path-hint + os.Stat heuristic. Hero card uses `--accent-success` semantic token (no hardcoded greens).
`gitmap clone <url>` cds into cloned folder via WriteShellHandoff (v3.118.0+) — single-URL only; multi-URL deliberately skips handoff.
`gitmap inject` / `inj` (v3.119.0+): register existing folder with Desktop + VS Code, conditional DB upsert (only if `git remote get-url origin` succeeds). cwd default + optional positional via `resolveCloneNextFolder`. Any folder accepted (no `.git/` check). WriteShellHandoff at end.
Site theme: `--primary` is amber gold (`38 92% 50%` light / `41 96% 56%` dark) — was blue. Hero card uses borderless `max-w-5xl` softer panel; install/uninstall sit side-by-side.
Templates Phase 1+2+3+4+5 complete (v3.108.0+): full 11-lang corpus, `add ignore`/`add attributes` (sorted-tag marker blocks), `templates list --kind/--lang`, `templates init`, `templates show` (pretty/raw), `templates diff` (alias `td`, standard diff(1) exit codes 0/1/2, block-scoped, TTY-aware coloring). Pretty renderer corpus at 9 fixtures.
Clone audit (v3.99.0+): `gitmap clone --audit <manifest>` is read-only; never invokes git, refuses direct URLs, prints diff-style markers (+/~/=/?/!).
Cross-platform install/update reference (v3.100.0+): canonical matrix at `spec/01-app/108-cross-platform-install-update.md`, mirrored on `/install-gitmap` page and linked from README top.
Clone parallel + hierarchy (v3.101.0+): `gitmap clone --max-concurrency N` opt-in parallel runner; default 1 = sequential. Hierarchy preserved at any N via filepath.Join(targetDir, rec.RelativePath).
Commit-transfer family complete (v3.102.0+): `commit-left` / `commit-right` / `commit-both` all wired through `committransfer.runOneDirection`. commit-both = sequential L→R then R→L (interleave deferred).
Shell handoff sentinel (v3.103.0+): `GITMAP_HANDOFF_FILE` temp-file pattern wired into clone-next/as/cd. Replaces broken `GITMAP_SHELL_HANDOFF` env var. Wrappers in `constants.CDFunc{Bash,Zsh,PowerShell}`.
Commit-both --interleave (v3.104.0+): author-date merged stream variant via `RunBothInterleaved`; sequential remains default. CLI guard exits 2 if --interleave passed to commit-left/commit-right.
Clone-pick / cpk (v3.153.0+, spec 100): sparse-checkout subset of repo into cwd. Auto-saves every run to `CloneInteractiveSelection` table; `--replay <id|name>` re-runs. `--ask` opens bubbletea tree picker. Short alias is `cpk` (NOT `ci` — collides with CI/CD).

## Memories
- [Code Constraints](mem://style/code-constraints) — Strict rules for code style, structure, and pull requests
- [Code Quality Process](mem://style/code-quality-improvement-process) — Architectural principles and resilience patterns
- [README Branding](mem://style/readme-branding) — Strict layout and linking requirements for the project author section
- [Windows Environment](mem://constraints/windows-environment) — Long paths, short root recommendations for git
- [PowerShell Encoding](mem://constraints/powershell-encoding) — ASCII punctuation, Virtual Terminal Processing, stdout vs stderr
- [Navigation Helper](mem://features/navigation-helper) — Shell wrapper using GITMAP_SHELL_HANDOFF for cd/clone-next
- [Command Help System](mem://features/command-help-system) — 120-line limit per help file, 3-8 line realistic simulations
- [Clone-Next Flatten](mem://features/clone-next-flatten) — Default flatten: clone into base-name folder, version tracking in DB with RepoVersionHistory table (DONE v2.75.0)
- [CN Find-Next Bridge](mem://features/cn-find-next-bridge) — PLANNED v3.55.0: `gitmap cn` no-args detects scope, auto-probes (spec 103, depth=5, parallel), interactive TUI picker, parallel updates. `find-next` stays read-only.
- [Clone Direct URL](mem://features/clone-direct-url) — gitmap clone accepts direct HTTPS/SSH URLs with optional folder name, auto-flattens versioned URLs
- [Clone Audit](mem://features/clone-audit) — `gitmap clone --audit <manifest>` plans+prints diff-style report (+/~/=/?/!) without invoking git (v3.99.0)
- [Clone Pick](mem://features/clone-pick) — `gitmap clone-pick`/`cpk`: sparse-checkout subset of a repo, optional `--ask` tree picker, auto-saves to `CloneInteractiveSelection` table, `--replay <id|name>` (spec 100, v3.153.0)
- [Cross-Platform Install/Update](mem://features/cross-platform-install-update) — Canonical Win/macOS/Linux install · update · uninstall · verify matrix at spec/01-app/108-cross-platform-install-update.md, mirrored at /install-gitmap (v3.100.0)
- [Clone Parallel + Hierarchy](mem://features/clone-parallel-hierarchy) — `gitmap clone --max-concurrency N` opt-in worker pool, hierarchy preserved at any N, thread-safe Progress + CloneCache (v3.101.0)
- [Shell Handoff File](mem://features/shell-handoff-file) — `GITMAP_HANDOFF_FILE` sentinel-file pattern wired into clone-next/as/cd; replaces broken `os.Setenv("GITMAP_SHELL_HANDOFF", ...)` (v3.103.0)
- [Move & Merge Commands](mem://features/movemerge) — gitmap mv / merge-both / merge-left / merge-right with L/R/S/A/B/Q prompt + --prefer-* bypass + URL-side commit/push (v2.96.0)
- [Release Alias](mem://features/release-alias) — gitmap as / release-alias (ra) / release-alias-pull (rap) with auto-stash labeled by alias-version-unixts, label-match pop for concurrent safety (v3.0.0)
- [Self Install Uninstall](mem://features/self-install-uninstall) — gitmap self-install / self-uninstall manage the binary itself (separate from third-party install/uninstall). Embedded scripts via go:embed, Windows handoff, marker-block PATH cleanup
- [Startup Management Unix](mem://features/startup-management-unix) — Linux/Unix `startup-list` (sl) + `startup-remove` (sr) for XDG autostart entries, scoped to `gitmap-` prefix + `X-Gitmap-Managed=true` marker (v3.133.0)
- [Stable JSON Encoding](mem://features/stable-json-encoding) — `gitmap/stablejson` package: key-by-key encoding, byte-compat with json.Encoder, used by `startup-list --format=json` (v3.152.0)
- [Replace Command](mem://features/replace-command) — gitmap replace literal "old" "new" / -N / --audit / all bumps `<base>-vN` and `<base>/vN` from git remote URL, interactive confirm before write, atomic temp+rename, binary skip (v3.96.0)
- [Marker Comments](mem://features/marker-comments) — Decentralized opt-in for completion generator: `// gitmap:cmd top-level` + `// gitmap:cmd skip`, CI drift check enforces sync (v3.0.0)
- [VS Code Project Manager Sync](mem://features/vscode-project-manager-sync) — gitmap scan auto-syncs and `gitmap code` registers + opens repos in alefragnani.project-manager projects.json (v3.38.0)
- [Database Architect](mem://tech/database-architecture) — Idempotent SQLite migrations, PascalCase schema helpers
- [Database Constraints](mem://tech/database-constraints) — Recursive reconciliation pattern, explicitly re-query database IDs
- [Database Location](mem://tech/database-location) — SQLite state anchored to binary execution path via filepath.EvalSymlinks
- [Process Sync](mem://tech/process-synchronization) — Advisory file-based locking via gitmap.lock
- [DB Migration Strategy](mem://tech/database-migration-strategy) — Graceful recovery for breaking schema changes, intercepting scan errors
- [Static Analysis](mem://tech/static-analysis-security) — Linter setup, vulnerability response times, @latest installations prohibited
- [Security Hardening](mem://tech/security-hardening) — Zip extraction path validation, io.LimitReader for decompression bombs
- [Changelog System](mem://project/changelog-system) — Dual-mode Markdown/React changelog synced with local release metadata
- [Flag Parsing Logic](mem://tech/flag-parsing-logic) — Reordering flags before args to bypass Go's default flag package limitations
- [Go Namespace Rules](mem://tech/go-namespace-constraints) — Preventing redeclaration across files in the same Go package
- [Vulnerability Mitigation](mem://tech/vulnerability-mitigation-strategy) — Bypassing GO-2026-4601 in Go 1.24 via custom http Request
- [Config Pattern](mem://tech/config-pattern) — Three-layer configuration merge (defaults < config.json < CLI flags)
- [Script Generation](mem://tech/script-generation) — PowerShell text/template encoding with UTF-8 BOM
- [Constants Structure](mem://tech/constants-structure) — Avoiding redeclaration errors with unique suffixes and domain-specific files
- [Constants Ownership](mem://constraints/constants-ownership) — Keep constants in the owning package/domain; avoid forced prefix-only naming rules
- [Code Red Error Mgmt](mem://tech/code-red-error-management) — Zero-swallow error policy and os.Stderr standardized format
- [Internal Memory Standard](mem://project/internal-memory-standard) — Folder structure and file naming conventions for project planning
- [Templates: ignore/attributes/pretty](mem://features/templates-ignore-attributes) — Embedded `.gitignore`/`.gitattributes` templates per language, idempotent marker-block merge, `~/.gitmap/templates/` overlay, `add ignore`/`add attributes`/`add lfs-install` subcommands, pretty markdown renderer with fixture corpus (Phase 0 scaffolded; spec 109, plan 04)
- [CI Hardening Session 2026-04-29](mem://workflow/04-ci-hardening-session) — Compile gate, GOMODCACHE/GOCACHE caching, gofmt -l, goimports@v0.24.0 with go.mod-derived `-local`, golangci-lint strict `--issues-exit-code=1`, regression-guard semantics audit (hard-floor vs baseline-diff)
