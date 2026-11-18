.PHONY: build test lint run codestyle

build:
	go build -ldflags "-X github.com/jamesread/greyvar/dat-editor/service/buildinfo.Version=$$(cat ../VERSION 2>/dev/null || echo development)" -o dat-editor-service .

test:
	go test ./...

lint: codestyle

codestyle:
	go fmt ./...
	go vet ./...

run:
	go run .
