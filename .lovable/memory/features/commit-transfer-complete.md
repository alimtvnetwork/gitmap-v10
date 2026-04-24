---
name: Commit Transfer Complete
description: commit-left / commit-right / commit-both all live (v3.102.0). All three reuse committransfer.runOneDirection. commit-both = sequential L→R then R→L (interleave deferred).
type: feature
---

# Commit Transfer (Phases 1–3 complete)

## Status by phase

| Phase | Command | Shipped | Engine |
|-------|---------|---------|--------|
| 1 | `commit-right` (cmr) | v3.76.0 | `RunRight` (now thin wrapper over `runOneDirection`) |
| 2 | `commit-left` (cml) | v3.102.0 | `RunLeft` (swap source/target, then `runOneDirection`) |
| 3 | `commit-both` (cmb) | v3.102.0 | `RunBoth` = two sequential passes (L→R, then R→L) |

## commit-both algorithm (sequential, not interleaved)

The original spec proposed an author-date interleave. v3.102.0 ships a
simpler **two-pass sequential** model:

1. Pass 1: LEFT → RIGHT (build plan, prompt, replay, push)
2. Pass 2: RIGHT → LEFT (rebuild plan, prompt, replay, push)
3. Pass 1 failure aborts before Pass 2 (avoids merge-base drift)

Each pass appends a directional suffix to LogPrefix (`(left→right)` /
`(right→left)`) so output is visually attributable. Options struct is
copied per pass (`withDirectionLabel`) so the original LogPrefix is
never mutated — locked in by `TestRunBothImmutableOptions`.

## Files

- `gitmap/committransfer/runleftboth.go` — RunLeft, RunBoth, runOneDirection, withDirectionLabel
- `gitmap/committransfer/runleftboth_test.go` — direction-label + immutability tests
- `gitmap/committransfer/run.go` — RunRight (now a 1-line wrapper)
- `gitmap/cmd/committransfer.go` — `dispatchDirection` routes spec.Name → RunX
- `spec/01-app/106-commit-left-right-both.md` — §5 rewritten for sequential passes
- `gitmap/helptext/commit-{left,both}.md` — status updated, examples added

## Future: author-date interleave

Deferred. The sequential pass model covers the primary use case
(unified history on both sides). A future v3.x.0 may add
`--interleave` flag to commit-both for the original strict interleave.
