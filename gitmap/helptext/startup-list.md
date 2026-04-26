# startup-list (sl)

List Linux/Unix XDG autostart entries created and managed by gitmap.

## Synopsis

```
gitmap startup-list
gitmap sl
```

## Behavior

Scans `$XDG_CONFIG_HOME/autostart/` (falling back to
`$HOME/.config/autostart/`) for `.desktop` files that satisfy
**both** conditions:

1. Filename starts with `gitmap-`
2. Body contains `X-Gitmap-Managed=true`

Third-party autostart entries are silently ignored, even if their
filename happens to start with `gitmap-`.

## Output

```
Linux/Unix autostart entries managed by gitmap (/home/user/.config/autostart):
  • gitmap-sync-watcher  →  /usr/local/bin/gitmap watch ~/projects
  • gitmap-status-tray   →  /usr/local/bin/gitmap-tray

Total: 2 entry(ies). Remove one with: gitmap startup-remove <name>
```

A fresh user account with no autostart directory at all prints the
header followed by `(none — no gitmap-managed autostart entries found)`
and exits 0 — never an error.

## Platform notes

Linux/Unix only. On Windows or macOS the command exits with a clear
"Linux/Unix-only" message — use the platform-specific startup commands
on those systems instead.
