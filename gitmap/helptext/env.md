# Environment Variables

Manage persistent environment variables and PATH entries across platforms.

## Alias

ev

## Usage

    gitmap env <subcommand> [flags]

## Subcommands

| Subcommand  | Description                              |
|-------------|------------------------------------------|
| set         | Set a persistent environment variable    |
| get         | Get a managed variable's value           |
| delete      | Remove a managed variable                |
| list        | List all managed variables               |
| path add    | Add a directory to PATH                  |
| path remove | Remove a directory from PATH             |
| path list   | List managed PATH entries                |

## Flags

| Flag      | Default | Description                                          |
|-----------|---------|------------------------------------------------------|
| --system  | false   | Target system-level variables (Windows, requires admin) |
| --shell   | (auto)  | Target shell profile: bash, zsh (Unix only)          |
| --verbose | false   | Show detailed operation output                       |
| --dry-run | false   | Preview changes without applying                     |

## Prerequisites

- Windows: setx available (built-in)
- Unix: shell profile (~/.bashrc or ~/.zshrc) writable

## Examples

### Set and retrieve a variable

    $ gitmap env set GOPATH "/home/user/go"
      Set GOPATH=/home/user/go

    $ gitmap env get GOPATH
      GOPATH=/home/user/go

    $ gitmap env list
      Managed variables:
        GOPATH = /home/user/go

### Add a directory to PATH

    $ gitmap ev path add /usr/local/go/bin
      Added to PATH: /usr/local/go/bin

    $ gitmap env path list
      Managed PATH entries:
        /usr/local/go/bin

### Preview changes with dry-run

    $ gitmap env set NODE_ENV "production" --dry-run
      [dry-run] Would set NODE_ENV=production

    $ gitmap env path add /opt/tools/bin --dry-run
      [dry-run] Would add to PATH: /opt/tools/bin

## See Also

- [install](install.md) — Install developer tools
- [doctor](doctor.md) — Diagnose PATH and version issues
- [setup](setup.md) — Configure Git global settings
