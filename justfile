_default:
    @just --list

build:
    go build -o ./bin/shelf ./cmd/shelf

install:
    go install ./cmd/shelf

test:
    go test ./...

tag version:
    git tag v{{version}}
    @echo "tagged v{{version}} — reinstall (just install) to embed it"
