// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddonsListRequestInterfaces(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	assert.Implements(
		(*addonsListRequester)(nil),
		new(addonsListRequest),
		"should implement addonsListRequester interface",
	)
}

func TestAddonsListResponseInterfaces(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	assert.Implements(
		(*addonsListResponser)(nil),
		new(addonsListResponse),
		"should implement addonsListResponser interface",
	)
}
