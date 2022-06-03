package ocm_test

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/stretchr/testify/require"
)

func TestAddonInstallationInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.AddonInstallation))
}
