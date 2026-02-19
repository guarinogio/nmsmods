# Makefile
BIN := nmsmods
PREFIX ?= $(HOME)/.local/bin

.PHONY: tidy test vet build install uninstall

tidy:
	go mod tidy

test:
	go test ./...

vet:
	go vet ./...

build: tidy
	go build -o $(BIN) ./

install: build
	mkdir -p $(PREFIX)
	install -m 0755 $(BIN) $(PREFIX)/$(BIN)
	@echo "Installed to $(PREFIX)/$(BIN)"

uninstall:
	rm -f $(PREFIX)/$(BIN)
	@echo "Removed $(PREFIX)/$(BIN)"
