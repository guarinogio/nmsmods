# Makefile
BIN := nmsmods
PREFIX ?= $(HOME)/.local/bin
DESKTOP_DIR := $(HOME)/.local/share/applications
DESKTOP_FILE := $(DESKTOP_DIR)/nmsmods.desktop

.PHONY: tidy fmt vet build install uninstall vuln install-handler

fmt:
	test -z "$$(gofmt -l .)"

vuln:
	govulncheck ./...

check: fmt tidy vet build

tidy:
	go mod tidy

vet:
	go vet ./...

build: tidy
	go build -o $(BIN) ./

install: build
	mkdir -p $(PREFIX)
	install -m 0755 $(BIN) $(PREFIX)/$(BIN)
	@$(MAKE) install-handler
	@echo "Installed to $(PREFIX)/$(BIN)"
	@if ! echo ":$$PATH:" | grep -q ":$(PREFIX):" ; then \
		echo "NOTE: $(PREFIX) is not on your PATH in this shell."; \
		echo "      Open a new terminal or add $(PREFIX) to PATH (see README)."; \
	fi

install-handler:
	@mkdir -p $(DESKTOP_DIR)
	@printf '%s\n' \
		"[Desktop Entry]" \
		"Type=Application" \
		"Name=nmsmods" \
		"Exec=$(PREFIX)/$(BIN) nxm handle %u" \
		"Terminal=false" \
		"NoDisplay=true" \
		"MimeType=x-scheme-handler/nxm;" \
		"Categories=Game;" \
		> $(DESKTOP_FILE)
	@command -v xdg-mime >/dev/null 2>&1 && xdg-mime default nmsmods.desktop x-scheme-handler/nxm || true
	@echo "Registered nxm:// handler (x-scheme-handler/nxm)"

uninstall:
	rm -f $(PREFIX)/$(BIN)
	@rm -f $(DESKTOP_FILE) || true
	@echo "Removed $(PREFIX)/$(BIN)"
