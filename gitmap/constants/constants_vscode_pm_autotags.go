package constants

// Auto-derived tag detection for VS Code Project Manager sync (v3.40.0+).
//
// Each marker is a top-level file or directory inside a project's rootPath.
// When present, the corresponding tag is added (additively) to the entry's
// `tags` array in projects.json. User-added tags are NEVER removed.
//
// Detection is shallow (top-level only) and order-stable: results are
// emitted following AutoTagOrder so diffs stay clean across runs.

// Canonical tag identifiers — keep in sync with AutoTagMarkers / AutoTagOrder.
const (
	AutoTagGit    = "git"
	AutoTagNode   = "node"
	AutoTagGo     = "go"
	AutoTagPython = "python"
	AutoTagRust   = "rust"
	AutoTagDocker = "docker"
)

// AutoTagMarkers maps a top-level filesystem entry name to the tag it
// implies. Both files and directories qualify (.git can be either).
var AutoTagMarkers = map[string]string{
	".git":               AutoTagGit,
	"package.json":       AutoTagNode,
	"go.mod":             AutoTagGo,
	"pyproject.toml":     AutoTagPython,
	"requirements.txt":   AutoTagPython,
	"Cargo.toml":         AutoTagRust,
	"Dockerfile":         AutoTagDocker,
	"compose.yaml":       AutoTagDocker,
	"compose.yml":        AutoTagDocker,
	"docker-compose.yml": AutoTagDocker,
}

// AutoTagOrder is the canonical emission order. Tags not listed here are
// dropped (the detector never invents tags outside this list).
var AutoTagOrder = []string{
	AutoTagGit,
	AutoTagNode,
	AutoTagGo,
	AutoTagPython,
	AutoTagRust,
	AutoTagDocker,
}

// CLI flag for opting out of auto-tag detection during sync.
const (
	FlagNoAutoTags     = "no-auto-tags"
	FlagDescNoAutoTags = "skip auto-derived tags (git/node/go/...) when syncing VS Code Project Manager projects.json"
)
