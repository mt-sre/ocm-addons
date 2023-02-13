// SPDX-FileCopyrightText: 2023 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

const sigExitMarker = 128

func NewRunner(opts ...RunnerOption) *Runner {
	var cfg RunnerConfig

	cfg.Option(opts...)

	return &Runner{
		cfg: cfg,
	}
}

type Runner struct {
	cfg RunnerConfig
}

func (r *Runner) Run(f func(context.Context) error) int {
	var code int

	ctx, cancel := context.WithCancel(context.Background())

	signals := shutdownSignals()

	sigRecv := make(chan os.Signal, len(signals))

	signal.Notify(sigRecv, signals...)
	defer signal.Stop(sigRecv)

	go func() {
		code = sigExitMarker + sigToInt(<-sigRecv)

		cancel()

		code = sigExitMarker + sigToInt(<-sigRecv)
	}()

	if err := f(ctx); err != nil && !errors.Is(err, context.Canceled) {
		if r.cfg.ErrHandler != nil {
			r.cfg.ErrHandler(err)
		}

		code = 1
	}

	return code
}

func sigToInt(sig os.Signal) int {
	if sysSig, ok := sig.(syscall.Signal); ok {
		return int(sysSig)
	}

	return 0
}

type RunnerConfig struct {
	ErrHandler func(error)
}

func (c *RunnerConfig) Option(opts ...RunnerOption) {
	for _, opt := range opts {
		opt.ConfigureRunner(c)
	}
}

type RunnerOption interface {
	ConfigureRunner(c *RunnerConfig)
}
