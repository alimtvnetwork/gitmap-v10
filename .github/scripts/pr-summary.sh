#!/usr/bin/env bash
# pr-summary.sh — Build a concise PR comment summarizing full-suite-guard.
#
# Inputs:
#   $1  artifact_dir  — directory containing test-output.txt + lint-output.txt
#                       (downloaded full-suite-outputs artifact)
#   $2  out_path      — file to write the rendered Markdown comment to
#
# Required env:
#   FULL_SUITE_RESULT   — "success" | "failure" | "cancelled" | "skipped"
#   GITHUB_REPO         — "owner/repo"
#   GITHUB_RUN_ID       — Actions run ID for log deep links
#   GITHUB_SHA_SHORT    — full or short SHA for display
#
# Sentinel: <!-- gitmap-ci-summary --> is embedded so the sticky-comment
# action can find and replace this comment on subsequent pushes.

set -uo pipefail

readonly ARTIFACT_DIR="${1:?artifact_dir required}"
readonly OUT_PATH="${2:?out_path required}"

readonly TEST_OUT="$ARTIFACT_DIR/test-output.txt"
readonly LINT_OUT="$ARTIFACT_DIR/lint-output.txt"

readonly REPO="${GITHUB_REPO:-unknown/unknown}"
readonly RUN_ID="${GITHUB_RUN_ID:-0}"
readonly SHA="${GITHUB_SHA_SHORT:-HEAD}"
readonly RESULT="${FULL_SUITE_RESULT:-unknown}"

readonly RUN_URL="https://github.com/${REPO}/actions/runs/${RUN_ID}"
readonly ARTIFACTS_URL="${RUN_URL}#artifacts"

# Tally counts from the captured output files. Defensive defaults so a
# missing artifact (e.g. job cancelled before upload) still renders cleanly.
hasTestOutput=true
if [ -f "$TEST_OUT" ]; then
  hasTestOutput=true
else
  hasTestOutput=false
fi

hasLintOutput=true
if [ -f "$LINT_OUT" ]; then
  hasLintOutput=true
else
  hasLintOutput=false
fi

testsPassed=0
testsFailed=0
if [ "$hasTestOutput" = "true" ]; then
  testsPassed=$(grep -cE '^ok[[:space:]]'   "$TEST_OUT" || true)
  testsFailed=$(grep -cE '^FAIL[[:space:]]' "$TEST_OUT" || true)
fi

lintIssues=0
if [ "$hasLintOutput" = "true" ]; then
  lintIssues=$(grep -cE '^[^[:space:]].+:[0-9]+:[0-9]+:' "$LINT_OUT" || true)
fi

# Per-status emoji + headline so the comment is glanceable in the PR feed.
testsStatus="✅ pass"
if [ "$testsFailed" -gt 0 ]; then
  testsStatus="❌ fail"
fi
if [ "$hasTestOutput" = "false" ]; then
  testsStatus="⚠ no output"
fi

lintStatus="✅ clean"
if [ "$lintIssues" -gt 0 ]; then
  lintStatus="❌ $lintIssues issue(s)"
fi
if [ "$hasLintOutput" = "false" ]; then
  lintStatus="⚠ no output"
fi

overallEmoji="✅"
if [ "$RESULT" != "success" ]; then
  overallEmoji="❌"
fi

# First failing tests / first lint issues, capped to keep the comment short.
firstFailures=""
if [ "$testsFailed" -gt 0 ]; then
  firstFailures=$(grep -E '^--- FAIL:|^FAIL[[:space:]]' "$TEST_OUT" | head -n 10 || true)
fi

firstLintLines=""
if [ "$lintIssues" -gt 0 ]; then
  firstLintLines=$(grep -E '^[^[:space:]].+:[0-9]+:[0-9]+:' "$LINT_OUT" | head -n 10 || true)
fi

{
  echo "<!-- gitmap-ci-summary -->"
  echo "## ${overallEmoji} CI Summary — \`${SHA:0:10}\`"
  echo ""
  echo "Full Suite Guard: **${RESULT}**"
  echo ""
  echo "| Stage | Result | Detail |"
  echo "| --- | --- | --- |"
  echo "| \`go test ./...\` | ${testsStatus} | ${testsPassed} package(s) ok, ${testsFailed} failed |"
  echo "| \`golangci-lint\` (strict) | ${lintStatus} | v1.64.8, --max-issues=0 |"
  echo ""

  if [ -n "$firstFailures" ]; then
    echo "### First failing tests"
    echo ""
    echo '```'
    echo "$firstFailures"
    echo '```'
    echo ""
  fi

  if [ -n "$firstLintLines" ]; then
    echo "### First lint findings"
    echo ""
    echo '```'
    echo "$firstLintLines"
    echo '```'
    echo ""
  fi

  echo "### Logs & artifacts"
  echo ""
  echo "- [Full run logs](${RUN_URL})"
  echo "- [Download artifacts](${ARTIFACTS_URL}) — \`full-suite-outputs\` contains the raw \`test-output.txt\` and \`lint-output.txt\`"
  echo ""
  echo "<sub>Reproduce locally: \`./scripts/preflight-ci.sh\`</sub>"
} > "$OUT_PATH"

echo "✓ pr-summary: wrote $OUT_PATH"
