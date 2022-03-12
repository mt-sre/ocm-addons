package notification

import (
	"testing"
	"testing/fstest"

	"github.com/mt-sre/ocm-addons/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGetTeams(t *testing.T) {
	t.Parallel()

	testFS := setupTestFS(t)

	tree, err := loadNotifications(testFS)
	require.NoError(t, err)

	teams := tree.GetTeams()
	assert.ElementsMatch(t, []string{"test-team"}, teams)
}

func TestGetProducts(t *testing.T) {
	t.Parallel()

	testFS := setupTestFS(t)

	tree, err := loadNotifications(testFS)
	require.NoError(t, err)

	teams := tree.GetProducts("test-team")
	assert.ElementsMatch(t, []string{"test-product"}, teams)
}

func TestGetNotification(t *testing.T) {
	t.Parallel()

	testFS := setupTestFS(t)

	tree, err := loadNotifications(testFS)
	require.NoError(t, err)

	cfg, ok := tree.GetNotification("test-team", "test-product", "test-notification")
	assert.True(t, ok)
	assert.Equal(t, "TestNotificationSummary", cfg.Summary)
}

const _testConfig = `
---
test-notification:
  summary: "TestNotificationSummary"
`

func setupTestFS(t *testing.T) fstest.MapFS {
	t.Helper()

	result := fstest.MapFS{
		"data/test-team/test-product.yaml": {
			Data: []byte(_testConfig),
		},
	}

	return result
}

func TestBadConfigs(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		Data string
	}{
		"top-level list": {
			Data: `---\n- test-notification:\n    summary: "TestNotificationSummary"\n`,
		},
		"top-level object": {
			Data: `---\nsummary: "TestNotificationSummary"\n`,
		},
		"invalid type": {
			Data: `---\ntest-notification:\n  internalOnly: arbitrary_string\n`,
		},
	} {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			badFS := fstest.MapFS{
				"data/test-team/test-product.yaml": {
					Data: []byte(tc.Data),
				},
			}

			_, err := loadNotifications(badFS)
			assert.Error(t, err)
		})
	}
}

func TestConfigInterfaces(t *testing.T) {
	t.Parallel()

	require.Implements(t, new(yaml.Unmarshaler), new(Config))
	require.Implements(t, new(output.RowDataProvider), new(Config))
}
