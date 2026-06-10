_default:
    @just --list

build:
    go build -o ./shelf ./cmd/shelf

install: build
    cp ./shelf ~/.local/bin/shelf

test:
    go test ./...

tag version:
    git tag v{{version}}
    @echo "tagged v{{version}} — rebuild (just install) to embed it"
