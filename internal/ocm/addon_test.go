package ocm

import (
	"testing"

	"github.com/mt-sre/ocm-addons/internal/output"
	"github.com/stretchr/testify/require"
)

func TestAddonInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(Addon))
}

func TestAddonParameterInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(AddonParameter))
}

func TestAddonRequirementInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(AddonRequirement))
}

func TestAddonSubOperatorInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(AddonSubOperator))
}

func TestAddonVersionInterfaces(t *testing.T) {
	require.Implements(t, new(output.RowDataProvider), new(AddonVersion))
}
