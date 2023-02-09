// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"strings"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/notification"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("team, product, id, severity, summary")

	return generateCommand(&opts, run(&opts))
}

type options struct {
	cli.CommonOptions
}

func generateCommand(opts *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list available notifications",
		Long:    "List available notifications",
		Args:    cobra.NoArgs,
		RunE:    run,
	}

	flags := cmd.Flags()

	opts.AddNoHeadersFlag(flags)
	opts.AddNoColorFlag(flags)
	opts.AddColumnsFlag(flags)

	return cmd
}

func run(opts *options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sess, err := cli.NewSession()
		if err != nil {
			return fmt.Errorf("starting session: %w", err)
		}

		defer sess.End()

		table, err := cli.NewTable(
			cli.WithColumns(opts.Columns),
			cli.WithNoColor(opts.NoColor),
			cli.WithNoHeaders(opts.NoHeaders),
			cli.WithPager(sess.Pager()),
			cli.WithOutput{Out: cmd.OutOrStdout()},
		)
		if err != nil {
			return fmt.Errorf("creating table: %w", err)
		}

		defer table.Flush()

		trace := sess.Logger().
			WithFields(log.Fields{
				"command": "notify list",
			}).
			Trace("running command")
		defer trace.Stop(nil)

		for team, teamTree := range notification.GetAllNotifications() {
			for prod, configs := range teamTree {
				for id, cfg := range configs {
					cfg := cfg

					if err := table.Write(&cfg, cli.WithAdditionalFields(
						map[string]interface{}{
							"ID":      id,
							"Product": prod,
							"Team":    strings.ToUpper(team),
						},
					)); err != nil {
						return fmt.Errorf("writing table row: %w", err)
					}
				}
			}
		}

		return nil
	}
}
