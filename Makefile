.DEFAULT_GOAL := help

VERSION := $(shell git describe --tags --abbrev=0)
VERSION_LONG := $(shell git describe --tags)
VAR_VERSION := main.Version

LDFLAGS := -ldflags "-X $(VAR_VERSION)=$(VERSION) \
		     -X $(VAR_VERSION)Long=$(VERSION_LONG)"

#SRCS := $(shell find . -type f -name '*.go')
SRCS := $(shell \ls *.go | grep -v _test.go)

.PHONY: build
build: es-query  ## Build here

es-query: $(SRCS)
	time go build $(LDFLAGS) -o es-query

.PHONY: install
install:  ## Install in GOPATH
	time go install $(LDFLAGS) .

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2> /dev/null),)
	go get github.com/Masterminds/glide
endif

.PHONY: test
test:  ## Test all packages
	go test --ldflags -s `glide novendor`

vet:
	go vet `glide novendor`

.PHONY: clean
clean:  ## Clean up temporary files
	rm -f *.test *.out

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

##
## Release tasks
##
release: release-all  ## Release on github.com

release-all: release-build \
	 release-compress \
	 release-upload

release-upload:
	ghr -u tmtk75 --prerelease $(VERSION) pkg

release-compress:
	#mv pkg/es-query_windows_amd64 pkg/es-query_windows_amd64.exe
	#chmod +x pkg/*
	#zip pkg/es-query_windows_amd64.zip pkg/es-query_windows_amd64.exe
	#rm pkg/es-query_windows_amd64.exe
	gzip --force pkg/es-query_darwin_amd64 pkg/es-query_linux_amd64

XC_ARCH := amd64
XC_OS := darwin linux # windows

release-build:
	for arch in $(XC_ARCH); do \
	  for os in $(XC_OS); do \
	    echo $$arch $$os; \
	    GOARCH=$$arch GOOS=$$os time go build \
	      -o pkg/es-query_$${os}_$$arch \
	      $(LDFLAGS) \
	      $(SRCS); \
	  done \
	done\
