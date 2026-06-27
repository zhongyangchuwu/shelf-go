_default:
    @just --list

build:
    go build -o ./bin/shelf ./cmd/shelf

install:
    ./scripts/install.sh

test:
    go test ./...

vet:
    go vet ./...

release-check:
    ./scripts/release.sh check

release-snapshot:
    ./scripts/release.sh snapshot

tag version:
    ./scripts/release.sh tag {{version}}
