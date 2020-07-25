build:
	go build -o ./bin/deen ./cmd/deen

run:
	go run ./cmd/deen/main.go

build-all: build-freebsd build-macos build-linux build-windows

build-freebsd:
	GOOS=freebsd GOARCH=386 go build -o bin/deen-freebsd-x86 ./cmd/deen
	GOOS=freebsd GOARCH=amd64 go build -o bin/deen-freebsd-x86_64 ./cmd/deen

build-macos:
	GOOS=darwin GOARCH=386 go build -o bin/deen-darwin-x86 ./cmd/deen
	GOOS=darwin GOARCH=amd64 go build -o bin/deen-darwin-x86_64 ./cmd/deen

build-linux:
	GOOS=linux GOARCH=386 go build -o bin/deen-linux-x86 ./cmd/deen
	GOOS=linux GOARCH=amd64 go build -o bin/deen-linux-x86_64 ./cmd/deen

build-windows:
	GOOS=windows GOARCH=386 go build -o bin/deen-windows-x86.exe ./cmd/deen
	GOOS=windows GOARCH=amd64 go build -o bin/deen-windows-x86_64.exe ./cmd/deen
