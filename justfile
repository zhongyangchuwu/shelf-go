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
    ./scripts/release-check.sh

release-snapshot:
    ./scripts/release-snapshot.sh

tag version:
    ./scripts/tag-release.sh {{version}}

workflow-check:
    ./scripts/check-workflows.sh
