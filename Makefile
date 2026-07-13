# TradelogCLI — build e release del instalador del SDK iOS.

BINARY   := tradelog
PKG      := ./cmd/tradelog
VERSION  ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
DIST     := dist

.PHONY: build install test clean release

## Compila el binario para la plataforma actual.
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

## Instala en /usr/local/bin (local).
install: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

test:
	go vet ./...
	go test ./...

clean:
	rm -rf $(BINARY) $(DIST)

## Cross-compila macOS (arm64 + amd64) y arma tar.gz + shasums para Homebrew.
release: clean
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY) $(PKG)
	@cd $(DIST) && tar -czf $(BINARY)_$(VERSION)_darwin_arm64.tar.gz $(BINARY) && rm $(BINARY)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY) $(PKG)
	@cd $(DIST) && tar -czf $(BINARY)_$(VERSION)_darwin_amd64.tar.gz $(BINARY) && rm $(BINARY)
	@cd $(DIST) && shasum -a 256 *.tar.gz | tee SHASUMS256.txt
	@echo "→ Artefactos en $(DIST)/ (sube a la release de GitHub y actualiza la fórmula)"
