PKGS := $(shell go list ./... | grep -v /vendor)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: test

test:
	$(GOTEST) -v $(PKGS)

.PHONY: clean

clean:
	$(GOCLEAN)
	rm -r release

BINARY := gojet
VERSION ?= vlatest
PLATFORMS := linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)

$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 $(GOBUILD) -o release/$(BINARY)-$(VERSION)-$(os)-amd64 main/*.go

.PHONY: release
release: linux darwin
