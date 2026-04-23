# gitmap revert

Revert the repository to a specific release version by checking out the tag.

## Alias

None

## Usage

    gitmap revert <version>

## Flags

None.

## Prerequisites

- Must be inside a Git repository with release tags
- Run `gitmap list-versions` to see available versions (see list-versions.md)

## Examples

### Example 1: Revert to a specific version

    gitmap revert v2.20.0

**Output:**

    Current version: v2.22.0
    Reverting to v2.20.0...
    Checking out tag v2.20.0... done
    Rebuilding gitmap.exe... done
    Deploying to E:\bin-run\gitmap.exe... done
    ✓ Reverted to v2.20.0
    → Run 'gitmap version' to confirm

### Example 2: Revert to an older version

    gitmap revert v2.15.0

**Output:**

    Current version: v2.22.0
    Reverting to v2.15.0 (7 versions back)...
    Checking out tag v2.15.0... done
    Rebuilding gitmap.exe... done
    Deploying to E:\bin-run\gitmap.exe... done
    ✓ Reverted to v2.15.0

### Example 3: Version tag not found

    gitmap revert v9.9.9

**Output:**

    ✗ Error: tag v9.9.9 not found
    Available versions:
      v2.22.0, v2.21.0, v2.20.0, v2.19.0, ...
    → Use 'gitmap list-versions' to see all available tags

## See Also

- [list-versions](list-versions.md) — List available versions to revert to
- [release](release.md) — Create a new release
- [changelog](changelog.md) — View release notes before reverting
- [update](update.md) — Update to the latest version instead
