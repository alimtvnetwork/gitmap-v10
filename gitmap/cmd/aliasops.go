package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v9/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v9/gitmap/store"
)

// runAliasSet handles "alias set <alias> <slug>".
func runAliasSet(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, constants.ErrAliasEmpty)
		os.Exit(1)
	}

	alias := args[0]
	slug := args[1]

	executeAliasSet(alias, slug)
}

// executeAliasSet resolves the slug and creates or updates the alias.
func executeAliasSet(alias, slug string) {
	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrListDBFailed, err)
		os.Exit(1)
	}
	defer db.Close()

	repos, err := db.FindBySlug(slug)
	if err != nil || len(repos) == 0 {
		fmt.Fprintf(os.Stderr, constants.ErrAliasRepoMissing, slug)
		os.Exit(1)
	}

	repoID := repos[0].ID

	if db.AliasExists(alias) {
		err = db.UpdateAlias(alias, repoID)
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
			os.Exit(1)
		}

		fmt.Printf(constants.MsgAliasUpdated, alias, slug)
		printHints(aliasSetHints())

		return
	}

	_, err = db.CreateAlias(alias, repoID)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgAliasCreated, alias, slug)
	printHints(aliasSetHints())
}

// runAliasRemove handles "alias remove <alias>".
func runAliasRemove(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrAliasEmpty)
		os.Exit(1)
	}

	alias := args[0]

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrListDBFailed, err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.DeleteAlias(alias)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgAliasRemoved, alias)
}

// runAliasList handles "alias list".
func runAliasList() {
	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrListDBFailed, err)
		os.Exit(1)
	}
	defer db.Close()

	aliases, err := db.ListAliasesWithRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
		os.Exit(1)
	}

	printAliasList(aliases)
	printHints(aliasListHints())
}

// printAliasList renders the alias table to stdout.
func printAliasList(aliases []store.AliasWithRepo) {
	if len(aliases) == 0 {
		fmt.Println("  No aliases defined.")

		return
	}

	fmt.Printf(constants.MsgAliasListHeader, len(aliases))

	for _, a := range aliases {
		fmt.Printf(constants.MsgAliasListRow, a.Alias.Alias, a.Slug)
	}
}

// runAliasShow handles "alias show <alias>".
func runAliasShow(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrAliasEmpty)
		os.Exit(1)
	}

	alias := args[0]

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrListDBFailed, err)
		os.Exit(1)
	}
	defer db.Close()

	resolved, err := db.ResolveAlias(alias)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgAliasResolved, resolved.Alias, resolved.AbsolutePath, resolved.Slug)
}

// isLegacyDataError checks if an error indicates legacy UUID-format data.
func isLegacyDataError(err error) bool {
	return strings.Contains(err.Error(), "Scan error") ||
		strings.Contains(err.Error(), "converting driver.Value type string")
}
