package e2e

import (
	"os/exec"
	"testing"
)

// gitRepo wraps a throwaway repository on disk. All helpers fail the
// test on the first git error so the call sites stay short.
type gitRepo struct {
	t   *testing.T
	dir string
}

func newGitRepo(t *testing.T, dir string) *gitRepo {
	t.Helper()

	r := &gitRepo{t: t, dir: dir}

	r.run("init", "-q", "-b", "main")
	r.run("config", "user.email", "ci@example.com")
	r.run("config", "user.name", "ci")
	r.run("config", "commit.gpgsign", "false")

	return r
}

// commit makes a single commit with subject `subject` at author
// timestamp `unix`. We pin both author and committer dates so test
// runs are identical regardless of the wall clock.
func (r *gitRepo) commit(subject string, unix int64) {
	r.t.Helper()

	r.run("commit", "--allow-empty", "-m", subject,
		"--date", formatUnix(unix))
	r.runEnv(map[string]string{
		"GIT_COMMITTER_DATE": formatUnix(unix),
	}, "commit", "--amend", "--no-edit",
		"--date", formatUnix(unix))
}

// tag creates a lightweight tag pointing at HEAD.
func (r *gitRepo) tag(name string) {
	r.t.Helper()

	r.run("tag", name)
}

func (r *gitRepo) run(args ...string) {
	r.t.Helper()

	r.runEnv(nil, args...)
}

func (r *gitRepo) runEnv(env map[string]string, args ...string) {
	r.t.Helper()

	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)

	for k, v := range env {
		cmd.Env = append(cmd.Environ(), k+"="+v)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
