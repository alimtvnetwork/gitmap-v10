# Suggestions Tracker

## Completed Suggestions

- ✅ Add `direct-clone-ssh.ps1` output
- ✅ Implement copy-and-handoff for `gitmap update`
- ✅ Add deploy retry logic in `run.ps1`
- ✅ Document `version` command in specs
- ✅ Bump version on every code change
- ✅ Update all spec docs for new features
- ✅ Create `spec/03-general/` with reusable design guidelines
- ✅ Add `desktop-sync` command
- ✅ Enhanced terminal output with HTTPS and SSH clone instructions
- ✅ Remove GitHub Release integration
- ✅ Nested deploy structure
- ✅ Update enhancements: skip-if-current, version comparison, rollback safety
- ✅ `update-cleanup` command with auto-run
- ✅ Made all `spec/03-general/` files fully generic
- ✅ Full compliance audit (Wave 1 + Wave 2)
- ✅ Constants inventory documentation
- ✅ `list-versions` and `revert` commands
- ✅ Changelog in release metadata JSON
- ✅ Releases table in SQLite database
- ✅ PascalCase for all DB table/column names
- ✅ `seo-write` command with templates, CSV, rotation, and dry-run
- ✅ Unit test infrastructure with PowerShell runner (`run.ps1 -t`)
- ✅ `--compress`, `--checksums`, Go cross-compilation pipeline
- ✅ Config-driven release targets, checksums, and compress booleans
- ✅ Build documentation site with actual gitmap docs
- ✅ Add Linux/macOS support with cross-compile binary and CI/CD
- ✅ Add progress bar for clone
- ✅ **`--flatten` for `clone-next`** → Promoted to default behavior in v2.75.0 (no flag needed)
- ✅ **`gitmap clone <url>` auto-flatten** versioned URLs (v2.75.0)
- ✅ **`RepoVersionHistory` table** for tracking version transitions (v2.75.0)
- ✅ **`gitmap version-history` (`vh`) command** with `--limit`/`--json` (v2.76.0)
- ✅ **Database ERD** covering all 22 tables as Mermaid diagram (v2.76.0)
- ✅ **Spec updates** for flatten-by-default behavior (v2.76.0)
- ✅ **Tab completion** for `version-history`/`vh` (v2.76.0)
- ✅ **Docs site page** for version-history with terminal previews (v2.76.0)
- ✅ **`gitmap doctor setup`** checks: config resolution + wrapper detection (v2.74.0)
- ✅ **Shell wrapper `GITMAP_WRAPPER=1`** for raw binary vs wrapper detection (v2.74.0)
- ✅ **Post-setup verification** warns if shell function not loaded (v2.74.0)
- ✅ **VS Code admin-mode bypass** with 3-tier launch strategy (v2.72.0)

## Pending Suggestions

- 🔲 Add `version-history` to docs site sidebar/commands navigation
- 🔲 Add `clone` page to docs site (file-based + URL clone documentation)
- 🔲 Add `--dry-run` flag to `clone-next` for previewing actions without executing
- 🔲 Expand `install` command with database tools (MySQL, PostgreSQL, Redis, etc.)
- 🔲 Add `gitmap uninstall <tool>` command
- 🔲 Enhanced `install --list` grouped by category with installed status
- 🔲 Unit tests for task, env, and install commands
- 🔲 Update `helptext/env.md` with `--shell` flag usage examples
