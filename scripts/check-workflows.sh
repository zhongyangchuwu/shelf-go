#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage: scripts/check-workflows.sh

Runs lightweight checks for workflow scripts without creating release tags or artifacts.
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

bash -n scripts/*.sh

scripts/install.sh --help >/dev/null 2>&1
scripts/release-check.sh --help >/dev/null 2>&1
scripts/release-snapshot.sh --help >/dev/null 2>&1
scripts/tag-release.sh --help >/dev/null 2>&1

if scripts/tag-release.sh >/dev/null 2>&1; then
  printf 'expected scripts/tag-release.sh without arguments to fail\n' >&2
  exit 1
fi

if scripts/tag-release.sh v0.1.1 >/dev/null 2>&1; then
  printf 'expected scripts/tag-release.sh with leading v to fail\n' >&2
  exit 1
fi

if scripts/release-check.sh unexpected >/dev/null 2>&1; then
  printf 'expected scripts/release-check.sh with extra arguments to fail\n' >&2
  exit 1
fi

printf 'workflow script checks passed\n'
