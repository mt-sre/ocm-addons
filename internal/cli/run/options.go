// SPDX-FileCopyrightText: 2023 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package run

type WithErrHandler func(error)

func (w WithErrHandler) ConfigureRunner(c *RunnerConfig) {
	c.ErrHandler = w
}
