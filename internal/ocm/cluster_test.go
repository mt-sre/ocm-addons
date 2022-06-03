package ocm_test

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/stretchr/testify/require"
)

func TestClusterInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.Cluster))
}

func TestSubscriptionInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.Subscription))
}
