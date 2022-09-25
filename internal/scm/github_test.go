// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package scm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/ocm-addons/internal/cli"
)

func TestGitHubClientInterfaces(t *testing.T) {
	require.Implements(t, new(cli.VersionUpdater), new(GitHubClient))
}
