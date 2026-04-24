# gitmap startup version check

A one-line banner printed to **stderr** on every invocation that shows
the active gitmap binary version and warns when the typed command or
flag requires a newer version than what is actually installed.

## What you see

    $ gitmap clone https://github.com/foo/bar
    [gitmap v3.90.0]
    ... rest of normal clone output on stdout ...

When the command needs a newer binary:

    $ gitmap pending clear   # active binary is v3.85.0, command needs v3.88.0
    [gitmap v3.85.0]
      ⚠ pending clear requires gitmap v3.88.0 — active binary is v3.85.0.
        Run `gitmap update` to upgrade, or pass --no-version-check to silence this warning.

## Why stderr?

The banner is informational scaffolding. Putting it on **stderr** means
scriptable commands (`gitmap scan --output json | jq ...`) keep clean
stdout. The pipeline still sees the banner on the terminal.

## Suppressing the banner

Any one of these silences it:

| Method | When to use |
|--------|-------------|
| `--no-version-check` | Per-invocation, explicit opt-out |
| `--no-banner` | Re-uses the existing bare-invocation banner suppression flag |
| `GITMAP_QUIET=1` | Shell-wide silence (CI, scripts, hot loops) |
| Safe-list commands | `version`, `v`, `help`, `update`, `doctor`, `self-install`, `self-uninstall` always run silent so you can recover from a mismatch |

## How the warning is triggered

The check is **purely local** — no network call, no exec of the
deployed binary. Two pieces of information are compared:

1. **Active version** = `constants.Version` of the running binary
   (the binary you're invoking literally IS this build).
2. **Required version** = highest minimum from the `cmdMinVersions`
   map for the command + its flags + detected behaviours.

A behaviour is something like passing `;` as a separator in a
multi-URL clone, which only works in v3.89.0+ — even though the
flag list looks innocuous to an older binary's parser, the user is
relying on a feature it doesn't have.

## Adding a new requirement

Edit `gitmap/cmd/startupversioncheck.go` → `cmdMinVersions`:

```go
var cmdMinVersions = map[string]string{
    constants.CmdPending + " clear": "3.88.0",
    "clone:--no-replace":            "3.55.0",
    "your-new-command":              "3.91.0",
}
```

Keys: bare command name, `<cmd> <subcommand>`, `<cmd>:--<flag>`, or
`<cmd>:<behaviour-label>`. Append-only — never remove or downgrade
an entry, even if a feature later becomes default; users on older
binaries still need the warning.

## Exit codes

The check **never** changes the exit code or aborts the command.
It is advisory only — the user might be deliberately running an
older binary against a newer source repo, and we don't want to
break that workflow.

## See also

- `gitmap doctor` — full audit including PATH vs deployed vs source
- `gitmap version` — bare version, no banner
- `gitmap update` — fetch and install the latest version
