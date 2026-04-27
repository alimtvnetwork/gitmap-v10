# gitmap startup-add

Create a Linux/Unix XDG autostart entry that runs gitmap (or any
command) at login. The created `.desktop` file is tagged with
`X-Gitmap-Managed=true` so `startup-list` and `startup-remove` can
safely manage it without touching third-party autostart files.

## Alias

sa

## Usage

    gitmap startup-add --name <id> [--exec <path>] [--display-name <s>]
                       [--comment <s>] [--working-dir <path>]
                       [--backend registry|registry-hklm|startup-folder]
                       [--no-display] [--force]

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| --name           | yes | Logical name; filename becomes `gitmap-<name>.desktop` |
| --exec           | no  | Command to run at login (default: path to running gitmap binary) |
| --display-name   | no  | Override the `Name=` field shown in session managers |
| --comment        | no  | Optional `Comment=` text |
| --working-dir    | no  | Working directory the entry runs in (see *Working directory* below) |
| --no-display     | no  | Set `NoDisplay=true` (hide from app menus, still autostarts) |
| --force          | no  | Overwrite an existing **gitmap-managed** entry (never overwrites third-party files) |
| --backend        | no  | Windows only: `registry` (default, HKCU per-user), `registry-hklm` (HKLM machine-wide; **requires admin**), or `startup-folder` |
| --output         | no  | Output mode: `terminal` (default human lines) or `json` (status object — see below) |
| --json-indent    | no  | Spaces per indent level for `--output=json` (`0` = minified). Range 0..8. Ignored for terminal |

## `--output=json`

Emits a single-element JSON array containing one consistent status
object — the SAME shape `startup-remove --output=json` produces, so
a single jq filter handles both:

```json
[
  {
    "command": "startup-add",
    "action": "created",
    "name": "watch",
    "target": "/home/me/.config/autostart/gitmap-watch.desktop",
    "owner": "gitmap",
    "force_used": false,
    "dry_run": false
  }
]
```

- **`action`** — one of `created`, `overwritten`, `exists`, `refused`, `bad_name`.
- **`owner`** — `gitmap` (we own the entry), `third-party` (refused), or `unknown` (bad name).
- **`target`** — absolute file path or `HKCU\...` registry path; empty for `bad_name`.
- **`force_used`** — reflects whether `--force` was passed.
- **`dry_run`** — always `false` for `startup-add` (kept so add/remove records are rectangular).

Key order is byte-locked across Go versions (stablejson encoder).

## Working directory

`--working-dir <path>` records a directory the entry should run in.
The value is rendered differently per OS but is always read back by
`startup-list`:

- **Linux/Unix**: written as `Path=<dir>` in the `.desktop` file
  (XDG-spec field). The session manager `chdir`s here before
  invoking `Exec=`.
- **macOS**: written as `<key>WorkingDirectory</key>` in the
  LaunchAgent plist. `launchd` `chdir`s here before exec'ing
  `ProgramArguments`.
- **Windows**: stored as a `WorkingDir` REG_SZ value in the gitmap
  tracking subkey at `HKCU\Software\Gitmap\StartupRegistry\<name>`
  (registry backend) or `HKCU\Software\Gitmap\StartupFolder\<name>`
  (startup-folder backend). The autostart command itself (Run-key
  value or `.lnk` target) is unchanged — Windows reads cwd from the
  `.lnk` `WorkingDirectory` field, which the current minimal Shell
  Link writer does not yet emit; the tracking-subkey value is the
  source of truth for tooling.

Pass an absolute path. Relative paths are accepted as-is and
interpreted by the OS at login time. Omit the flag (or pass `""`)
to inherit whatever directory the login session provides.

## Windows backends

| `--backend`         | Hive | Scope             | Admin? | Path |
|---------------------|------|-------------------|--------|------|
| `registry` (default)| HKCU | Current user      | no     | `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\gitmap-<name>` |
| `registry-hklm`     | HKLM | Every user (machine-wide) | **yes** | `HKLM\Software\Microsoft\Windows\CurrentVersion\Run\gitmap-<name>` |
| `startup-folder`    | —    | Current user      | no     | `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\gitmap-<name>.lnk` |

The `registry-hklm` backend writes the autostart value under
`HKEY_LOCAL_MACHINE` so it fires for **every interactive user** on
the machine — useful for shared workstations, kiosks, and lab
images. Tracking metadata is mirrored at
`HKLM\Software\Gitmap\StartupRegistry\<name>` so `startup-list
--backend=registry-hklm` can attribute each entry to gitmap
without touching `HKCU`.

Writes (add / remove) require administrator privileges. The
process token is checked **before** any registry mutation; if you
are not elevated the command exits with a friendly:

    startup-add: --backend=registry-hklm requires administrator privileges
    (re-run from an elevated shell, e.g. `Run as administrator` from the
    Start menu, or use the per-user `--backend=registry` default)

`startup-list --backend=registry-hklm` and `startup-remove
--backend=registry-hklm --dry-run` are read-only and work for any
user (no elevation needed).

## Prerequisites

- Linux or other Unix with `~/.config/autostart` (XDG-compliant).
- macOS uses LaunchAgents — not handled here.
- Windows: any backend works for `--output=json`/terminal alike;
  `registry-hklm` additionally requires UAC elevation for writes.

## Safety

- Refuses to overwrite a `.desktop` file that does NOT carry the
  `X-Gitmap-Managed=true` marker, even with `--force`.
- Names containing path separators (`/`, `\`) or NUL are rejected
  before any I/O.
- Atomic write (temp file + rename) so a crash mid-write cannot
  leave a half-written file the next login session would execute.

## Examples

### Example 1: Add gitmap itself with default args

    gitmap startup-add --name watch --exec "$(command -v gitmap) watch"

**Output:**

    ✓ Created gitmap-managed autostart entry: /home/me/.config/autostart/gitmap-watch.desktop

### Example 2: Re-run is idempotent

    gitmap startup-add --name watch --exec "$(command -v gitmap) watch"

**Output:**

      (exists) gitmap-managed entry already at /home/me/.config/autostart/gitmap-watch.desktop — pass --force to overwrite

### Example 3: Update an existing entry

    gitmap startup-add --name watch \
      --exec "$(command -v gitmap) watch --quiet" --force

**Output:**

    ✓ Overwrote gitmap-managed autostart entry: /home/me/.config/autostart/gitmap-watch.desktop

## See Also

- [startup-list](startup-list.md) — List entries gitmap created
- [startup-remove](startup-remove.md) — Delete a gitmap-managed entry
