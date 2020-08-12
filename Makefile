build:
ifeq ($(OS),Windows_NT)
	go build -o ./bin/deen.exe ./cmd/deen
else
	go build -o ./bin/deen ./cmd/deen
endif
