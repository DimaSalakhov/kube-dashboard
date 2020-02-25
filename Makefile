ARTIFACT = kd

GOARCH ?= amd64
GOOS ?= $(shell uname | tr '[:upper:]' '[:lower:]')

build:
	@echo ">> Building"
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o "$(ARTIFACT)" -a .

run: build
	./$(ARTIFACT)
