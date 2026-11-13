BINARY := pipebox
CMD := ./cmd/pipebox

.PHONY: build clean test lint lint-fix doctor

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
