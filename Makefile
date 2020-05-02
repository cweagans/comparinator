SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.PHONY: bin
bin: bin/windows/comparinator.exe bin/linux/comparinator bin/darwin/comparinator

bin/windows/comparinator.exe: vendor main.go results.go
	env GOOS=windows GOARCH=amd64 go build -o bin/windows/comparinator.exe

bin/linux/comparinator: vendor main.go results.go
	env GOOS=linux GOARCH=amd64 go build -o bin/linux/comparinator

bin/darwin/comparinator: vendor main.go results.go
	env GOOS=darwin GOARCH=amd64 go build -o bin/darwin/comparinator

.PHONY: release
release: bin release/comparinator.win64.tar.gz release/comparinator.linux64.tar.gz release/comparinator.darwin64.tar.gz release/sha256sums.txt

release/sha256sums.txt: release/comparinator.win64.tar.gz release/comparinator.linux64.tar.gz release/comparinator.darwin64.tar.gz
	shasum --algorithm 256 release/* > release/sha256sums.txt

release/comparinator.win64.tar.gz: bin/windows/comparinator.exe
	[ -d release ] || mkdir release
	tar czf release/comparinator.win64.tar.gz -C bin/windows comparinator.exe

release/comparinator.linux64.tar.gz: bin/linux/comparinator
	[ -d release ] || mkdir release
	tar czf release/comparinator.linux64.tar.gz -C bin/linux comparinator

release/comparinator.darwin64.tar.gz: bin/darwin/comparinator
	[ -d release ] || mkdir release
	tar czf release/comparinator.darwin64.tar.gz -C bin/darwin comparinator

vendor: go.sum
	go mod vendor

go.sum: go.mod

go.mod:

clean:
	rm -r bin
	rm -r release
