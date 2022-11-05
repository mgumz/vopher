VERSION=$(shell cat VERSION)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH=$(shell git rev-parse HEAD)

TARGETS=linux.amd64 linux.386 linux.arm64 linux.mips64 windows.amd64.exe darwin.amd64 darwin.arm64 freebsd.amd64
BINARIES=$(addprefix bin/vopher-$(VERSION)., $(TARGETS))
RELEASES=$(subst windows.amd64.tar.gz,windows.amd64.zip,$(foreach r,$(subst .exe,,$(TARGETS)),releases/vopher-$(VERSION).$(r).tar.gz))

LDFLAGS=-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitHash=$(GIT_HASH)
TAGS=


toc:
	@echo "list of targets:"
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | \
		awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | \
		sort | \
		egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | \
		awk '{ print " ", $$1 }'

binaries: $(BINARIES)
releases: $(RELEASES)
	make $(RELEASES)

vopher: bin/vopher
vopher-full:
	make TAGS="-tags lzma,zstd" vopher
vopher-small:
	make LDFLAGS="$(LDFLAGS) -s -w" vopher
bin/vopher:
	go build $(TAGS) -ldflags "$(LDFLAGS)" -v -o $@ ./cmd/vopher

bin/vopher-$(VERSION).%:
	env GOARCH=$(subst .,,$(suffix $(subst .exe,,$@))) \
		GOOS=$(subst .,,$(suffix $(basename $(subst .exe,,$@)))) \
		CGO_ENABLED=0 \
		go build -ldflags "$(LDFLAGS)" -o $@ ./cmd/vopher

releases/vopher-$(VERSION).%.zip: bin/vopher-$(VERSION).%.exe
	mkdir -p releases
	zip -9 -j -r $@ README.md $<
releases/vopher-$(VERSION).%.tar.gz: bin/vopher-$(VERSION).%
	mkdir -p releases
	tar -cf $(basename $@) README.md && \
		tar -rf $(basename $@) --strip-components 1 $< && \
		gzip -9 $(basename $@)


deps-vendor:
	go mod vendor
deps-cleanup:
	go mod tidy


test:
	cd pkg/utils && go test -v
	cd pkg/plugin && go test -v
	cd cmd/vopher && go test -v

release: $(BINARIES)

build-docker:
	docker build -f docker/Dockerfile -t vopher \
		--build-arg BUILD_DIR=/vopher/src/vopher \
		--build-arg VOPHER=bin/vopher-$(VERSION).linux.amd64 .

# https://github.com/nektos/act
run-github-workflow-lint:
	act -j lint --container-architecture linux/amd64
run-github-workflow-test:
	act -j test --container-architecture linux/amd64
run-github-workflow-buildLinux:
	act -j buildLinux --container-architecture linux/amd64

report: report-cyclo report-staticcheck report-mispell report-ineffassign report-vet
report-cyclo:
	@echo '####################################################################'
	gocyclo ./cmd/vopher
report-mispell:
	@echo '####################################################################'
	misspell ./cmd/
report-lint:
	@echo '####################################################################'
	golint ./cmd/... ./pkg/...
report-ineffassign:
	@echo '####################################################################'
	ineffassign ./cmd/... ./pkg/...
report-vet:
	@echo '####################################################################'
	go vet ./cmd/... ./pkg/...
report-staticcheck:
	@echo '####################################################################'
	staticcheck ./cmd/... ./pkg/...

fetch-report-tools:
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/gordonklaus/ineffassign@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/lint/golint@latest

.PHONY: vopher bin/vopher
