# startup-remove (sr)

Remove a single user-scoped autostart entry that was created by
gitmap. Third-party entries are NEVER touched, even if you pass
their name.

## Synopsis

```
gitmap startup-remove <name>
gitmap sr <name>
```

`<name>` is the entry name as printed by `gitmap startup-list` — the
basename without the platform extension. A trailing platform
extension is tolerated for convenience:

- Linux/Unix: `gitmap-foo` or `gitmap-foo.desktop`
- macOS: `gitmap.foo` or `gitmap.foo.plist`

## Outcomes

| Status   | Meaning                                              | Exit |
|----------|------------------------------------------------------|------|
| Removed  | File existed, carried the gitmap marker, was deleted | 0    |
| No-op    | No file by that name in the autostart dir            | 0    |
| Refused  | File exists but lacks the gitmap marker              | 0    |
| Bad name | Name is empty or contains a path separator           | 0    |

All four outcomes exit 0 — the command is idempotent and safe to
script. A real I/O error (permission denied, etc.) exits 1.

## Safety

- The marker is re-checked at remove time (not trusted from a stale
  `startup-list` snapshot), so a race between listing and removing
  cannot trick the command into deleting a third-party file that
  appeared after the listing.
- The marker grammar is OS-specific:
  - Linux/Unix: `X-Gitmap-Managed=true` line in the `.desktop` body.
  - macOS: `<key>XGitmapManaged</key><true/>` in the `.plist` dict.
- Names containing `/`, `\`, or NUL bytes are rejected up-front to
  prevent path traversal.

## Platform notes

Linux/Unix and macOS are supported. On Windows the command exits
with a clear "unsupported OS" message.

### macOS LaunchAgent caveat

`startup-remove` deletes the `.plist` file but does NOT call
`launchctl unload`. A removed agent takes effect at the next login
or after a manual `launchctl unload <path>` while the user's GUI
session is active. This keeps automated uninstall scripts working in
CI / SSH sessions where `launchctl` is unavailable.
