APP = course_project
PROJECT = github.com/DaryaFesenko/course_project
STDERR=/tmp/.$(APP)-stderr.txt

HAS_LINT := $(shell command -v golangci-lint;)
HAS_IMPORTS := $(shell command -v goimports;)

all: run

lint: bootstrap
	@echo "+ $@"
	@golangci-lint run

run: clean build
	@echo "+ $@"
	./${APP}

build: lint
	@echo "+ $@"
	@go build

clean:
	@echo "+ $@"
	@rm -f ./${APP}

bootstrap:
	@echo "+ $@"
ifndef HAS_LINT
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
endif
ifndef HAS_IMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

test: 
	@go test -v -coverprofile ../cover.out ./...
	
.PHONY: all \
	lint \
	run \
	build \
	clean \
	bootstrap