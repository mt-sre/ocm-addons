// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package installations

import (
	"strings"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("addon_id, addon_name, installed_version_id, cluster_id, cluster_name, cluster_state, state")

	return generateCommand(&opts, run(&opts))
}

type options struct {
	cli.CommonOptions
}

const longDescription = `List all installations of a given add-on by cluster in the current OCM environment.
If no argument is provided all installations of all addo-ons will be listed.`

func generateCommand(options *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "installations [ADDON_ID|ADDON_NAME|ADDON_NAME_SEARCH]",
		Short: "list all installations of a given add-on",
		Long:  longDescription,
		RunE:  run,
	}

	flags := cmd.Flags()

	options.AddColumnsFlag(flags)
	options.AddNoColorFlag(flags)
	options.AddNoHeadersFlag(flags)

	return cmd
}

func run(opts *options) func(cmd *cobra.Command, args []string) error { //nolint: cyclop
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		sess, err := cli.NewSession()
		if err != nil {
			return err
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

		requiresSub := hasSubscriptionField(opts.Columns)

		var pattern string

		if len(args) > 0 {
			pattern = args[0]
		}

		trace := sess.Logger().
			WithFields(log.Fields{
				"command": "installations",
				"search":  pattern,
			}).
			Trace("running command")
		defer trace.Stop(nil)

		clusters, err := ocm.RetrieveClusters(sess.Conn(), trace)
		if err != nil {
			return err
		}

		if err := clusters.ForEach(ctx, func(cluster *ocm.Cluster) error {
			cluster, err := cluster.WithAddonInstallations(ctx)
			if err != nil {
				return err
			}

			if requiresSub {
				cluster, err = cluster.WithSubscription(ctx)
				if err != nil {
					return err
				}
			}

			addons := cluster.AddonInstallations

			if pattern != "" {
				addons = addons.Matching(pattern)
			}

			for i := range addons {
				if err := table.Write(&addons[i]); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	}
}

func hasSubscriptionField(columns string) bool {
	subFields := new(ocm.Subscription).ProvideRowData()

	fields := make([]string, 0, len(subFields))

	for f := range subFields {
		fields = append(fields, "Cluster "+f)
	}

	requestedFields := strings.Split(columns, ",")

	for _, c := range requestedFields {
		for _, f := range fields {
			if cli.Normalize(c) != cli.Normalize(f) {
				continue
			}

			return true
		}
	}

	return false
}
