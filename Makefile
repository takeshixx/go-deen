ldflags = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current)"
ldflagsstripped = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current) -w -s"

build:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(ldflags) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(ldflags) -o ./bin/deen ./cmd/deen
endif

stripped:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(ldflagsstripped) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(ldflagsstripped) -o ./bin/deen ./cmd/deen
endif

gui:
ifeq ($(OS),Windows_NT)
	go build $(ldflagsstripped) --tags gui -o ./bin/deen.exe ./cmd/deen
else
	go build $(ldflagsstripped) --tags gui -o ./bin/deen ./cmd/deen
endif

web: 
	rm extras/web/deen.wasm extras/web/wasm_exec.js || true
	GOOS=js GOARCH=wasm go build $(ldflagsstripped) -o extras/web/deen.wasm ./cmd/deen
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js extras/web/wasm_exec.js
	http_server -no-auth -root extras/web -port 9090

run:
	go run ./cmd/deen/main.go

clean:
	rm -rf ./bin
	rm extras/web/deen.wasm extras/web/wasm_exec.js || true

test:
	go test -timeout 20s -cover ./...

all: clean build test