BINARY := pipebox-mcp
CMD := ./cmd/pipebox-mcp

.PHONY: build clean test lint lint-indent lint-fix

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint-indent:
	@if grep -RInF --include='*.go' "$$(printf '\t')" .; then \
		echo >&2 "error: Go files must use 2-space indent (no tabs; see .editorconfig)"; \
		exit 1; \
	fi

lint: lint-indent
	go vet ./...

expand-go-tabs:
	@python3 -c 'from pathlib import Path; [p.write_text(p.read_text().replace(chr(9), "  ")) for p in Path(".").rglob("*.go")]'

lint-fix: expand-go-tabs
	go vet ./...
