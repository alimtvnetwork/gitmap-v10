# Plan 05: Templates Polish — More Languages, `init`, `diff`

Spec: `spec/01-app/110-templates-polish.md` ✅ (v3.108.0)
Memory: `mem://features/templates-ignore-attributes` (extend), new `mem://features/templates-init-diff`
Builds on: Plan 04 (templates infrastructure, merge engine, pretty renderer)

## Goals

1. **Broader language coverage** — add `java`, `ruby`, `php`, `swift`, `kotlin` to both `ignore` and `attributes` corpora.
2. **`gitmap templates init`** — one-shot scaffold of `.gitignore` + `.gitattributes` (+ optional LFS) for a chosen language stack.
3. **`gitmap templates diff`** — show what `add` would change without writing (overlay vs embed, or current-file vs template).

All three pieces share the existing resolver / merge / marker-block primitives. No new infra, just composition + corpus.

## Open questions (resolve before Phase 0)

| Question | Default proposal |
|----------|------------------|
| Stack presets for `init` (single lang vs combos like `node+python`)? | Allow comma list: `--lang go,node`. Apply each in order; common always first. |
| `init` overwrite policy if files already exist? | Refuse unless `--force`; suggest `add` instead. Always allow `--dry-run`. |
| `diff` output format | Unified diff via `github.com/sergi/go-diff` OR hand-rolled marker-block diff (no new dep). Lean hand-rolled to keep dep surface small. |
| Should `diff` color via pretty renderer? | Yes, reuse `render.RenderANSI` tokens (`+` green, `-` red, `@` cyan) — TTY-aware via existing helper. |
| LFS in `init` | Off by default; opt-in with `--lfs`. Calls existing `add lfs-install` path. |
| Aliases | `ti` → `templates init`, `td` → `templates diff` |

## Phases

### Phase 0 — Spec + scaffolding

- [ ] Write `spec/01-app/110-templates-polish.md` covering: new langs, `init` UX, `diff` UX, exit codes, dry-run semantics.
- [ ] Decide and lock the open questions above.
- [ ] Add `mem://features/templates-init-diff` summary.
- [ ] No code yet.

### Phase 1 — Expand language corpus

- [ ] `assets/ignore/{java,ruby,php,swift,kotlin}.gitignore`
- [ ] `assets/attributes/{java,ruby,php,swift,kotlin}.gitattributes`
- [ ] Each file: `# source: github/gitignore@<sha-or-date>` + `# version: 1` header (per Plan 04 audit-trail rule).
- [ ] Update `constants_templates.go` lang enum + validation.
- [ ] Extend `corpus_test.go` to assert each file parses, has header, and is non-empty.
- [ ] Update `templates list` output to include the new langs (verify via `list_test.go`).

### Phase 2 — `templates init`

- [ ] `gitmap/cmd/templatesinit.go` — flags: `--lang <csv>`, `--lfs`, `--force`, `--dry-run`, `--cwd <path>`.
- [ ] Reuse `templates.Resolve` + `templates.Merge` from Plan 04 (no new merge logic).
- [ ] Order: common → each lang in CSV order → optional LFS attributes block.
- [ ] Refuse on existing non-empty `.gitignore`/`.gitattributes` unless `--force`.
- [ ] Idempotent: `init` then `add ignore <lang>` → no further changes.
- [ ] Helptext: `gitmap/helptext/templatesinit.md` (markdown, picked up by pretty renderer).
- [ ] Register alias `ti`.

### Phase 3 — `templates diff` ✅ (v3.108.0)

- [x] `gitmap/cmd/templatesdiff.go` — flags: `--lang <name>`, `--kind ignore|attributes` (default both), `--cwd <path>`.
- [x] `gitmap/templates/diff.go` — marker-block aware, pure (never writes).
  Status enum (NoChange / MissingFile / MissingBlock / BlockChanged) drives exit codes.
  Reuses `blockRegex(tag)` from `merge.go` so parser can't drift from writer.
  Hand-rolled (no Myers / no external diff dep) — block bodies are small enough
  that a flat removal-then-addition is honest and ≤180 LOC.
- [x] `diff_test.go` — 5 cases pinning all 4 branches + blank-line preservation.
- [x] TTY-aware coloring via `render.HighlightQuotesANSI` (cyan `+`, yellow `-`, dim `@@`).
- [x] Exit codes mirror `diff(1)`: `0` no change, `1` differences, `2` error.
- [x] Helptext `gitmap/helptext/templates-diff.md` with exit-code table + pre-commit example.
- [x] Alias `td` registered alongside `diff`.

### Phase 4 — Wiring + docs

- [ ] Register `init` and `diff` under `templatescli.go` dispatcher.
- [ ] Update `gitmap/helptext/templates.md` with the two new subcommands.
- [ ] Update `src/data/changelog.ts` (`v3.19.0` — new langs + init; `v3.20.0` — diff). Two version bumps if shipped sequentially, or one combined.
- [ ] README: short snippet under "Templates" section.
- [ ] Manual smoke matrix:
  - Empty repo → `templates init --lang go,node --lfs` → expect 3 files.
  - Existing `.gitignore` → refuse without `--force`.
  - `templates diff --lang python` after `add ignore python` → expect zero diff, exit 0.
  - `templates diff --lang java` on a repo missing the marker block → expect addition hunk, exit 1.

### Phase 5 — QA + release

- [ ] `go build ./... && go test ./...` clean.
- [ ] `golangci-lint run` clean (per `mem://tech/static-analysis-security`).
- [ ] Pretty-render the new helptexts on a TTY; confirm no token leakage.
- [ ] Tag release, update plan to `done`.

## Out of scope (defer)

- IDE-specific templates (VSCode, IntelliJ workspaces) — separate plan if requested.
- Template version upgrades / migration prompts — needs version bumping in `# version:` header system; defer.
- Custom user-defined templates discoverable from `~/.gitmap/templates/custom/` — interesting, but defer until a user asks.

## Sequencing recommendation

Ship Phase 1 (langs) standalone first as `v3.19.0` — low risk, immediate value. Then Phase 2+3 together as `v3.20.0`. Phase 0 spec covers both releases.
