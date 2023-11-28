# Copyright (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package_version_name := github.com/intel/powertelemetry/internal/version
version := $(shell cat VERSION)
tag := $(shell git describe --exact-match --tags 2>/dev/null)

branch := $(shell git rev-parse --abbrev-ref HEAD)
commit := $(shell git rev-parse --short=8 HEAD)

LDFLAGS := $(LDFLAGS) -X $(package_version_name).Commit=$(commit) -X $(package_version_name).Branch=$(branch)
ifneq ($(tag),)
	LDFLAGS += -X $(package_version_name).Version=$(version)
else
	LDFLAGS += -X $(package_version_name).Version=$(version)-$(commit)
endif

GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(GOFILES))

build:
	go build -ldflags "$(LDFLAGS)" ./cmd/example

test:
	go test -race -cover -v ./...

coverage:
	go test ./... -coverprofile=coverage.out

fmtcheck:
	@if [ ! -z "$(GOFMT)" ]; then \
		echo "[ERROR] gofmt has found errors in the following files:"  ; \
		echo "$(GOFMT)" ; \
		echo "" ;\
		echo "Run make fmt to fix them." ; \
		exit 1 ;\
	fi

tidy:
	go mod verify
	go mod tidy
	go fix ./...

.PHONY : build test coverage fmtcheck tidy
