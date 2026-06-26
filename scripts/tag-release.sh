#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage: scripts/tag-release.sh <version>

Creates release tag v<version>. Pass the version without the leading v.
Example: scripts/tag-release.sh 0.1.1
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  usage
  exit 2
fi

version="$1"
if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  printf 'invalid version %q; expected format like 0.1.1, without leading v\n' "${version}" >&2
  exit 2
fi

tag="v${version}"
if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  printf 'tag already exists: %s\n' "${tag}" >&2
  exit 1
fi

git tag "${tag}"
printf 'tagged %s — reinstall (just install) to embed it\n' "${tag}"
