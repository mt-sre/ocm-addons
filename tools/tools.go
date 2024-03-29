// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "github.com/openshift-online/ocm-cli/cmd/ocm"
	_ "github.com/uw-labs/lichen"
)
