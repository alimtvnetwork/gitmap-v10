
# Plan: Install System Overhaul + README Redesign

> Status legend: ✅ Done · 🔄 In Progress · ⏳ Pending · 🚫 Blocked

## v3.39.0 Release-Version Script (spec 105) — 2026-04-21
- ✅ Authored `spec/01-app/105-release-version-script.md` (full contract: URL flow, generic + snapshot artefacts, flags, exit codes, missing-version interactive flow)
- ✅ Authored `.lovable/memory/features/release-version-script.md` and indexed
- ⏳ Implement `gitmap/scripts/release-version.ps1` (generic, parameterized, embedded via `go:embed`)
- ⏳ Implement `gitmap/scripts/release-version.sh` (bash counterpart, identical contract)
- ⏳ Add `constants/constants_install.go` entries: `ScriptReleaseVersionPS1`, `ScriptReleaseVersionSh`, snapshot filename format `release-version-vX.Y.Z.{ps1,sh}`
- ⏳ Wire snapshot generation into `cmd/release.go` release pipeline (copy generic, prepend `$Version = '<tag>'`, upload as release asset alongside binaries + checksums)
- ⏳ Update `src/pages/Release.tsx` to render TWO install boxes per `/release/:version` page: pinned (snapshot URL) and generic (`-Version` parameter form)
- ⏳ Confirm front-page `install.ps1` (latest-resolving) is **untouched** — out of scope
- ⏳ Add Vitest coverage for snapshot generator: input version → script body has `$Version = 'vX.Y.Z'` at line 1
- ⏳ Add Go test for missing-version interactive flow (mock GitHub API 404 + simulated TTY)
- ⏳ Add CHANGELOG v3.39.0 entry + bump `Version` constant
- 🚫 Decision needed before implementation: confirm GitHub release asset upload is automated (release pipeline) or manual (release notes checklist)


## v3.12.1 Session Snapshot (2026-04-20)
- ✅ Migrated all stale `Draft`/`PreRelease` field references to `IsDraft`/`IsPreRelease` (`release/metadata_test.go`, `tests/release_test/skipmeta_test.go`)
- ✅ Fixed `cmd/probe.go` `go vet` non-constant format string error
- ✅ Implemented `TestTopLevelCmdRegistryMatchesAST` AST parity test
- ✅ Cross-linked uniqueness CI guard from `spec/01-app/02-cli-interface.md` and `38-command-help.md`
- ✅ Bumped `Version` constant → `3.12.1`; added CHANGELOG v3.12.1 entry
- ✅ v15 legacy compat shim audit — KEEP through v3.x, remove in v4.0.0 (`mem://02-v15-legacy-compat-audit`)
- ✅ Generated fresh 28-table ERD `spec/01-app/gitmap-database-erd-v3.12.1.mmd`
- ⏳ Run `.\run.ps1` then `go test ./...` end-to-end build/test sweep
- ⏳ Tag and publish v3.12.1 GitHub release
- ⏳ Author `spec/01-app/v4-breaking-change-matrix.md`
- ⏳ Audit `migrate_v15phase4.go` for v4.0 removal schedule
- ⏳ Promote new ERD to canonical (delete stale ERDs, rename to `gitmap-database-erd.mmd`)
- ⏳ Add CI test for ERD ↔ `SQLCreate*` parity

## v3.0.0 Session Snapshot (2026-04-19)
- ✅ `as` / `release-alias` / `release-alias-pull` shipped with auto-stash + label-match pop
- ✅ `db-migrate` shipped + auto-invoked from `gitmap update`
- ✅ Marker-comment generator refactor (`// gitmap:cmd top-level` / `// gitmap:cmd skip`)
- ✅ CI `generate-check` drift detection
- ✅ Spec `spec/01-app/98-as-and-release-alias.md` authored (matches 97-move-and-merge format)
- ✅ CHANGELOG v3.0.0 entry + Migration guide block for constants contributors
- ✅ Docs layout shows `v3.0.0` badge (`src/components/docs/DocsLayout.tsx`)
- ⏳ Centralize `VERSION` constant in `src/constants/index.ts`
- ⏳ Add version badge to `Index.tsx` landing page hero
- ⏳ Add `## Migration guide` link to docs sidebar
- ⏳ Lint rule for missing `// gitmap:cmd top-level` markers in `constants/*.go`
- ⏳ Integration test for `release-alias` auto-stash round-trip

## Guardrail: Go Refactor Validation
- After any Go file split or refactor, run `go test ./<affected-package>` before marking the work done.
- Treat unused imports and stale references as blocking regressions, not cleanup for later.
- For install-flow changes under `gitmap/cmd`, verify `go test ./cmd` and `go vet ./cmd` before finalizing.

## Guardrail: Installer Output Contract
- Every installer flow must end with a visible summary showing installed version, binary path, install directory, and PATH target/status.
- Unix installers must print which shell/profile file received the PATH entry and how to reload it.
- Unix installers must explicitly warn that OTHER shells (sh, bash, fish) will NOT have gitmap unless the user manually adds the PATH line to those shells' profiles too.
- Windows installers must print whether User PATH was updated or already present.
- PowerShell installers must show the installed version and binary path.

## Part A: README Redesign (styled after scripts-fixer-v5)
1. **Center-aligned header** with badges, tagline, and horizontal rules
2. **Quick Start** section at the top (one-liner install + first scan)
3. **Clean grouped tables** with consistent formatting (ID-based like scripts-fixer-v5)
4. **Installation section** with all variants (one-liner, pinned version, custom dir, Linux/macOS)
5. **Project Structure** tree view section

---

## Part B: Expand Supported Tools (from scripts-fixer-v5)

### New tools to add to `gitmap install`:

**Core Tools (already have):** vscode, node, yarn, bun, pnpm, python, go, git, git-lfs, gh, github-desktop, cpp, php, powershell

**New tools to add:**
| Tool | Keyword | Choco Package | Winget Package | Apt Package | Brew Package | Snap Package |
|------|---------|---------------|----------------|-------------|-------------|-------------|
| MySQL | `mysql` | `mysql` | — | `mysql-server` | `mysql` | — |
| MariaDB | `mariadb` | `mariadb` | — | `mariadb-server` | `mariadb` | — |
| PostgreSQL | `postgresql` | `postgresql` | — | `postgresql` | `postgresql` | — |
| SQLite | `sqlite` | `sqlite` | — | `sqlite3` | `sqlite` | — |
| MongoDB | `mongodb` | `mongodb` | — | `mongod` | `mongodb-community` | — |
| CouchDB | `couchdb` | `couchdb` | — | `couchdb` | `couchdb` | `couchdb` |
| Redis | `redis` | `redis-64` | — | `redis-server` | `redis` | `redis` |
| Cassandra | `cassandra` | — | — | `cassandra` | `cassandra` | — |
| Neo4j | `neo4j` | `neo4j-community` | — | — | `neo4j` | — |
| Elasticsearch | `elasticsearch` | `elasticsearch` | — | `elasticsearch` | `elasticsearch` | — |
| DuckDB | `duckdb` | `duckdb` | — | — | `duckdb` | — |
| Chocolatey | `chocolatey` | (self) | — | — | — | — |
| Winget | `winget` | — | (self) | — | — | — |

---

## Part C: SQLite Installation Tracking (New DB Table)

### 1. New `InstalledTools` table schema:
```sql
CREATE TABLE IF NOT EXISTS InstalledTools (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Tool TEXT NOT NULL,
    VersionMajor INTEGER NOT NULL DEFAULT 0,
    VersionMinor INTEGER NOT NULL DEFAULT 0,
    VersionPatch INTEGER NOT NULL DEFAULT 0,
    VersionBuild INTEGER NOT NULL DEFAULT 0,
    VersionString TEXT NOT NULL DEFAULT '',
    PackageManager TEXT NOT NULL DEFAULT '',
    InstalledAt TEXT NOT NULL DEFAULT '',
    UpdatedAt TEXT NOT NULL DEFAULT '',
    InstallPath TEXT NOT NULL DEFAULT '',
    UNIQUE(Tool)
);
```

### 2. New model: `model/installedtool.go`
- `InstalledTool` struct with all fields
- `ParseVersion(versionStr string) (major, minor, patch, build int)` — parse version strings like `20.11.1`, `3.12.4`, `1.23.5`
- `CompileVersionString(major, minor, patch, build int) string` — build `"1.2.3.4"` from parts
- `CompareVersions(a, b InstalledTool) int` — compare two versions (-1, 0, 1)

### 3. Store operations: `store/installedtools.go`
- `SaveInstalledTool(tool InstalledTool) error` — INSERT OR REPLACE
- `GetInstalledTool(name string) (InstalledTool, error)`
- `ListInstalledTools() ([]InstalledTool, error)`
- `RemoveInstalledTool(name string) error`
- `IsInstalled(name string) bool`

### 4. Post-install recording
After successful `installTool()`, detect the installed version and save a record to the DB with parsed version components.

---

## Part D: Multi-Platform Package Manager Resolution

### 1. Config-based default manager (`config.json`):
```json
{
  "install": {
    "defaultManager": "choco",
    "managers": {
      "windows": "choco",
      "darwin": "brew",
      "linux": "apt"
    }
  }
}
```

### 2. Resolution priority:
1. `--manager` CLI flag (explicit override)
2. `install.defaultManager` from config.json
3. Platform auto-detect:
   - **Windows** → Chocolatey (fallback: Winget)
   - **macOS** → Homebrew
   - **Linux** → apt (fallback: snap, dnf)

### 3. Add Snap package manager support:
- New `PkgMgrSnap = "snap"` constant
- `buildSnapCommand(pkg string) []string` → `["sudo", "snap", "install", pkg]`
- Snap package name mappings for databases (redis, couchdb, etc.)

### 4. Expand package name mappings:
- `resolveAptPackage(tool) string` — Ubuntu/Debian package names
- `resolveBrewPackage(tool) string` — Homebrew package/cask names  
- `resolveSnapPackage(tool) string` — Snap package names
- Each function has a complete mapping for all ~27 tools

---

## Part E: Uninstall Support

### 1. New `gitmap uninstall <tool>` command:
- Check if tool exists in `InstalledTools` DB
- Build uninstall command based on the package manager that was used to install
- Remove the DB record after successful uninstall

### 2. Uninstall command builders:
- `buildChocoUninstallCommand(pkg) []string` → `["choco", "uninstall", pkg, "-y"]`
- `buildWingetUninstallCommand(pkg) []string` → `["winget", "uninstall", pkg]`
- `buildAptUninstallCommand(pkg) []string` → `["sudo", "apt", "remove", "-y", pkg]`
- `buildBrewUninstallCommand(pkg) []string` → `["brew", "uninstall", pkg]`
- `buildSnapUninstallCommand(pkg) []string` → `["sudo", "snap", "remove", pkg]`

### 3. Flags:
- `--dry-run` — show command without executing
- `--force` — skip confirmation
- `--purge` — remove config files too (apt: `purge`, choco: `-x`)

---

## Part F: Install List/Status Enhancements

### 1. `gitmap install --list` improvements:
- Group tools by category (Core, Databases, Utilities)
- Show installed status from DB (✓/✗ indicator)
- Show installed version from DB

### 2. `gitmap install --status` (new flag):
- Show all tools from DB with version, manager, install date
- Highlight outdated packages (compare DB version vs detected version)

### 3. `gitmap install --upgrade <tool>` (new flag):
- Re-run install for an already-installed tool to upgrade it
- Update the DB record with new version

---

## Execution Order

| Phase | Steps | Files Changed |
|-------|-------|---------------|
| **Phase 1** | README redesign (centered badges, clean structure) | `README.md` |
| **Phase 2** | Add new database tool constants + package mappings | `constants_install.go`, `installtools.go` |
| **Phase 3** | Add `InstalledTools` DB table + model + store CRUD | `store/`, `model/`, migration |
| **Phase 4** | Wire post-install DB recording + version parsing | `cmd/install.go`, `cmd/installtools.go` |
| **Phase 5** | Add config-based manager resolution | `config.json` schema, `cmd/installtools.go` |
| **Phase 6** | Add Snap package manager support | `constants_install.go`, `installtools.go` |
| **Phase 7** | Add uninstall command | `cmd/uninstall.go`, constants, helptext |
| **Phase 8** | Enhanced `--list`, `--status`, `--upgrade` flags | `cmd/install.go` |
| **Phase 9** | Completion support for install/uninstall tool names | Shell scripts, completion handler |

Each phase is independently shippable and testable.

---

## Part G: Pending Task Workflow (Task-Based Deletion)

Spec: `spec/01-app/83-pending-task-workflow.md`
Prevention: `spec/02-app-issues/21-pending-task-durability.md`

### Rule
Every `os.Remove` / `os.RemoveAll` must be preceded by a `PendingTask` insert.
No silent loss of delete intent is acceptable.

### Phase 1 — Database Layer
| Step | Files |
|------|-------|
| Add `TaskType`, `PendingTask`, `CompletedTask` SQL to constants | `constants/constants_pending_task.go` |
| Add model structs | `model/pendingtask.go`, `model/tasktype.go` |
| Add store CRUD (insert, list, complete, fail, find) | `store/pendingtask.go`, `store/tasktype.go` |
| Add seed logic for TaskType (Delete, Remove) | `store/store.go` (Migrate) |
| Add create/drop to migration + reset | `store/store.go` |
| Run `go test ./store/...` | — |

### Phase 2 — Delete Workflow Integration
| Step | Files |
|------|-------|
| Wrap `clone-next --delete` removal in task flow | `cmd/clonenext.go` |
| Create helpers: `CreateTask`, `CompleteTask`, `FailTask` | `cmd/pendingtaskhelper.go` |
| Duplicate prevention (same type + path) | `store/pendingtask.go` |
| Run `go vet ./cmd` + `go test ./cmd` | — |

### Phase 3 — CLI Commands
| Step | Files |
|------|-------|
| Add `pending` command (list all pending tasks) | `cmd/pending.go` |
| Add `do-pending` / `dp` command (retry all) | `cmd/dopending.go` |
| Add `do-pending <id>` (retry single) | `cmd/dopending.go` |
| Route in dispatcher | `cmd/roottooling.go` |
| Add constants (commands, messages, errors) | `constants/constants_cli.go`, `constants/constants_pending_task.go` |

### Phase 4 — Help Integration
| Step | Files |
|------|-------|
| Create `helptext/pending.md` | `helptext/pending.md` |
| Create `helptext/do-pending.md` | `helptext/do-pending.md` |
| Add to root usage output | `cmd/rootusage.go` |
| Add to UI commands data | `src/data/commands.ts` |
| Update documentation site help page | `src/pages/` |

### Phase 5 — Validation & Edge Cases
| Step | Files |
|------|-------|
| Test missing folder retry | tests |
| Test permission failure | tests |
| Test duplicate prevention | tests |
| Test completed-task transactional move | tests |
| Run full `golangci-lint` | — |
