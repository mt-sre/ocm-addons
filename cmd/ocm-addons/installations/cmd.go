package installations

import (
	"context"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/mt-sre/ocm-addons/internal/output"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("addon_id, addon_name, addon_version_id, cluster_id, cluster_name, cluster_state, state")

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

func run(opts *options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		sess, err := cli.NewSession()
		if err != nil {
			return err
		}

		defer sess.End()

		table, err := output.NewTable(
			output.WithColumns(opts.Columns),
			output.WithNoColor(opts.NoColor),
			output.WithNoHeaders(opts.NoHeaders),
			output.WithPager(sess.Pager()),
		)
		if err != nil {
			return err
		}

		defer table.Flush()

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
