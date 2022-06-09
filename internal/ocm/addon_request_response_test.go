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
