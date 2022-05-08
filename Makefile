ldflags = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current)"
ldflagsstripped = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current) -w -s"

.PHONY: build
build:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(ldflags) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(ldflags) -o ./bin/deen ./cmd/deen
endif

.PHONY: cross
cross:
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(ldflags) -o ./bin/linux-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(ldflags) -o ./bin/linux-386/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build $(ldflags) -o ./bin/linux-arm/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(ldflags) -o ./bin/linux-arm64/deen ./cmd/deen
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(ldflags) -o ./bin/windows-amd64/deen.exe ./cmd/deen
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(ldflags) -o ./bin/windows-386/deen.exe ./cmd/deen
	# Darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(ldflags) -o ./bin/darwin-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(ldflags) -o ./bin/darwin-arm64/deen ./cmd/deen

.PHONY: stripped
stripped:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(ldflagsstripped) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(ldflagsstripped) -o ./bin/deen ./cmd/deen
endif

.PHONY: cross-stripped
cross-stripped:
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(ldflagsstripped) -o ./bin/linux-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(ldflagsstripped) -o ./bin/linux-386/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build $(ldflagsstripped) -o ./bin/linux-arm/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(ldflagsstripped) -o ./bin/linux-arm64/deen ./cmd/deen
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(ldflagsstripped) -o ./bin/windows-amd64/deen.exe ./cmd/deen
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(ldflagsstripped) -o ./bin/windows-386/deen.exe ./cmd/deen
	# Darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(ldflagsstripped) -o ./bin/darwin-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(ldflagsstripped) -o ./bin/darwin-arm64/deen ./cmd/deen

.PHONY: gui
gui:
ifeq ($(OS),Windows_NT)
	go build $(ldflagsstripped) --tags gui -o ./bin/deen.exe ./cmd/deen
else
	go build $(ldflagsstripped) --tags gui -o ./bin/deen ./cmd/deen
endif

.PHONY: fyne-cross
fyne-cross:
	fyne-cross linux $(ldflags) --tags gui --arch=* -output deen ./cmd/deen
	fyne-cross windows $(ldflags) --tags gui --arch=* -output deen.exe ./cmd/deen
	#fyne-cross darwin $(ldflags) --tags gui --arch=* -output deen-darwin ./cmd/deen

.PHONY: web
web: 
	rm extras/web/deen.wasm extras/web/wasm_exec.js || true
	GOOS=js GOARCH=wasm go build $(ldflagsstripped) -o extras/web/deen.wasm ./cmd/deen
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js extras/web/wasm_exec.js
	http_server -no-auth -root extras/web -port 9090

.PHONY: run
run:
	go run ./cmd/deen/main.go

.PHONY: clean
clean:
	rm -rf ./bin
	rm -f extras/web/deen.wasm extras/web/wasm_exec.js

.PHONY: test
test:
	go test -timeout 20s -count=1 -cover ./...

.PHONY: all
all: clean build test