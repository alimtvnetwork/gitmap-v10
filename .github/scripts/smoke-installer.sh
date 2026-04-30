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

load_deploy_manifest() {
  local manifest_path="$REPO_ROOT/gitmap/constants/deploy-manifest.json"
  APP_SUBDIR="gitmap-cli"
  BINARY_NAME_UNIX="gitmap"
  LEGACY_APP_SUBDIRS=("gitmap")

  if [ ! -f "$manifest_path" ]; then
    return
  fi

  APP_SUBDIR="$(awk -F'"' '/"appSubdir"/ {print $4; exit}' "$manifest_path")"
  BINARY_NAME_UNIX="$(awk -F'"' '/"unix"/ {print $4; exit}' "$manifest_path")"

  mapfile -t LEGACY_APP_SUBDIRS < <(
    awk '
      /"legacyAppSubdirs"/ { in_block=1 }
      in_block {
        while (match($0, /"[^"]+"/)) {
          value = substr($0, RSTART + 1, RLENGTH - 2)
          if (value != "legacyAppSubdirs") {
            print value
          }
          $0 = substr($0, RSTART + RLENGTH)
        }
      }
      in_block && /\]/ { exit }
    ' "$manifest_path"
  )

  if [ -z "$APP_SUBDIR" ]; then
    APP_SUBDIR="gitmap-cli"
  fi
  if [ -z "$BINARY_NAME_UNIX" ]; then
    BINARY_NAME_UNIX="gitmap"
  fi
  if [ "${#LEGACY_APP_SUBDIRS[@]}" -eq 0 ]; then
    LEGACY_APP_SUBDIRS=("gitmap")
  fi
}

if [ -z "$EXPECTED" ]; then
  echo "::error::Could not determine expected version" >&2
  exit 2
fi

load_deploy_manifest

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
    # Resolve from the deploy manifest so the smoke test tracks the
    # installer's canonical layout instead of stale hardcoded paths.
    BIN=""
    for candidate in \
      "$DEST/$APP_SUBDIR/$BINARY_NAME_UNIX" \
      "$DEST/$BINARY_NAME_UNIX"; do
      if [ -f "$candidate" ] && [ -x "$candidate" ]; then
        BIN="$candidate"
        break
      fi
    done
    if [ -z "$BIN" ]; then
      for legacy in "${LEGACY_APP_SUBDIRS[@]}"; do
        candidate="$DEST/$legacy/$BINARY_NAME_UNIX"
        if [ -f "$candidate" ] && [ -x "$candidate" ]; then
          BIN="$candidate"
          break
        fi
      done
    fi
    if [ -z "$BIN" ]; then
      echo "▶ Searching for $BINARY_NAME_UNIX under $DEST"
      BIN="$(find "$DEST" -type f -name "$BINARY_NAME_UNIX" -perm -u+x 2>/dev/null | head -n1 || true)"
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

if [ ! -f "$BIN" ] || [ ! -x "$BIN" ]; then
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
