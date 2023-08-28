// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("id, external_id, name, organization_id, product_id, installed_addons, dns_base_domain")

	return generateCommand(&opts, run(&opts))
}

type options struct {
	cli.CommonOptions
}

func generateCommand(opts *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [CLUSTER_ID|EXTERNAL_ID|CLUSTER_NAME|CLUSTER_NAME_SEARCH]",
		Short: "retrieve cluster information",
		Long:  "Retrieve cluster information including summary data related to add-ons.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  run,
	}

	flags := cmd.Flags()

	opts.AddColumnsFlag(flags)
	opts.AddNoColorFlag(flags)
	opts.AddNoHeadersFlag(flags)

	return cmd
}

func run(opts *options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		sess, err := cli.NewSession()
		if err != nil {
			return fmt.Errorf("starting new session: %w", err)
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
			return err
		}

		defer table.Flush()

		search := args[0]

		trace := sess.Logger().
			WithFields(log.Fields{
				"command": "cluster info",
				"search":  search,
			}).
			Trace("running command")
		defer trace.Stop(nil)

		clusters, err := ocm.RetrieveClusters(sess.Conn(), trace)
		if err != nil {
			return err
		}

		matchingClusters := clusters.SearchByNameOrID(search)

		err = matchingClusters.ForEach(ctx, func(cluster *ocm.Cluster) error {
			cluster, err := cluster.WithSubscription(ctx)
			if err != nil {
				return err
			}

			cluster, err = cluster.WithAddonInstallations(ctx)
			if err != nil {
				return err
			}

			if err := table.Write(cluster); err != nil {
				return fmt.Errorf("writing cluster to table: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	}
}
