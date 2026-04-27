package goldenguard

// Tests for the AllowUpdate dual gate. The function is tiny but its
// failure mode (silently allowing a CI fixture rewrite) is severe, so
// every branch is pinned: trigger-off, trigger-on+allow-on, and
// trigger-on+allow-bad cases that MUST fail loudly.
//
// Failure-path tests use t.Run + subT.Failed(): when the inner sub-
// test calls Fatalf, the parent observes a failure flag without the
// outer test itself failing — the standard idiom for testing code
// that calls *testing.T.Fatalf.

import (
	"os"
	"testing"
)

// TestAllowUpdate_TriggerOff_IsFalse: when the per-test trigger is
// off the function must short-circuit to false WITHOUT consulting
// the env var. This is the hot path in CI — it must never touch
// os.Getenv-driven branches that could call t.Fatalf.
func TestAllowUpdate_TriggerOff_IsFalse(t *testing.T) {
	t.Setenv(AllowUpdateEnv, "1") // even with allow ON, trigger OFF wins
	if AllowUpdate(t, false) {
		t.Fatalf("AllowUpdate(false, allow=1) = true, want false "+
			"(trigger-off must short-circuit before reading %s)",
			AllowUpdateEnv)
	}
}

// TestAllowUpdate_BothOn_IsTrue: the only path that returns true.
// Documents the exact value pairing — trigger=true AND env=="1".
func TestAllowUpdate_BothOn_IsTrue(t *testing.T) {
	t.Setenv(AllowUpdateEnv, "1")
	if !AllowUpdate(t, true) {
		t.Fatalf("AllowUpdate(true, allow=1) = false, want true")
	}
}

// TestAllowUpdate_TriggerOnAllowMissing_Fails: the failure path that
// catches a stray -update flag or GITMAP_UPDATE_GOLDEN=1 in CI when
// the dedicated allow var was (correctly) NOT set.
func TestAllowUpdate_TriggerOnAllowMissing_Fails(t *testing.T) {
	// Clear inherited value so the sub-test sees a truly-empty env.
	_ = os.Unsetenv(AllowUpdateEnv)
	if !expectFatal(t, true) {
		t.Fatalf("AllowUpdate(true, allow=<unset>) did NOT fail — "+
			"missing %s must abort the regenerate path", AllowUpdateEnv)
	}
}

// TestAllowUpdate_TriggerOnAllowWrongValue_Fails: typo guard. The
// allow var is intentionally narrow (literal "1" only) so common
// misspellings ("true", "yes") fail closed instead of unlocking.
func TestAllowUpdate_TriggerOnAllowWrongValue_Fails(t *testing.T) {
	for _, bad := range []string{"true", "yes", "y", "TRUE", "0", " 1 "} {
		bad := bad
		t.Run(bad, func(tt *testing.T) {
			tt.Setenv(AllowUpdateEnv, bad)
			if !expectFatal(tt, true) {
				tt.Fatalf("AllowUpdate accepted bogus allow=%q "+
					"(only literal \"1\" must unlock the gate)",
					bad)
			}
		})
	}
}

// expectFatal runs AllowUpdate(_, trigger) inside a sub-test and
// reports whether that sub-test failed. Idiomatic Go pattern for
// asserting that a function calls t.Fatalf without aborting the
// outer test. The sub-test name is the call-site's t.Name() to keep
// output readable when the parent loops.
func expectFatal(parent *testing.T, trigger bool) bool {
	parent.Helper()
	var failed bool
	parent.Run("expect-fatal", func(child *testing.T) {
		_ = AllowUpdate(child, trigger)
		failed = child.Failed()
	})

	return failed
}
