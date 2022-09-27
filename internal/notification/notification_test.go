// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package notification

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	assert.Equal(t, "Test", cfg.Description)
	assert.Equal(t, "Error", cfg.Severity)
	assert.Equal(t, "SREManualAction", cfg.ServiceName)
	assert.Equal(t, false, cfg.InternalOnly)
}

func setupTestFS(t *testing.T) fstest.MapFS {
	t.Helper()

	result := fstest.MapFS{
		"data/test-team/test-product.yaml": {
			Data: []byte(validConfig()),
		},
	}

	return result
}

func validConfig() string {
	result := []string{
		"---",
		"test-notification:",
		"  summary: TestNotificationSummary",
		"  description: Test",
		"  severity: Error",
	}

	return strings.Join(result, "\n")
}

func TestBadConfigs(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		Data string
	}{
		"top-level list": {
			Data: topLevelList(),
		},
		"top-level object": {
			Data: topLevelObject(),
		},
		"invalid type": {
			Data: invalidType(),
		},
		"invalid description": {
			Data: invalidDescription(),
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

			tree, err := loadNotifications(badFS)
			assert.Error(t, err, tree)
		})
	}
}

func topLevelList() string {
	result := []string{
		"---",
		"- test-notification:",
		"    summary: TestNotificationSummary",
		"    description: Test",
		"    severity: Error",
	}

	return strings.Join(result, "\n")
}

func topLevelObject() string {
	result := []string{
		"---",
		"summary: TestNotificationSummary",
		"description: Test",
		"severity: Error",
	}

	return strings.Join(result, "\n")
}

func invalidType() string {
	result := []string{
		"---",
		"test-notification:",
		"  summary: TestNotificationSummary",
		"  description: Test",
		"  severity: Error",
		"  internalOnly: some_string",
	}

	return strings.Join(result, "\n")
}

func invalidDescription() string {
	result := []string{
		"---",
		"test-notification:",
		"  summary: TestNotificationSummary",
		"  description: Test!",
		"  severity: Error",
	}

	return strings.Join(result, "\n")
}

func TestConfigInterfaces(t *testing.T) {
	t.Parallel()

	require.Implements(t, new(yaml.Unmarshaler), new(Config))
	require.Implements(t, new(cli.RowDataProvider), new(Config))
}
