build:
ifeq ($(OS),Windows_NT)
	go build -o ./bin/deen.exe ./cmd/deen
else
	go build -o ./bin/deen ./cmd/deen
endif

build-all: build build-freebsd build-macos build-linux build-windows build-wasm

build-freebsd:
	GOOS=freebsd GOARCH=386 go build -o bin/deen-freebsd-x86 ./cmd/deen
	GOOS=freebsd GOARCH=amd64 go build -o bin/deen-freebsd-x86_64 ./cmd/deen

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/deen-darwin-x86_64 ./cmd/deen

build-linux:
	GOOS=linux GOARCH=386 go build -o bin/deen-linux-x86 ./cmd/deen
	GOOS=linux GOARCH=amd64 go build -o bin/deen-linux-x86_64 ./cmd/deen

build-windows:
	GOOS=windows GOARCH=386 go build -o bin/deen-windows-x86.exe ./cmd/deen
	GOOS=windows GOARCH=amd64 go build -o bin/deen-windows-x86_64.exe ./cmd/deen

build-wasm:
	GOOS=js GOARCH=wasm go build -o bin/deen.wasm ./cmd/deen

run:
	go run ./cmd/deen/main.go

clean:
	rm -rf ./bin

test:
	go test -timeout 20s -cover ./...

all: clean build-all test