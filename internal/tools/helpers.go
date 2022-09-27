// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"os/exec"

	"github.com/magefile/mage/mg"
)

// GoVerboseFlag checks if 'Mage' was run with the verbose flag and
// if so returns the Go verbose flag '-v' as a slice of strings.
// An empty slice of strings is returned otherwise.
func GoVerboseFlag() []string {
	if !mg.Verbose() {
		return []string{}
	}

	return []string{"-v"}
}

// Runtime attempts to find an available container runtime in the PATH.
// The path to the first available runtime is returned along with a boolean
// value indicating if any runtimes were found.
func Runtime() (string, bool) {
	prefferedRuntimes := []string{
		"podman",
		"docker",
	}

	for _, runtime := range prefferedRuntimes {
		runtimePath, err := exec.LookPath(runtime)
		if err == nil {
			return runtimePath, true
		}
	}

	return "", false
}
