package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/render"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/templates"
)

const (
	cmdTemplatesList      = "list"
	cmdTemplatesListAlias = "tl"
	cmdTemplatesShow      = "show"
	cmdTemplatesShowAlias = "ts"
	cmdTemplatesInit      = "init"
	cmdTemplatesInitAlias = "ti"
	usageTemplatesRoot    = `Usage: gitmap templates <subcommand>

Subcommands:
  list                       List every available template (alias: tl)
  show <kind> <lang>         Print a single template to stdout (alias: ts)
  init <lang> [<lang>...]    Scaffold .gitignore / .gitattributes for languages (alias: ti)

Kinds:
  ignore | attributes | lfs

Flags (show):
  --raw                      Disable pretty markdown rendering even on a TTY

Flags (init):
  --lfs                      Also merge lfs/common.gitattributes into .gitattributes
  --dry-run                  Preview every block; do not touch disk
  --force                    Replace existing .gitignore/.gitattributes outright

Examples:
  gitmap templates list
  gitmap tpl tl
  gitmap templates show ignore go
  gitmap tpl ts attributes node
  gitmap templates show ignore go --raw   # bypass pretty renderer
  gitmap templates init go
  gitmap templates init go node --lfs
  gitmap tpl ti python --dry-run
`
	headerTemplatesList    = "KIND        LANG            SOURCE  PATH\n"
	fmtTemplatesListRow    = "%-10s  %-14s  %-6s  %s\n"
	labelTemplatesUser     = "user"
	labelTemplatesEmbed    = "embed"
	msgTemplatesEmpty      = "(no templates registered — embedded corpus is empty)\n"
	errTemplatesShowArgs   = "templates show requires <kind> <lang>; e.g. 'templates show ignore go'\n"
	errTemplatesShowFail   = "templates show: %v\n"
	errTemplatesListFail   = "templates list: %v\n"
	errUnknownTemplatesSub = "unknown 'templates' subcommand: %s\n"
	flagTemplatesShowRaw   = "raw"
	flagDescTemplatesRaw   = "Deprecated alias for --no-pretty (kept for v3.23.x back-compat)"
)

// dispatchTemplates routes `gitmap templates <subcommand>` calls.
func dispatchTemplates(command string) bool {
	if command != constants.CmdTemplates && command != constants.CmdTemplatesAlias {
		return false
	}
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stderr, usageTemplatesRoot)
		os.Exit(1)
	}

	sub, rest := os.Args[2], os.Args[3:]
	switch sub {
	case cmdTemplatesList, cmdTemplatesListAlias:
		runTemplatesList()
	case cmdTemplatesShow, cmdTemplatesShowAlias:
		runTemplatesShow(rest)
	case cmdTemplatesInit, cmdTemplatesInitAlias:
		runTemplatesInit(rest)
	default:
		fmt.Fprintf(os.Stderr, errUnknownTemplatesSub, sub)
		fmt.Fprint(os.Stderr, usageTemplatesRoot)
		os.Exit(1)
	}

	return true
}

// runTemplatesList prints every available template grouped by kind.
func runTemplatesList() {
	entries, err := templates.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesListFail, err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Print(msgTemplatesEmpty)

		return
	}

	fmt.Print(headerTemplatesList)
	for _, e := range entries {
		fmt.Printf(fmtTemplatesListRow, e.Kind, e.Lang, sourceLabel(e.Source), e.Path)
	}
}

// runTemplatesShow prints one template to stdout. Markdown templates
// (.md / .markdown) are routed through render.RenderANSI when the
// shared render.Decide() ladder says so for the parsed PrettyMode.
// Non-markdown templates (.gitignore, .gitattributes, …) are always
// written byte-for-byte regardless of mode — render.Decide enforces
// that via its isMarkdown gate, so the dominant
// `templates show ignore go > .gitignore` redirect workflow is safe.
//
// Flag precedence: --pretty / --no-pretty (preferred) win over the
// legacy --raw flag, which is kept as a deprecated alias for
// --no-pretty (back-compat with v3.23.x scripts).
func runTemplatesShow(args []string) {
	rest, mode := parseTemplatesShowFlags(args)
	if len(rest) < 2 {
		fmt.Fprint(os.Stderr, errTemplatesShowArgs)
		os.Exit(1)
	}
	kind, lang := rest[0], rest[1]
	r, err := templates.Resolve(kind, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}

	out := r.Content
	if render.Decide(mode, render.StdoutIsTerminal(), isMarkdownTemplatePath(r.Path)) {
		out = []byte(render.RenderANSI(string(r.Content)))
	}

	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}
}

// parseTemplatesShowFlags extracts --pretty / --no-pretty (preferred) and
// the legacy --raw alias from args, returning the cleaned positional
// slice + the resolved render.PrettyMode. --raw is treated as
// --no-pretty for back-compat with v3.23.x. When both are present,
// --pretty wins (--raw only downgrades when mode is still Auto).
func parseTemplatesShowFlags(args []string) ([]string, render.PrettyMode) {
	cleaned, mode := ParsePrettyFlag(args)

	fs := flag.NewFlagSet(cmdTemplatesShow, flag.ExitOnError)
	rawFlag := fs.Bool(flagTemplatesShowRaw, false, flagDescTemplatesRaw)
	reordered := reorderFlagsBeforeArgs(cleaned)
	_ = fs.Parse(reordered)

	if *rawFlag && mode == render.PrettyAuto {
		mode = render.PrettyOff
	}

	return fs.Args(), mode
}

// isMarkdownTemplatePath returns true for .md / .markdown files
// (case-insensitive). Templates today are .gitignore / .gitattributes —
// this guard future-proofs the renderer for markdown overlays
// (e.g. ~/.gitmap/templates/notes/*.md) without changing existing UX.
func isMarkdownTemplatePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	return ext == ".md" || ext == ".markdown"
}

// sourceLabel maps a templates.Source to the user-facing column value.
func sourceLabel(s templates.Source) string {
	if s == templates.SourceUser {
		return labelTemplatesUser
	}

	return labelTemplatesEmbed
}
