# startup-remove (sr)

Remove a single Linux/Unix XDG autostart entry that was created by
gitmap. Third-party entries are NEVER touched, even if you pass
their name.

## Synopsis

```
gitmap startup-remove <name>
gitmap sr <name>
```

`<name>` is the entry name as printed by `gitmap startup-list` — the
basename without the `.desktop` extension. A trailing `.desktop` is
tolerated for convenience.

## Outcomes

| Status   | Meaning                                              | Exit |
|----------|------------------------------------------------------|------|
| Removed  | File existed, carried the gitmap marker, was deleted | 0    |
| No-op    | No file by that name in the autostart dir            | 0    |
| Refused  | File exists but lacks the `X-Gitmap-Managed` marker  | 0    |
| Bad name | Name is empty or contains a path separator           | 0    |

All four outcomes exit 0 — the command is idempotent and safe to
script. A real I/O error (permission denied, etc.) exits 1.

## Safety

- The marker is re-checked at remove time (not trusted from a stale
  `startup-list` snapshot), so a race between listing and removing
  cannot trick the command into deleting a third-party file that
  appeared after the listing.
- Names containing `/`, `\`, or NUL bytes are rejected up-front to
  prevent path traversal.

## Platform notes

Linux/Unix only. On Windows or macOS the command exits with a clear
"Linux/Unix-only" message.
