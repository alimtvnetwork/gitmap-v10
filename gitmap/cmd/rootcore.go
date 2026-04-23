package cmd

import (
	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// dispatchCore routes scan, clone, pull, and status commands.
func dispatchCore(command string) bool {
	return runDispatchTable(command, coreDispatchEntries())
}

// coreDispatchEntries returns the routing table for core commands.
func coreDispatchEntries() []dispatchEntry {
	return []dispatchEntry{
		{[]string{constants.CmdScan, constants.CmdScanAlias}, func() { runScan(argsTail()) }},
		{[]string{constants.CmdClone, constants.CmdCloneAlias}, func() { runClone(argsTail()) }},
		{[]string{constants.CmdPull, constants.CmdPullAlias}, func() { runPull(argsTail()) }},
		{[]string{constants.CmdStatus, constants.CmdStatusAlias}, func() { runStatus(argsTail()) }},
		{[]string{constants.CmdExec, constants.CmdExecAlias}, func() { runExec(argsTail()) }},
		{
			[]string{
				constants.CmdHasAnyUpdates, constants.CmdHasAnyUpdatesAlias,
				constants.CmdHasAnyChanges, constants.CmdHasAnyChangesAlias,
			},
			func() { runHasAnyUpdates(argsTail()) },
		},
		{[]string{constants.CmdHasChange, constants.CmdHasChangeAlias}, func() { runHasChange(argsTail()) }},
		{[]string{constants.CmdCloneNext, constants.CmdCloneNextAlias}, func() { runCloneNext(argsTail()) }},
		{[]string{constants.CmdAs, constants.CmdAsAlias}, func() { runAs(argsTail()) }},
		{[]string{constants.CmdCode}, func() { runCode(argsTail()) }},
	}
}
