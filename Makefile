VERSION=0.7.2
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH=$(shell git rev-parse HEAD)

BINARIES=bin/vopher-$(VERSION).linux.amd64 \
		 bin/vopher-$(VERSION).linux.386 \
		 bin/vopher-$(VERSION).linux.arm64 \
		 bin/vopher-$(VERSION).linux.mips64 \
		 bin/vopher-$(VERSION).windows.amd64.exe \
		 bin/vopher-$(VERSION).freebsd.amd64 \
		 bin/vopher-$(VERSION).darwin.amd64


LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitHash=$(GIT_HASH)"


simple:
	cd cmd/vopher && go build -o ../../vopher -v

test:
	cd pkg/utils && go test -v
	cd pkg/plugin && go test -v
	cd cmd/vopher && go test -v

release: $(BINARIES)

build-docker:
	docker build -f docker/Dockerfile -t vopher \
		--build-arg BUILD_DIR=/vopher/src/vopher \
		--build-arg VOPHER=bin/vopher-$(VERSION).linux.amd64 .


bin/vopher-$(VERSION).linux.mips64: bin
	env GOOS=linux GOARCH=mips64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).linux.amd64: bin
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

bin/vopher-$(VERSION).linux.386: bin
	env GOOS=linux GOARCH=386 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

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
