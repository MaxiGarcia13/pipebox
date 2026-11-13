BINARY := pipebox
CMD := ./cmd/pipebox
ACT_IMAGE := pipebox-act:latest

.PHONY: build clean test lint lint-fix doctor image

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

doctor: build
	./$(BINARY) doctor

image:
	docker build -t $(ACT_IMAGE) ./docker/act
