.PHONY: build test clean

build:
	go build -o faaah ./cmd/faaah/

test:
	go test ./...

clean:
	rm -f faaah
