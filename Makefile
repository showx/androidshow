VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)
BIN := bin/ashow
ifeq ($(OS),Windows_NT)
	BIN := bin/ashow.exe
endif

.PHONY: build dist test clean

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/ashow

dist:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-windows-amd64.exe ./cmd/ashow
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-windows-arm64.exe ./cmd/ashow
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-linux-amd64 ./cmd/ashow
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-linux-arm64 ./cmd/ashow
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-darwin-amd64 ./cmd/ashow
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ashow-darwin-arm64 ./cmd/ashow

test:
	go test ./...

clean:
	rm -rf bin dist
