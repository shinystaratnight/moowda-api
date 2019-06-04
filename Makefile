MAIN_VERSION:=$(shell git describe --abbrev=0 --tags || echo "0.1")
VERSION:=${MAIN_VERSION}\#$(shell git log -n 1 --pretty=format:"%h")
PACKAGES:=$(shell go list ./... | sed -n '1!p' | grep -v /vendor/)
LDFLAGS:=-ldflags
default: build

cover: test
	go tool cover -html=coverage-all.out

build: clean
	go build -a -o api

clean:
	rm -rf server coverage.out coverage-all.out
