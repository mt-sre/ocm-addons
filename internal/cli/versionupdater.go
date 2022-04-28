package cli

import "context"

type VersionUpdater interface {
	LatestVersionGetter
	LatestPluginBinaryGetter
}

type LatestVersionGetter interface {
	GetLatestVersion(context.Context) (string, error)
}

type LatestPluginBinaryGetter interface {
	GetLatestPluginBinary(context.Context) ([]byte, error)
}
