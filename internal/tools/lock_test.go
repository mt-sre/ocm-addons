package tools

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func DummyLock() Lock {
	return Lock{
		Dependencies: DependencyManifest{
			Items: []Dependency{
				{
					Name:    "linter",
					Version: "1.0.0",
				},
				{
					Name:    "checker",
					Version: "2.0.0",
				},
				{
					Name:    "builder",
					Version: "3.0.0",
				},
			},
		},
	}
}

func TestReadWriteLock(t *testing.T) {
	t.Parallel()

	filePath := path.Join(t.TempDir(), "cfg.yaml")

	err := DumpLock(filePath, DummyLock())
	require.NoError(t, err)
	require.FileExists(t, filePath)

	man, err := LoadLock(filePath)
	require.NoError(t, err)
	require.Equal(t, *man, DummyLock())
}
