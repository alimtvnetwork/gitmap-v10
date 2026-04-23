---
name: Clone multi-URL
description: gitmap clone accepts many URLs in one call — space-separated, comma-separated, or both. Each positional arg is split on commas and flattened into a single ordered list. Planned for v3.38.0.
type: feature
---

# Feature: `gitmap clone <url1> <url2,url3> ...` (v3.38.0)

**Spec:** `spec/01-app/104-clone-multi.md`
**Site route:** `/clone-multi`
**Depends on:** existing direct-URL clone (`mem://features/clone-direct-url`)

## Behaviour summary

- **Both syntaxes accepted, mixable.** `gitmap clone a b c`, `gitmap clone a,b,c`, and `gitmap clone a,b c d,e` all work. Parser: for each positional arg, split on `,`, strip whitespace, drop empties, append to ordered list.
- **Dedup case-insensitively** with trailing `.git` normalised, preserving first-seen order.
- **Existing flags unchanged:** `--target-dir`, `--github-desktop`, `--workers` (default 4), `--stop-on-fail`.
- **`--github-desktop` registers each successful clone immediately** (not at the end of the batch).
- **Per-repo progress + final summary.** Continue on failure unless `--stop-on-fail`.

## Exit codes

| 0 | All cloned (and registered if requested) |
| 1 | One or more clones failed |
| 2 | Lock contention |
| 3 | All URLs invalid — nothing attempted |

## Implementation

| File | What changes |
|------|-------------|
| `cmd/clone.go` | Add `flattenURLArgs([]string) []string` at top of `runClone`; loop the result through the existing concurrent worker pool |
| `desktop/desktop.go` | No change — `RegisterRepo(absPath)` already exists |
| `constants/constants_clone.go` | New: `MsgCloneInvalidURLFmt`, `MsgCloneSummaryMultiFmt`, `MsgCloneRegisteredInline` |

## Why both syntaxes

User asked for "comma or a space" literally. Accepting both costs ~6 lines of parser code and makes copy-paste from issue trackers / CI configs trivial. Unix idiom favours space; comma-separated wins for single-string parameters.
