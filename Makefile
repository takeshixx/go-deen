build:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build -ldflags="-X main.version=$$(git describe --abbrev=0 --tags) -X main.branch=$$(git branch --show-current)" -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build -ldflags="-X main.version=$$(git describe --abbrev=0 --tags) -X main.branch=$$(git branch --show-current)" -o ./bin/deen ./cmd/deen
endif

stripped:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build -ldflags="-X main.version=$$(git describe --abbrev=0 --tags) -X main.branch=$$(git branch --show-current) -w -s" -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build -ldflags="-X main.version=$$(git describe --abbrev=0 --tags) -X main.branch=$$(git branch --show-current) -w -s" -o ./bin/deen ./cmd/deen
endif

build-all: build build-freebsd build-macos build-linux build-windows

build-freebsd:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o bin/deen-freebsd-x86 ./cmd/deen
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o bin/deen-freebsd-x86_64 ./cmd/deen

build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/deen-darwin-x86_64 ./cmd/deen

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o bin/deen-linux-x86 ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/deen-linux-x86_64 ./cmd/deen

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o bin/deen-windows-x86.exe ./cmd/deen
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/deen-windows-x86_64.exe ./cmd/deen

web: 
	rm extras/web/deen.wasm extras/web/wasm_exec.js || true
	GOOS=js GOARCH=wasm go build -ldflags="-w -s" -o extras/web/deen.wasm ./cmd/deen
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js extras/web/wasm_exec.js
	http_server -no-auth -root extras/web -port 9090

run:
	go run ./cmd/deen/main.go

clean:
	rm -rf ./bin
	rm extras/web/deen.wasm extras/web/wasm_exec.js

test:
	go test -timeout 20s -cover ./...

all: clean build-all test