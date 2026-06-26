#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage: scripts/release-snapshot.sh

Builds a local GoReleaser snapshot with --clean. This does not publish a release.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 0 ]]; then
  usage
  exit 2
fi

go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot
