# SPDX-FileCopyrightText: 2025 Red Hat, Inc. <sd-mt-sre@redhat.com>
#
# SPDX-License-Identifier: Apache-2.0

build:
	goreleaser release --snapshot --clean

release:
	goreleaser release --rm-dist
.PHONY: release

check: lint test
.PHONY: check

lint:
	pre-commit run \
		--show-diff-on-failure \
		--from-ref "origin/main" \
		--to-ref "HEAD"
.PHONY: lint

BIN_PATH ?= dist/ocm-addons_linux_amd64_v1/ocm-addons

license-check:
	lichen -c .lichen.yaml "${BIN_PATH}"
.PHONY: license-check

golangci-lint:
	golangci-lint run -v --new-from-rev HEAD --fix
.PHONY: golangci-lint

tidy:
	go mod tidy
.PHONY: tidy

verify:
	go mod verify
.PHONY: verify

test: test-units test-integration

test-units:
	go test -v -race -count=1 -v ./cmd/... ./internal/...
.PHONY: test-units

test-integration:
	ginkgo -r -v \
		--randomize-all \
		--randomize-suites \
		--fail-on-pending \
		--keep-going \
		--race \
		--trace \
		integration
.PHONY: test-integration
