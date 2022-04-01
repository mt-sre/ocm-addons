package ocm

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/output"
	"github.com/stretchr/testify/require"
)

func TestAddonInstallationInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(AddonInstallation))
}