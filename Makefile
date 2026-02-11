.PHONY: all build clean test lint

BINARY_NAME=adc
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS=-ldflags "-X github.com/api7/adc-go/internal/commands.Version=${VERSION} \
                  -X github.com/api7/adc-go/internal/commands.BuildTime=${BUILD_TIME} \
                  -X github.com/api7/adc-go/internal/commands.GitCommit=${GIT_COMMIT}"

all: build

build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/adc

build-linux:
	@echo "Building ${BINARY_NAME} for Linux..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/adc

build-mac:
	@echo "Building ${BINARY_NAME} for macOS..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/adc

build-windows:
	@echo "Building ${BINARY_NAME} for Windows..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/adc

clean:
	@echo "Cleaning..."
	rm -rf bin/

test:
	@echo "Running tests..."
	go test ./... -v

lint:
	@echo "Linting..."
	# TODO: Добавить линтер

install:
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	go install ${LDFLAGS} ./cmd/adc

release: clean build-linux build-mac build-windows
	@echo "Release builds created in bin/ directory"

version:
	@echo "Version: ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"
