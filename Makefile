GO=go
GOFLAG=-a -installsuffix cgo
PROJ=tricarboxylic
OBJ=tricarb

repo=github.com/GreysTone/$(PROJ)
ver=$(shell cat ./VERSION)
time=$(shell date "+%m/%d/%Y %R %Z")
hash=$(shell git rev-parse --short HEAD)
gover=$(shell go version)

LDFLAG=-ldflags '-X "$(repo)/config.buildVersion=$(ver)" -X "$(repo)/config.buildTime=$(time)" -X "$(repo)/config.buildHash=$(hash)" -X "$(repo)/config.goVersion=$(gover)"'

.PHONY: cli-nix-amd64 cli-nix-arm64 dmn-nix-amd64 dmn-nix-arm64

golang-proto:
	go version
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	go get -u google.golang.org/grpc
	cd rpc; protoc --go_out=plugins=grpc:. *.proto; cd -

cli-nix-amd64: main-cli.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<

cli-nix-arm64: main-cli.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<

dmn-nix-amd64: main-dmn.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<

dmn-nix-arm64: main-dmn.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<

#windows-amd64: main.go
#	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<
	# mv $(OBJ)-$@ bin
	# docker build -t $(PROJ)-$@:$(ver) .
	# rm bin

#darwin-amd64: main.go
#	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<
