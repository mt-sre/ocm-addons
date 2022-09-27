// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

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
