_default:
    @just --list

build:
    go build -o ./bin/shelf ./cmd/shelf

install:
    ./scripts/install.sh

test:
    ./scripts/test.sh

tag version:
    ./scripts/release.sh tag {{version}}
