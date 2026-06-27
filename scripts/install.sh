#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage: scripts/install.sh

Installs shelf with `go install ./cmd/shelf` and writes zsh completion.

Environment overrides:
  GOBIN                  Go install destination directory
  SHELF_COMPLETION_DIR   Completion directory, default: $HOME/.zfunc
  SHELF_COMPLETION_FILE  Completion file path, default: $SHELF_COMPLETION_DIR/_shelf
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

go install ./cmd/shelf

completion_dir="${SHELF_COMPLETION_DIR:-${HOME}/.zfunc}"
completion_file="${SHELF_COMPLETION_FILE:-${completion_dir}/_shelf}"
mkdir -p "$(dirname "${completion_file}")"
go run ./cmd/shelf completion zsh > "${completion_file}"

bin_dir="$(go env GOBIN)"
if [[ -z "${bin_dir}" ]]; then
  bin_dir="$(go env GOPATH)/bin"
fi

printf 'Installed shelf to %s and zsh completion to %s\n' "${bin_dir}/shelf" "${completion_file}"
