# docs-site Missing After `run.ps1` / `run.sh` Deploy and `install.*` Install

**Status:** Fixed in v2.84.0
**Affects:** `gitmap help-dashboard` (`gitmap hd`) on every install path

---

## Symptom

```
PS J:\...> gitmap hd
  ✗ Docs site directory not found at E:\bin-run\docs-site
    (operation: resolve, reason: directory does not exist)
```

The error reproduces in **all three install paths**:

1. Local source build → `run.ps1` deploy to `$env:DeployPath\gitmap\`
2. Local source build → `run.sh` deploy to `$DEPLOY_TARGET/gitmap/`
3. Remote release install → `install.ps1` / `install.sh` to `$LOCALAPPDATA\gitmap` / `~/.local/bin`

---

## Root Cause

`gitmap help-dashboard` resolves the docs folder relative to the binary directory
(`resolveBinaryDir()` in `cmd/helpdashboard.go`), expecting:

```
<binary-dir>/
  gitmap.exe
  docs-site/        ← required
    dist/           ← preferred (static mode)
```

The release pipeline **does** bundle `docs-site.zip` as a separate release asset
(`release/workflowdocs.go` + `release/workflowfinalize.go:45-50`). The defect was
in the **deploy and install scripts**:

| Script | Defect |
|---|---|
| `run.ps1` `Deploy-Binary` | Copied `data/` only — never `docs-site/` |
| `run.sh` `deploy_binary` | Copied `data/` only — never `docs-site/` |
| `install.ps1` `Main` | Downloaded the binary archive only — never `docs-site.zip` |
| `install.sh` `main`    | Downloaded the binary archive only — never `docs-site.zip` |

Auto-extract logic in `cmd/helpdashboard.go:34-44` would have rescued the
install scripts if `docs-site.zip` were placed next to the binary, but the
installers never downloaded that asset.

---

## Fix

### 1. `run.ps1` — new `Copy-DocsSite` helper

Called from `Deploy-Binary` after `data/` is copied:

- Prefers `<RepoRoot>/docs-site/dist/` (small, no `node_modules`) → copies to
  `<appDir>/docs-site/dist/`.
- Falls back to copying the full `docs-site/` source (excluding
  `node_modules`) so the npm-dev fallback in `serveDev` still works.
- Logs a clear warning if `docs-site/` is absent from the repo.

### 2. `run.sh` — new `copy_docs_site` helper

Same logic, mirrored for Bash. Excludes `node_modules` via
`find -mindepth 1 -maxdepth 1 ! -name node_modules`.

### 3. `install.ps1` — new `Install-DocsSite` step

After `Install-Binary`, downloads
`https://github.com/<repo>/releases/download/<version>/docs-site.zip` and
expands it directly into `$installDir` (the zip is layout-prefixed with
`docs-site/dist/...`, so it lands at `$installDir\docs-site\dist\`).

**Best-effort:** silently skips if the asset is absent (e.g. older releases
where `release/workflowdocs.go` was not yet wired in).

### 4. `install.sh` — new `install_docs_site` step

Same flow using `unzip`. Emits a clear error if `unzip` is missing on the
host instead of failing silently.

---

## Verification Checklist

After bumping to v2.84.0 and publishing a release:

- [ ] `run.ps1` deploy logs `Copied docs-site/dist to gitmap app directory`
- [ ] `<deploy>\gitmap\docs-site\dist\index.html` exists
- [ ] `gitmap hd` opens `http://localhost:5173` with the static dist served
- [ ] `install.ps1` (one-liner) logs `Installed docs-site to <installDir>\docs-site`
- [ ] `install.sh` mirrors the same on Linux/macOS
- [ ] Releases without `docs-site.zip` print `skipping (gitmap hd may not work)`
      and **do not** fail the install

---

## Related Files

- `run.ps1` — `Copy-DocsSite`, called in `Deploy-Binary`
- `run.sh` — `copy_docs_site`, called in `deploy_binary`
- `gitmap/scripts/install.ps1` — `Install-DocsSite`, called in `Main`
- `gitmap/scripts/install.sh` — `install_docs_site`, called in `main`
- `gitmap/cmd/helpdashboard.go` — auto-extract fallback (unchanged)
- `gitmap/release/workflowdocs.go` — release-side bundling (unchanged)

---

## Contributors

- AI-assisted audit and four-script coordinated fix
