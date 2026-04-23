package constants

// Canonical marker-block PATH snippet templates.
//
// Spec: spec/04-generic-cli/21-post-install-shell-activation/02-snippets.md
//
// SINGLE SOURCE OF TRUTH. run.sh, gitmap/scripts/install.sh, and the
// `gitmap setup print-path-snippet` subcommand all render their rc-file
// snippets from these constants. The shell scripts fetch the rendered
// bytes by shelling out to the gitmap binary, so the three drivers are
// byte-identical by construction.
//
// Format substitution rules:
//   - %[1]s = manager string (e.g. "run.sh", "installer", "gitmap setup")
//   - %[2]s = resolved deploy directory (absolute path, no trailing slash)
//
// Marker lines MUST stay constant across versions for idempotent rewrites.

// PathSnippet marker lines (do not change without bumping snippet version).
const (
	PathSnippetMarkerOpenFmt = "# gitmap shell wrapper v2 - managed by %s. Do not edit manually."
	PathSnippetMarkerClose   = "# gitmap shell wrapper v2 end"
)

// PathSnippet body templates per shell. Each MUST start with the marker
// open line and end with the marker close line so awk/sed-based
// rewriters in run.sh and install.sh can locate the block.
const (
	PathSnippetBashFmt = `# gitmap shell wrapper v2 - managed by %[1]s. Do not edit manually.
export GITMAP_WRAPPER=1
case ":${PATH}:" in *":%[2]s:"*) ;; *) export PATH="$PATH:%[2]s" ;; esac
# gitmap shell wrapper v2 end`

	PathSnippetZshFmt = `# gitmap shell wrapper v2 - managed by %[1]s. Do not edit manually.
export GITMAP_WRAPPER=1
case ":${PATH}:" in *":%[2]s:"*) ;; *) export PATH="$PATH:%[2]s" ;; esac
# gitmap shell wrapper v2 end`

	PathSnippetFishFmt = `# gitmap shell wrapper v2 - managed by %[1]s. Do not edit manually.
set -gx GITMAP_WRAPPER 1
fish_add_path %[2]s
# gitmap shell wrapper v2 end`

	PathSnippetPwshFmt = `# gitmap shell wrapper v2 - managed by %[1]s. Do not edit manually.
$env:GITMAP_WRAPPER = "1"
if (-not ($env:Path -split ';' | Where-Object { $_.TrimEnd('\') -ieq '%[2]s' })) { $env:Path = "$env:Path;%[2]s" }
# gitmap shell wrapper v2 end`
)

// Shell identifiers accepted by `gitmap setup print-path-snippet --shell`.
const (
	PathSnippetShellBash = "bash"
	PathSnippetShellZsh  = "zsh"
	PathSnippetShellFish = "fish"
	PathSnippetShellPwsh = "pwsh"
)

// CLI flag descriptions for the print-path-snippet subcommand.
const (
	FlagDescPathSnippetShell   = "Target shell: bash | zsh | fish | pwsh"
	FlagDescPathSnippetDir     = "Resolved deploy directory to inject into the snippet"
	FlagDescPathSnippetManager = "Manager string shown in the snippet header (e.g. run.sh, installer)"
)

// Errors.
const (
	ErrPathSnippetUnknownShell = "unknown shell %q (expected bash | zsh | fish | pwsh)"
	ErrPathSnippetDirRequired  = "--dir is required"
)
