# Makefile template borrowed from https://sohlich.github.io/post/go_makefile/
PROJECT=genkubeconfig
PROJECT_VERSION=`cat VERSION.txt`
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVERSION=1.16
GOFLAGS="-X main.BuildVersion=$(PROJECT_VERSION)"
BINARY_NAME=$(PROJECT)

build-all: test build-linux build-arm build-darwin build-raspi
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps:
	# TODO: check for GO111MODULE
	GO111MODULE=on $(GOGET) github.com/docker/docker/client@master

prep:
	mkdir -p bin

# Cross compilation
build-linux: prep
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o "bin/$(BINARY_NAME)_linux" -ldflags $(GOFLAGS) -v "main.go"
build-arm: prep
	CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GOBUILD) -o "bin/$(BINARY_NAME)_arm" -ldflags $(GOFLAGS) -v "main.go"
build-raspi: prep
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o "bin/$(BINARY_NAME)_raspi" -ldflags $(GOFLAGS) -v "main.go"
build-darwin: prep
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o "bin/$(BINARY_NAME)_darwin" -ldflags $(GOFLAGS) -v "main.go"
build-docker: prep
	docker run --rm -it -v "$(GOPATH)":/go -w "/go/src/github.com/novu/$(PROJECT)" golang:$(GOVERSION) $(GOBUILD) -o "bin/$(BINARY_NAME)" -ldflags $(GOFLAGS) -v "main.go"
