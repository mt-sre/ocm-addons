// SPDX-FileCopyrightText: 2023 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

//go:build windows
// +build windows

package run

import (
	"os"
)

func shutdownSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
