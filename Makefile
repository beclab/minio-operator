
.PHONY: all tidy fmt vet build

all: tidy build

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

build: fmt vet ;$(info $(M)...Begin to build minio-operator.) @
	go build -o bin/minio-operator cmd/main.go

build-linux: fmt vet ;$(info $(M)...Begin to build minio-operator (linux version).) @
	GOOS=linux GOARCH=amd64 go build -o bin/minio-operator cmd/main.go
