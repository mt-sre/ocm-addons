// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/stretchr/testify/require"
)

func TestLogEntryInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.LogEntry))
}
