// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("id, name, enabled")
	opts.SearchUsage("only return add-ons whose name or id matches the given pattern")

	return generateCommand(&opts, run(&opts))
}

type options struct {
	cli.CommonOptions
	cli.SearchOptions
}

func generateCommand(opts *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all add-ons",
		Long:  "List all add-ons known to the current OCM environment including those which are disabled.",
		Args:  cobra.NoArgs,
		RunE:  run,
	}

	flags := cmd.Flags()

	opts.AddColumnsFlag(flags)
	opts.AddNoColorFlag(flags)
	opts.AddNoHeadersFlag(flags)
	opts.AddSearchFlag(flags)

	return cmd
}

func run(opts *options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

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
			return fmt.Errorf("initializing table: %w", err)
		}

		defer table.Flush()

		trace := sess.Logger().
			WithFields(log.Fields{
				"command": "list",
			}).
			Trace("running command")
		defer trace.Stop(nil)

		addons, err := ocm.RetrieveAddons(sess.Conn(), trace)
		if err != nil {
			return fmt.Errorf("retrieving addons: %w", err)
		}

		matchingAddons := addons.SearchByNameOrID(opts.Search)

		err = matchingAddons.ForEach(ctx, func(a *ocm.Addon) error {
			addon, err := a.WithVersion(ctx)
			if err != nil {
				return fmt.Errorf("retrieving addon version: %w", err)
			}

			if err := table.Write(addon); err != nil {
				return fmt.Errorf("writing table row: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("populating table: %w", err)
		}

		return nil
	}
}
