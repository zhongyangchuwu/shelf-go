_default:
    @just --list

build:
    go build -o ./bin/shelf ./cmd/shelf

install:
    #!/usr/bin/env bash
    set -euo pipefail
    go install ./cmd/shelf
    mkdir -p "${HOME}/.zfunc"
    go run ./cmd/shelf completion zsh > "${HOME}/.zfunc/_shelf"

    bin_dir="$(go env GOBIN)"
    if [[ -z "${bin_dir}" ]]; then
      bin_dir="$(go env GOPATH)/bin"
    fi

    printf 'Installed shelf to %s and zsh completion to %s\n' "${bin_dir}/shelf" "${HOME}/.zfunc/_shelf"

test:
    go test ./...

tag version:
    git tag v{{version}}
    @echo "tagged v{{version}} — reinstall (just install) to embed it"
