# Variables
BUILD_DIR 		:= build
GITHASH 		:= $(shell git rev-parse HEAD)
VERSION			:= $(shell git describe --abbrev=0 --tags --always)
DATE			:= $(shell TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ UTC')
LINT_PATHS		:= ./...
FORMAT_PATHS 	:= .

# Compilation variables
CC 					:= go build
DFLAGS 				:= -race
CFLAGS 				:= -X 'main.githash=$(GITHASH)' \
            -X 'main.date=$(DATE)' \
            -X 'main.version=$(VERSION)'
CROSS				:= GOOS=linux GOARCH=amd64

# Makefile variables
VPATH 				:= $(BUILD_DIR)


.SECONDEXPANSION:
.PHONY: all
all: init build install rpm deb 

.PHONY: init
init:
	go get -u github.com/golang/dep/...
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/modocache/gover
	go get -u github.com/goreleaser/nfpm
	$(GOPATH)/bin/gometalinter --install --no-vendored-linters
	

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist

.PHONY: format
format:
	gofmt -w -s $(FORMAT_PATHS)

.PHONY: lint
lint:
	$(GOPATH)/bin/gometalinter --disable-all --config .gometalinter.json $(LINT_PATHS)

.PHONY: test
test:
	$(GOPATH)/bin/ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --progress --compilers=2

.PHONY: testrun
testrun:
	$(GOPATH)/bin/ginkgo watch -r ./

.PHONY: cover
cover:
	$(GOPATH)/bin/gover . coverage.txt

.PHONY: dev
dev: format build

.PHONY: build
build:
	$(CC) $(DFLAGS) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/ovh-spark-submit

.PHONY: release
release:
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/ovh-spark-submit

.PHONY: dist
dist:
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/ovh-spark-submit

.PHONY: install
install: release
	cp -v $(BUILD_DIR)/ovh-spark-submit $(GOPATH)/bin/ovh-spark-submit

.PHONY: deb
deb:
		rm -f ./build/ovh-spark-submit*.deb
		nfpm pkg --target ./build/ovh-spark-submit.deb
		

.PHONY: rpm
rpm:
		rm -f ovh-spark-submit*.rpm
		nfpm pkg --target ./build/ovh-spark-submit.rpm

