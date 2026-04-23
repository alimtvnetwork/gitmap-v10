# gitmap commit-both

> **Status (v3.74.0):** scaffold only. See commit-right.md for the
> shared status note. Lands in spec §18 Phase 3 (after Phases 1 & 2).

Bidirectional commit replay: each side ends up with the chronologically-
sorted union of both sides' commit timelines.

## Alias

cmb

> Spec §13 reserved `cb`. `cb` is currently free, but the family uses
> `cmb` for visual consistency with `cml` / `cmr`. The long-form
> `commit-both` always works.

## Usage

    gitmap commit-both LEFT RIGHT [flags]

## Algorithm (spec §5)

1. Compute LEFT-only commits (`base..LEFT-HEAD`) and RIGHT-only commits
   (`base..RIGHT-HEAD`) independently.
2. Concatenate both lists and sort by **author date ascending** (ties
   broken by source side: LEFT first).
3. Walk the merged sequence. For each entry, replay it onto the
   **opposite side** using the manual-reconstruct mechanism.
4. After the loop, both sides contain the same chronological union of
   commits. The original commits on each side remain (unchanged SHAs);
   the new commits are appended on top of the current branch.

Same flag set as `commit-right` (see
[commit-right.md](commit-right.md)).

## Examples (planned)

    gitmap commit-both ./repo-A ./repo-B

Output skeleton:

    [commit-both] LEFT-only: 3 commits, RIGHT-only: 2 commits
    [commit-both] interleaving by author date ...
    [commit-both] [1/5] a3f2c1d (LEFT) → RIGHT  feat: add OAuth flow
    [commit-both] [2/5] b7e4a9f (RIGHT) → LEFT  fix: typo
    ...
    [commit-both] done: replayed 5, skipped 0

## See Also

- [commit-left](commit-left.md), [commit-right](commit-right.md) — single-direction siblings
- [merge-both](merge-both.md) — file-state mirror (no commit replay)
- spec/01-app/106-commit-left-right-both.md — full design
