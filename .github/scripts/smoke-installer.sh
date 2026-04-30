#!/usr/bin/env bash
# Smoke test: verify a freshly-installed gitmap reports the expected version.
#
# Modes:
#   source   Build gitmap from the current checkout into a tempdir, then run
#            `<tempdir>/gitmap version` and assert it contains v$EXPECTED.
#            Used by ci.yml on every PR — no network release dependency.
#
#   release  Run gitmap/scripts/install.sh against a published GitHub release
#            (--version "v$EXPECTED" --no-discovery), then run the installed
#            binary and assert. Used by release.yml after the release is cut.
#
# Reads $EXPECTED (e.g. "4.1.0") from env. Falls back to constants.Version.
#
# Exit 0 on success, non-zero with diagnostic on failure.
set -euo pipefail

MODE="${1:-source}"
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
EXPECTED="${EXPECTED:-$(awk -F'"' '/^const Version/ {print $2}' "$REPO_ROOT/gitmap/constants/constants.go")}"

if [ -z "$EXPECTED" ]; then
  echo "::error::Could not determine expected version" >&2
  exit 2
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "▶ Smoke mode:    $MODE"
echo "▶ Expected:      v$EXPECTED"
echo "▶ Workdir:       $WORK"

case "$MODE" in
  source)
    echo "▶ Building gitmap from source into $WORK"
    (cd "$REPO_ROOT/gitmap" && go build -o "$WORK/gitmap" .)
    BIN="$WORK/gitmap"
    ;;
  release)
    echo "▶ Running install.sh --version v$EXPECTED --no-discovery"
    DEST="$WORK/install"
    mkdir -p "$DEST"
    bash "$REPO_ROOT/gitmap/scripts/install.sh" \
      --version "v$EXPECTED" \
      --dir "$DEST" \
      --no-path \
      --no-discovery
    # install.sh nests the binary under <dest>/gitmap-cli/. Search
    # defensively in case the layout changes (e.g. extra nesting).
    BIN=""
    for candidate in \
      "$DEST/gitmap-cli/gitmap" \
      "$DEST/gitmap" \
      "$DEST/gitmap-cli/bin/gitmap"; do
      if [ -x "$candidate" ]; then
        BIN="$candidate"
        break
      fi
    done
    if [ -z "$BIN" ]; then
      echo "▶ Searching for gitmap binary under $DEST"
      BIN="$(find "$DEST" -type f -name 'gitmap' -perm -u+x 2>/dev/null | head -n1 || true)"
    fi
    if [ -z "$BIN" ]; then
      echo "::error::Could not locate installed gitmap binary under $DEST" >&2
      echo "▶ Tree:" >&2
      find "$DEST" -maxdepth 4 >&2 || true
      exit 3
    fi
    echo "▶ Located binary: $BIN"
    ;;
  *)
    echo "::error::Unknown mode '$MODE' (expected 'source' or 'release')" >&2
    exit 2
    ;;
esac

if [ ! -x "$BIN" ]; then
  echo "::error::Binary not found or not executable at $BIN" >&2
  exit 3
fi

VERSION_OUTPUT="$("$BIN" version 2>&1)"
ACTUAL="$(printf '%s\n' "$VERSION_OUTPUT" | awk '/^gitmap v[0-9]/{print; exit}')"
echo "▶ Actual output: $ACTUAL"

EXPECTED_LINE="gitmap v$EXPECTED"
if [ "$ACTUAL" != "$EXPECTED_LINE" ]; then
  echo "::error::Version mismatch" >&2
  echo "  expected: $EXPECTED_LINE" >&2
  echo "  actual:   $ACTUAL" >&2
  exit 4
fi

echo "✅ Installer smoke test passed: $ACTUAL"
