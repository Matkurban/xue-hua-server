SHELL := /usr/bin/env bash

VERSION ?= dev
RELEASE ?= false
PLATFORMS ?=
TARGETS ?=
OUTPUT_DIR ?= _output
CONFIG_DIR ?= config
START_CONFIG ?= start-config.yml
JOBS ?= 4

MANAGE := VERSION="$(VERSION)" RELEASE="$(RELEASE)" PLATFORMS="$(PLATFORMS)" \
	OUTPUT_DIR="$(OUTPUT_DIR)" CONFIG_DIR="$(CONFIG_DIR)" START_CONFIG="$(START_CONFIG)" \
	JOBS="$(JOBS)" ./scripts/manage.sh

.DEFAULT_GOAL := build

.PHONY: all build build-only start stop restart check status pack export zip test test-race vet clean help

all: build

## build: Compile all server services and tools (supports PLATFORMS=linux_amd64, CGO_ENABLED=0, RELEASE=true)
build:
	@$(MANAGE) build

## build-only: Compile specific targets, e.g. make build-only TARGETS="server-api seq"
build-only:
	@if [ -z "$(strip $(TARGETS))" ]; then \
		echo "Usage: make build-only TARGETS=\"server-api seq\"" >&2; \
		exit 1; \
	fi
	@$(MANAGE) build "$(TARGETS)"

## start: Run pre-start tools and start all services defined in start-config.yml
start:
	@$(MANAGE) start

## stop: Stop all running services
stop:
	@$(MANAGE) stop

## restart: Stop then start all services
restart:
	@$(MANAGE) restart

## check: Check running status and listening ports of all services
check:
	@$(MANAGE) check

status: check

## pack: Package linux/amd64 binaries and runtime files into _output/server-linux-amd64.zip
pack:
	@$(MANAGE) pack

export: pack
zip: pack

## test: Run Go unit tests across the repository
test:
	@go test ./...

## test-race: Run Go unit tests with the race detector enabled
test-race:
	@go test -race ./...

## vet: Run go vet across the repository
vet:
	@go vet ./...

## clean: Remove build outputs under _output
clean:
	@rm -rf "$(OUTPUT_DIR)"

## help: Show available make targets
help:
	@awk '/^## / { sub(/^## /, ""); idx = index($$0, ":"); printf "  %-14s %s\n", substr($$0, 1, idx-1), substr($$0, idx+2) }' $(MAKEFILE_LIST)
