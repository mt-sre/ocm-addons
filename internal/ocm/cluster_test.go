package ocm

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/output"
	"github.com/stretchr/testify/require"
)

func TestClusterInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(Cluster))
}

func TestSubscriptionInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(Subscription))
}
