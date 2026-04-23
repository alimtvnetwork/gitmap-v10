// Package cloner — branch selection strategy.
//
// pickCloneStrategy decides whether `git clone` should be invoked with an
// explicit `-b <branch>` flag or without it (letting the remote's default
// HEAD decide). The decision is driven by ScanRecord.BranchSource so that
// untrustworthy values (literal "HEAD", a detached SHA, or "unknown") never
// reach the git command line and produce "Remote branch not found" errors.
package cloner

import (
	"github.com/alimtvnetwork/gitmap-v6/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// cloneStrategy describes how a clone should be invoked.
type cloneStrategy struct {
	// useBranch is true when `-b <branch>` should be passed to git clone.
	useBranch bool
	// branch is the branch name to check out (only used when useBranch).
	branch string
	// reason is a short human-readable description of why this strategy
	// was chosen. It is propagated into CloneResult.Notes for diagnostics.
	reason string
}

// pickCloneStrategy decides whether to checkout the recorded branch, the
// remote-tracking branch, the repo default, or fall back to the remote's
// default HEAD. Decisions are made from BranchSource so that scans that
// captured a detached / unknown state never produce "branch not found"
// errors during clone.
func pickCloneStrategy(rec model.ScanRecord) cloneStrategy {
	branch := rec.Branch

	switch rec.BranchSource {
	case gitutil.BranchSourceHEAD:
		if branch == "" || branch == gitutil.BranchSourceHEAD {
			return cloneStrategy{
				reason: "branchSource=HEAD but branch empty; using remote default",
			}
		}

		return cloneStrategy{
			useBranch: true,
			branch:    branch,
			reason:    "branchSource=HEAD; checking out detected branch",
		}

	case gitutil.BranchSourceRemoteTracking:
		if branch == "" {
			return cloneStrategy{
				reason: "branchSource=remote-tracking but branch empty; using remote default",
			}
		}

		return cloneStrategy{
			useBranch: true,
			branch:    branch,
			reason:    "branchSource=remote-tracking; checking out tracked branch",
		}

	case gitutil.BranchSourceDefault:
		if branch == "" {
			return cloneStrategy{
				reason: "branchSource=default but branch empty; using remote default",
			}
		}

		return cloneStrategy{
			useBranch: true,
			branch:    branch,
			reason:    "branchSource=default; checking out repo default branch",
		}

	case gitutil.BranchSourceDetached:
		return cloneStrategy{
			reason: "branchSource=detached; using remote default HEAD",
		}

	case gitutil.BranchSourceUnknown, "":
		return cloneStrategy{
			reason: "branchSource=unknown; using remote default HEAD",
		}
	}

	// Unrecognized source label — be safe and skip -b.
	return cloneStrategy{
		reason: "branchSource=" + rec.BranchSource + " (unrecognized); using remote default HEAD",
	}
}
