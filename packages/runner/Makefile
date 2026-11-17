BINARY := pipebox
CMD := ./cmd/pipebox
ACT_IMAGE := pipebox-act:latest
JENKINS_IMAGE := pipebox-jenkins:latest

.PHONY: build clean test lint lint-indent lint-fix doctor image image-act image-jenkins

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

# gofmt uses tabs; this repo uses 2-space indent (.editorconfig).
# printf: $'\t' is not POSIX. dash keeps it as the pattern "$t".
lint-indent:
	@if grep -RInF --include='*.go' "$$(printf '\t')" .; then \
		echo >&2 "error: Go files must use 2-space indent (no tabs; see .editorconfig)"; \
		exit 1; \
	fi

lint: lint-indent
	golangci-lint run ./...

# Expand tabs → 2 spaces (gofmt/editor may reintroduce tabs).
expand-go-tabs:
	@python3 -c 'from pathlib import Path; [p.write_text(p.read_text().replace(chr(9), "  ")) for p in Path(".").rglob("*.go")]'

lint-fix: expand-go-tabs
	golangci-lint run --fix ./...

doctor: build
	./$(BINARY) doctor

image: image-act image-jenkins

image-act:
	docker build -t $(ACT_IMAGE) ./docker/act

image-jenkins:
	docker build -t $(JENKINS_IMAGE) ./docker/jenkins
