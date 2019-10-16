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

.PHONY: linux-amd64 linux-arm64 windows-amd64 darwin-amd64

linux-amd64: main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<
	# mv $(OBJ)-$@ bin
	# docker build -t $(PROJ)-$@:$(ver) .
	# rm bin

linux-arm64: main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<
	# mv $(OBJ)-$@ bin
	# docker build -t $(PROJ)-$@:$(ver) .
	# rm bin

windows-amd64: main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<

darwin-amd64: main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAG) $(LDFLAG) -o $(OBJ)-$@ $<
