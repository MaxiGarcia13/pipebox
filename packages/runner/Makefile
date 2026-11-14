BINARY := pipebox
CMD := ./cmd/pipebox
ACT_IMAGE := pipebox-act:latest

.PHONY: build clean test lint lint-indent lint-fix doctor image

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

# gofmt uses tabs; this repo uses 2-space indent (.editorconfig).
lint-indent:
	@if grep -RIn --include='*.go' $$'\t' .; then \
		echo >&2 "error: Go files must use 2-space indent (no tabs; see .editorconfig)"; \
		exit 1; \
	fi

lint: lint-indent
	golangci-lint run ./...

lint-fix: lint-indent
	golangci-lint run --fix ./...

doctor: build
	./$(BINARY) doctor

image:
	docker build -t $(ACT_IMAGE) ./docker/act
