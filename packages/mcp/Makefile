BINARY := pipebox-mcp
CMD := ./cmd/pipebox-mcp
GO_INDENT := python3 $(CURDIR)/../../scripts/go-indent.py

.PHONY: build clean test lint lint-indent lint-fix

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint-indent:
	@$(GO_INDENT) --check .

lint: lint-indent
	golangci-lint run ./...

# gofmt/editor may reintroduce tabs; expand before other fixes.
lint-fix:
	@$(GO_INDENT) --fix .
	golangci-lint run --fix ./...
