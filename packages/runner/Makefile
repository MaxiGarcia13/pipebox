BINARY := pipebox
CMD := ./cmd/pipebox
ACT_IMAGE := pipebox-act:latest
JENKINS_IMAGE := pipebox-jenkins:latest
JENKINS_REVISION := $(shell sed -n 's/^[[:space:]]*Revision = "\(.*\)"/\1/p' internal/provider/jenkins/jfrimage/image.go)
GO_INDENT := node $(CURDIR)/../../scripts/go-indent.mjs

.PHONY: build clean test test-e2e lint lint-indent lint-fix doctor image image-act image-jenkins

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

test-e2e:
	go test -tags e2e -count=1 -timeout 30m ./internal/e2e/

# gofmt uses tabs; this repo uses 2-space indent (.editorconfig).
lint-indent:
	@$(GO_INDENT) --check .

lint: lint-indent
	golangci-lint run ./...

# gofmt/editor may reintroduce tabs; expand before other fixes.
lint-fix:
	@$(GO_INDENT) --fix .
	golangci-lint run --fix ./...

doctor: build
	./$(BINARY) doctor

image: image-act image-jenkins

image-act:
	docker build -t $(ACT_IMAGE) ./docker/act

image-jenkins:
	docker build \
		--build-arg PIPEBOX_JENKINS_REVISION=$(JENKINS_REVISION) \
		-t $(JENKINS_IMAGE) \
		./docker/jenkins
