// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/apex/log"
	apexcli "github.com/apex/log/handlers/cli"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/cluster"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/installations"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/list"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/notify"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/update"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/version"
	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/cli/signals"
	"github.com/spf13/cobra"
)

var verbosity int

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals.ShutdownSignals()...)

	var code int

	defer func() {
		stop()

		os.Exit(code)
	}()

	rootCmd := generateRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.
			WithError(err).
			Error("ocm addons exited unexpectedly")

		code = 1
	}
}

func generateRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "ocm addons [command]",
		Short:         "addon plug-in for the ocm-cli",
		Long:          "This plug-in extends the ocm-cli to provide additional commands for working with add-ons.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.AddCommand(cluster.Cmd())
	rootCmd.AddCommand(installations.Cmd())
	rootCmd.AddCommand(list.Cmd())
	rootCmd.AddCommand(notify.Cmd())
	rootCmd.AddCommand(update.Cmd())
	rootCmd.AddCommand(version.Cmd())

	flags := rootCmd.PersistentFlags()

	flags.CountVarP(
		&verbosity,
		"verbose",
		"v",
		"increase logging verbosity; '-vvv' for max verbosity",
	)

	cobra.OnInitialize(initLog)

	return rootCmd
}

func initLog() {
	log.SetHandler(apexcli.Default)
	log.SetLevel(cli.LogLevel(verbosity))
}
