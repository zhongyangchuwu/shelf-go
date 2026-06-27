#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage: scripts/release.sh <command> [args]

Commands:
  check              Validate GoReleaser configuration.
  snapshot           Build a local GoReleaser snapshot without publishing.
  tag <version>      Create release tag v<version>. Pass version without leading v.

Examples:
  scripts/release.sh check
  scripts/release.sh snapshot
  scripts/release.sh tag 0.1.1
USAGE
}

validate_version() {
  local version="$1"
  if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    printf 'invalid version %q; expected format like 0.1.1, without leading v\n' "${version}" >&2
    exit 2
  fi
}

release_check() {
  if [[ $# -ne 0 ]]; then
    usage
    exit 2
  fi

  go run github.com/goreleaser/goreleaser/v2@latest check
}

release_snapshot() {
  if [[ $# -ne 0 ]]; then
    usage
    exit 2
  fi

  go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot
}

release_tag() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 2
  fi

  local version="$1"
  validate_version "${version}"

  local tag="v${version}"
  if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
    printf 'tag already exists: %s\n' "${tag}" >&2
    exit 1
  fi

  git tag "${tag}"
  printf 'tagged %s — reinstall (just install) to embed it\n' "${tag}"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 1 ]]; then
  usage
  exit 2
fi

command="$1"
shift

case "${command}" in
  check)
    release_check "$@"
    ;;
  snapshot)
    release_snapshot "$@"
    ;;
  tag)
    release_tag "$@"
    ;;
  *)
    usage
    exit 2
    ;;
esac
