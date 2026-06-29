ldflags = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current)"
ldflagsstripped = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current) -w -s"
gomodflags = -mod=readonly

# GUI builds use cgo; on macOS the Xcode 15+ linker warns about Fyne passing
# -lobjc twice. Silence it there only (GNU ld would reject this flag).
guildflags = $(ldflagsstripped)
ifeq ($(shell uname -s),Darwin)
guildflags = -ldflags "-X github.com/takeshixx/deen/internal/core.version=$$(git describe --abbrev=0 --tags --always) -X github.com/takeshixx/deen/internal/core.branch=$$(git branch --show-current) -w -s -extldflags=-Wl,-no_warn_duplicate_libraries"
endif

.PHONY: build
build:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(gomodflags) $(ldflags) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(gomodflags) $(ldflags) -o ./bin/deen ./cmd/deen
endif

.PHONY: cross
cross:
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(gomodflags) $(ldflags) -o ./bin/linux-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(gomodflags) $(ldflags) -o ./bin/linux-386/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build $(gomodflags) $(ldflags) -o ./bin/linux-arm/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(gomodflags) $(ldflags) -o ./bin/linux-arm64/deen ./cmd/deen
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(gomodflags) $(ldflags) -o ./bin/windows-amd64/deen.exe ./cmd/deen
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(gomodflags) $(ldflags) -o ./bin/windows-386/deen.exe ./cmd/deen
	# Darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(gomodflags) $(ldflags) -o ./bin/darwin-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(gomodflags) $(ldflags) -o ./bin/darwin-arm64/deen ./cmd/deen

.PHONY: stripped
stripped:
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=0 go build $(gomodflags) $(ldflagsstripped) -o ./bin/deen.exe ./cmd/deen
else
	CGO_ENABLED=0 go build $(gomodflags) $(ldflagsstripped) -o ./bin/deen ./cmd/deen
endif

.PHONY: cross-stripped
cross-stripped:
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(gomodflags) $(ldflagsstripped) -o ./bin/linux-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(gomodflags) $(ldflagsstripped) -o ./bin/linux-386/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build $(gomodflags) $(ldflagsstripped) -o ./bin/linux-arm/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(gomodflags) $(ldflagsstripped) -o ./bin/linux-arm64/deen ./cmd/deen
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(gomodflags) $(ldflagsstripped) -o ./bin/windows-amd64/deen.exe ./cmd/deen
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(gomodflags) $(ldflagsstripped) -o ./bin/windows-386/deen.exe ./cmd/deen
	# Darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(gomodflags) $(ldflagsstripped) -o ./bin/darwin-amd64/deen ./cmd/deen
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(gomodflags) $(ldflagsstripped) -o ./bin/darwin-arm64/deen ./cmd/deen

.PHONY: gui
gui:
ifeq ($(OS),Windows_NT)
	go build $(gomodflags) $(ldflagsstripped) --tags gui -o ./bin/deen.exe ./cmd/deen
else
	go build $(gomodflags) $(guildflags) --tags gui -o ./bin/deen ./cmd/deen
endif

.PHONY: full
full: web-assets
ifeq ($(OS),Windows_NT)
	go build $(gomodflags) $(ldflagsstripped) --tags "gui webembed" -o ./bin/deen.exe ./cmd/deen
else
	go build $(gomodflags) $(guildflags) --tags "gui webembed" -o ./bin/deen ./cmd/deen
endif

.PHONY: fyne-cross
fyne-cross:
	fyne-cross linux $(ldflags) --tags gui --arch=* -output ./bin/linux-fyne/deen ./cmd/deen
	fyne-cross windows $(ldflags) --tags gui --arch=* -output ./bin/windows-fyne/deen.exe ./cmd/deen
	fyne-cross darwin $(ldflags) --tags gui --arch=* -output ./bin/darwin-fyne/deen ./cmd/deen

.PHONY: web-assets
web-assets:
	GOOS=js GOARCH=wasm go build $(gomodflags) $(ldflagsstripped) -o internal/web/assets/deen.wasm ./cmd/deen
	cp $$(go env GOROOT)/lib/wasm/wasm_exec.js internal/web/assets/wasm_exec.js

# Build the wasm interface, embed it into a self-contained binary and serve it.
.PHONY: web
web: web-assets
	go build $(gomodflags) $(ldflagsstripped) --tags webembed -o ./bin/deen ./cmd/deen
	./bin/deen serve --port 9090

.PHONY: run
run:
	go run $(gomodflags) ./cmd/deen/main.go

.PHONY: clean
clean:
	rm -rf ./bin
	rm -f internal/web/assets/deen.wasm internal/web/assets/wasm_exec.js
	rm -f extras/web/deen.wasm extras/web/wasm_exec.js

.PHONY: test
test:
	go test $(gomodflags) -timeout 20s -count=1 -cover ./...

.PHONY: test-all
test-all: test
	go test $(gomodflags) -tags gui ./internal/gui
	PATH=$$(go env GOROOT)/lib/wasm:$$PATH GOOS=js GOARCH=wasm go test $(gomodflags) ./internal/webui
	npm --prefix extras/vscode-deen run compile
	npm --prefix extras/vscode-deen run lint
	$(MAKE) test-web-browser

.PHONY: bench
bench:
	go test $(gomodflags) -run '^$$' -bench . -benchmem ./internal/pipeline ./internal/plugins

.PHONY: bench-gui
bench-gui:
	go test $(gomodflags) -tags gui -run '^$$' -bench . -benchtime=1x -benchmem ./internal/gui

.PHONY: bench-webui
bench-webui:
	PATH=$$(go env GOROOT)/lib/wasm:$$PATH GOOS=js GOARCH=wasm go test $(gomodflags) -run '^$$' -bench . -benchmem ./internal/webui

.PHONY: bench-web-browser
bench-web-browser: web-assets
	npm --prefix extras/web-perf install
	npm --prefix extras/web-perf exec playwright install chromium
	npm --prefix extras/web-perf run bench

.PHONY: test-web-browser
test-web-browser: web-assets
	npm --prefix extras/web-perf install
	npm --prefix extras/web-perf exec playwright install chromium
	npm --prefix extras/web-perf test

.PHONY: all
all: clean build test
