// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClustersListRequestInterfaces(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	assert.Implements(
		(*clustersListRequester)(nil),
		new(clustersListRequest),
		"should implement clustersListRequester interface",
	)
}

func TestClustersListResponseInterfaces(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	assert.Implements(
		(*clustersListResponser)(nil),
		new(clustersListResponse),
		"should implement clustersListResponser interface",
	)
}
