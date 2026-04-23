package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/store"
)

// runUninstall handles the "uninstall" command.
func runUninstall(args []string) {
	checkHelp("uninstall", args)

	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)

	var dryRun, force, purge bool

	fs.BoolVar(&dryRun, constants.FlagUninstallDryRun, false, constants.FlagDescUninstallDryRun)
	fs.BoolVar(&force, constants.FlagUninstallForce, false, constants.FlagDescUninstallForce)
	fs.BoolVar(&purge, constants.FlagUninstallPurge, false, constants.FlagDescUninstallPurge)
	fs.Parse(args)

	tool := fs.Arg(0)
	if tool == "" {
		fmt.Fprint(os.Stderr, constants.ErrInstallToolRequired)
		os.Exit(1)
	}

	validateToolName(tool)

	db, err := openDB()
	if err != nil {
		if !force {
			fmt.Fprintf(os.Stderr, constants.ErrUninstallNotFound, tool)
			os.Exit(1)
		}
	}

	if db != nil {
		defer db.Close()

		if !db.IsToolInstalled(tool) && !force {
			fmt.Fprintf(os.Stderr, constants.ErrUninstallNotFound, tool)
			os.Exit(1)
		}
	}

	if !force && !confirmUninstall(tool) {
		return
	}

	manager := resolveUninstallManager(db, tool)
	uninstallCmd := buildUninstallCommand(manager, tool, purge)

	if dryRun {
		fmt.Printf(constants.MsgUninstallDryCmd, strings.Join(uninstallCmd, " "))

		return
	}

	fmt.Printf(constants.MsgUninstallRemoving, tool)
	runInstallCommand(uninstallCmd, installOptions{Tool: tool, Verbose: true})

	if db != nil {
		if err := db.RemoveInstalledTool(tool); err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrUninstallDBRemove, tool, err)
		}
	}

	fmt.Printf(constants.MsgUninstallSuccess, tool)
}

// confirmUninstall prompts the user for confirmation.
func confirmUninstall(tool string) bool {
	fmt.Printf(constants.MsgUninstallConfirm, tool)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes"
}

// resolveUninstallManager determines which manager was used to install.
func resolveUninstallManager(db *store.DB, tool string) string {
	if db == nil {
		return resolvePackageManager("")
	}

	record, err := db.GetInstalledTool(tool)
	if err != nil || record.PackageManager == "" {
		return resolvePackageManager("")
	}

	return record.PackageManager
}

// buildUninstallCommand builds the uninstall command for a manager.
func buildUninstallCommand(manager, tool string, purge bool) []string {
	pkgName := resolvePackageName(manager, tool)

	if manager == constants.PkgMgrChocolatey {
		return buildChocoUninstall(pkgName, purge)
	}
	if manager == constants.PkgMgrWinget {
		return []string{"winget", "uninstall", pkgName}
	}
	if manager == constants.PkgMgrApt {
		return buildAptUninstall(pkgName, purge)
	}
	if manager == constants.PkgMgrBrew {
		return []string{"brew", "uninstall", pkgName}
	}
	if manager == constants.PkgMgrSnap {
		return []string{"sudo", "snap", "remove", pkgName}
	}

	return buildChocoUninstall(pkgName, purge)
}

// buildChocoUninstall builds a Chocolatey uninstall command.
func buildChocoUninstall(pkg string, purge bool) []string {
	args := []string{"choco", "uninstall", pkg, "-y"}

	if purge {
		args = append(args, "-x")
	}

	return args
}

// buildAptUninstall builds an apt uninstall command.
func buildAptUninstall(pkg string, purge bool) []string {
	action := "remove"

	if purge {
		action = "purge"
	}

	return []string{"sudo", "apt", action, "-y", pkg}
}
