# startup-list (sl)

List user-scoped autostart entries created and managed by gitmap.

## Synopsis

```
gitmap startup-list
gitmap startup-list --format=json
gitmap startup-list --format=jsonl
gitmap startup-list --format=csv
gitmap sl --format=table
```

## Behavior

Scans the OS-appropriate autostart directory for files that satisfy
**both** the filename prefix gate AND an in-file marker:

| OS         | Directory                                              | File prefix | Extension | Marker                          |
|------------|--------------------------------------------------------|-------------|-----------|---------------------------------|
| Linux/Unix | `$XDG_CONFIG_HOME/autostart/` or `~/.config/autostart/`| `gitmap-`   | `.desktop`| `X-Gitmap-Managed=true` line    |
| macOS      | `~/Library/LaunchAgents/`                              | `gitmap.`   | `.plist`  | `<key>XGitmapManaged</key><true/>` |

Third-party autostart entries are silently ignored, even if their
filename happens to start with the gitmap prefix. The marker is the
real authority — the prefix is just a cheap pre-filter so we don't
have to open every unrelated file in the directory.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format: `table`, `json`, `jsonl`, or `csv` |
| `--json-indent` | `2` | Spaces per indent level for `--format=json`. `0` = minified single line. Range: 0..8. Ignored for non-json formats. |

`table` (alias: `terminal`) is the legacy human-readable rendering.
Bad `--format` values and out-of-range `--json-indent` both exit 2.
`--json-indent` is validated even when the format ignores it.

## Output formats

### `--format=table` (default)

```
Linux/Unix autostart entries managed by gitmap (/home/user/.config/autostart):
  • gitmap-sync-watcher  →  /usr/local/bin/gitmap watch ~/projects
  • gitmap-status-tray   →  /usr/local/bin/gitmap-tray

Total: 2 entry(ies). Remove one with: gitmap startup-remove <name>
```

A fresh user account with no autostart directory at all prints the
header followed by `(none — no gitmap-managed autostart entries found)`
and exits 0 — never an error.

### `--format=json`

Array of `{name, path, exec}` objects. Empty results render as `[]`
(never `null`) so `jq` pipelines work without conditionals. On macOS
the `exec` field is the space-joined `ProgramArguments` array (or
`Program` if `ProgramArguments` is absent).

```json
[
  {
    "name": "gitmap-sync-watcher",
    "path": "/home/user/.config/autostart/gitmap-sync-watcher.desktop",
    "exec": "/usr/local/bin/gitmap watch ~/projects"
  }
]
```

With `--json-indent=0` the same output collapses to one minified
line. Key order is identical at every indent (whitespace only).
The `[]\n` empty-list contract holds across all indent settings.

### `--format=jsonl`

One compact JSON object per line, terminated by `\n`. Same field
order as `--format=json` (name, path, exec). Empty results emit
**zero bytes** (NOT `\n`, NOT `[]`) so `wc -l` of the stream equals
the entry count exactly — the contract every line-oriented pipeline
(jq `--compact-output`, fluentd, BigQuery, DuckDB `read_json_auto`)
relies on.

```
{"name":"gitmap-sync-watcher","path":"/home/user/.config/autostart/gitmap-sync-watcher.desktop","exec":"/usr/local/bin/gitmap watch ~/projects"}
{"name":"gitmap-status-tray","path":"/home/user/.config/autostart/gitmap-status-tray.desktop","exec":"/usr/local/bin/gitmap-tray"}
```

### `--format=csv`

RFC 4180 CSV with a header row. Header is always written even when
there are zero entries so spreadsheet imports get a consistent shape.

```
name,path,exec
gitmap-sync-watcher,/home/user/.config/autostart/gitmap-sync-watcher.desktop,/usr/local/bin/gitmap watch ~/projects
```

## Platform notes

Linux/Unix and macOS are supported. On Windows the command exits
with a clear "unsupported OS" message — use the Windows startup
commands on that platform instead.

### macOS LaunchAgent caveats

- `startup-list` and `startup-remove` operate on the `.plist` file
  ONLY — they do NOT call `launchctl load/unload`. A removed plist
  takes effect at the next login or after a manual
  `launchctl unload <path>`. Intentional: `launchctl` requires a
  GUI session, making automated uninstall scripts brittle on CI.
- Binary plists are not supported. Gitmap-managed entries are
  always XML, so a binary plist with our prefix is treated as
  third-party and refused.
