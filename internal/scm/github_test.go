package scm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/ocm-addons/internal/cli"
)

func TestGitHubClientInterfaces(t *testing.T) {
	require.Implements(t, new(cli.VersionUpdater), new(GitHubClient))
}
