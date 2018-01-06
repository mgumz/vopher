VERSION=0.5.0
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH=$(shell git rev-parse HEAD)

BINARIES=bin/vopher-$(VERSION).linux.amd64 \
		 bin/vopher-$(VERSION).linux.arm64 \
		 bin/vopher-$(VERSION).linux.mips64 \
		 bin/vopher-$(VERSION).windows.amd64.exe \
		 bin/vopher-$(VERSION).freebsd.amd64 \
		 bin/vopher-$(VERSION).darwin.amd64


LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitHash=$(GIT_HASH)"


simple:
	go build -v

release: $(BINARIES)

bin/vopher-$(VERSION).linux.mips64: bin
	env GOOS=linux GOARCH=mips64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).linux.amd64: bin
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).linux.arm64: bin
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).windows.amd64.exe: bin
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).darwin.amd64: bin
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).freebsd.amd64: bin
	env GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin:
	mkdir $@
