.PHONY: build test cover lint clean

build:
	go build -o bin/gitai

test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

clean:
	rm -rf bin/ coverage.out 