# gitmap commit-left

> **Status (v3.74.0):** scaffold only. See commit-right.md for the
> shared status note. Lands in spec §18 Phase 2 (after Phase 1).

Replay RIGHT's commits onto LEFT as a fresh, cleaned commit sequence.
The "-left" suffix names the **destination**, exactly like
`merge-left` writes files to LEFT.

## Alias

cml

> Spec §13 reserved `cl`, but `cl` is already taken by `changelog`.
> Use `cml` instead. The long-form `commit-left` always works.

## Usage

    gitmap commit-left LEFT RIGHT [flags]

Same endpoint syntax and flag set as `commit-right` (see
[commit-right.md](commit-right.md)). The only difference is direction:
RIGHT → LEFT instead of LEFT → RIGHT.

## Examples (planned)

    gitmap commit-left ./repo-A ./repo-B

Replays the commits **from `./repo-B`** onto `./repo-A`.

## See Also

- [commit-right](commit-right.md) — opposite direction (full flag table)
- [commit-both](commit-both.md) — bidirectional
- [merge-left](merge-left.md) — file-state mirror (no commit replay)
- spec/01-app/106-commit-left-right-both.md — full design
