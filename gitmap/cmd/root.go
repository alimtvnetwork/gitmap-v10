// Package cmd implements the CLI commands for gitmap.
package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v7/gitmap/constants"
)

// Run is the main entry point for the CLI.
func Run() {
	initConsole()

	if len(os.Args) < 2 {
		PrintBinaryLocations()
		printUsage()
		os.Exit(1)
	}

	// Skip migration for commands that must produce clean stdout
	cmd := os.Args[1]
	if cmd != constants.CmdVersion && cmd != constants.CmdVersionAlias {
		migrateLegacyDirs()
	}

	aliasName, cleaned := extractAliasFlag(os.Args[2:])
	if len(aliasName) > 0 {
		resolveAliasContext(aliasName)
		os.Args = append(os.Args[:2], cleaned...)
	}

	command := os.Args[1]
	dispatch(command)
}

// dispatch routes to the correct subcommand handler with audit tracking.
func dispatch(command string) {
	auditID, auditStart := recordAuditStart(command, os.Args[2:])

	if dispatchCore(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchRelease(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchUtility(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchData(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchTooling(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchProjectRepos(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchDiff(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchMoveMerge(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchAdd(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchTemplates(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}
	if dispatchCommitTransfer(command) {
		recordAuditEnd(auditID, auditStart, 0, "", 0)

		return
	}

	fmt.Fprintf(os.Stderr, constants.ErrUnknownCommand, command)
	printUsage()
	os.Exit(1)
}
