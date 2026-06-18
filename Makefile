.PHONY: build run test tidy seed clean

GO ?= go
BIN ?= bin/server

tidy:
	$(GO) mod tidy

build: tidy
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/server

run: build
	./$(BIN)

test:
	$(GO) test ./tests/... -v

seed:
	$(GO) run ./cmd/server --seed-admin=true

clean:
	rm -rf bin
