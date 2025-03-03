.PHONY: all clean build test


all: build


build:
	go build -o server ./cmd/main.go


test:
	go test ./tests... -v

clean:
	rm -rf bin/
