# find-next

List every repo whose latest version probe reports an available upgrade.

## Synopsis

```
gitmap find-next [--scan-folder <id>] [--json]
gitmap fn        [--scan-folder <id>] [--json]
```

`fn` is the short alias.

## What it does

Joins the `Repo` table against the **newest** `VersionProbe` row per repo
(via correlated subquery on `MAX(ProbedAt)`) and filters to rows with
`IsAvailable = 1`. Sorted by `NextVersionNum DESC` so the freshest tags
float to the top.

This is **read-only** — it never re-runs the probe. To refresh stale
results first, run `gitmap probe --all`.

## Flags

| Flag | Effect |
|---|---|
| `--scan-folder <id>` | Restrict to repos discovered under one ScanFolder (see `gitmap sf list` for ids) |
| `--json` | Emit a JSON array instead of the human-readable summary, for CI consumption |

## Examples

```
$ gitmap find-next
Available updates (3):
  awesome-cli → v2.4.0 [method=ls-remote, probed=2026-04-19 06:12:01]
      E:\src\awesome-cli
  helper-lib  → v1.9.2 [method=ls-remote, probed=2026-04-19 06:12:03]
      E:\src\helper-lib
Hint: run `gitmap pull` or `gitmap cn next all` to apply.

$ gitmap fn --scan-folder 2 --json
[
  {
    "repo": { "id": 17, "slug": "awesome-cli", ... },
    "nextVersionTag": "v2.4.0",
    "nextVersionNum": 2004000,
    "method": "ls-remote",
    "probedAt": "2026-04-19 06:12:01"
  }
]
```

## See also

- `gitmap probe` — run the hybrid HEAD-then-clone version probe
- `gitmap sf list` — list scan-folder ids for the `--scan-folder` flag
- `gitmap pull` / `gitmap cn next all` — apply the available upgrades
