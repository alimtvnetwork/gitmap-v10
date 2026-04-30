package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v10/gitmap/cliexit"
	"github.com/alimtvnetwork/gitmap-v10/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v10/gitmap/movemerge"
)

// runMove implements `gitmap mv LEFT RIGHT`.
//
// Spec: spec/01-app/97-move-and-merge.md
func runMove(args []string) {
	checkHelp(constants.CmdMv, args)
	left, right, opts := parseMoveArgs(args)
	leftEP := mustResolve(left, true, opts)
	rightEP := mustResolve(right, false, opts)
	logResolved(leftEP, rightEP, opts)
	if err := movemerge.RunMove(leftEP, rightEP, opts); err != nil {
		cliexit.Fail(constants.CmdMv, "move", leftEP.DisplayName+" -> "+rightEP.DisplayName, err, 1)
	}
}

// parseMoveArgs parses positional + flag arguments for mv.
func parseMoveArgs(args []string) (string, string, movemerge.Options) {
	fs := flag.NewFlagSet(constants.CmdMv, flag.ExitOnError)
	mf := &movemergeFlagSet{}
	mf.bindFlags(fs)
	if err := fs.Parse(reorderFlagsBeforeArgs(args)); err != nil {
		os.Exit(2)
	}
	rest := fs.Args()
	if len(rest) != 2 {
		fmt.Fprintf(os.Stderr, constants.ErrMMUsageFmt, constants.CmdMv)
		os.Exit(2)
	}
	opts := mf.toOptions(constants.CmdMv, constants.LogPrefixMv, constants.CommitMsgMv)

	return rest[0], rest[1], opts
}

// mustResolve resolves an endpoint or exits with code 1 on failure.
func mustResolve(raw string, isLeft bool, opts movemerge.Options) movemerge.Endpoint {
	ep, err := movemerge.ResolveEndpoint(raw, isLeft, opts)
	if err != nil {
		cliexit.Fail(constants.CmdMv, "resolve-endpoint", raw, err, 1)
	}

	return ep
}

// logResolved emits the [cmd] resolving LEFT/RIGHT lines.
func logResolved(l, r movemerge.Endpoint, opts movemerge.Options) {
	fmt.Printf("%s resolving LEFT  : %s\n", opts.LogPrefix, l.DisplayName)
	fmt.Printf("%s resolving RIGHT : %s\n", opts.LogPrefix, r.DisplayName)
}
