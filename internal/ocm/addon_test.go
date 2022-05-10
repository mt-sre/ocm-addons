package ocm_test

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/stretchr/testify/require"
)

func TestAddonInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.Addon))
}

func TestAddonParameterInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.AddonParameter))
}

func TestAddonRequirementInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.AddonRequirement))
}

func TestAddonSubOperatorInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.AddonSubOperator))
}

func TestAddonVersionInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.AddonVersion))
}

func TestCredentialRequestInterfaces(t *testing.T) {
	require.Implements(t, new(cli.RowDataProvider), new(ocm.CredentialRequest))
}
