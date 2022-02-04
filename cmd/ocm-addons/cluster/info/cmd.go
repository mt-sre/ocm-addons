package info

import (
	"context"
	"os"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("id, external_id, name, organization, product_id, installed_addons, domain")

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
	opts.AddNoHeadersFlag(flags)

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

		table, err := cli.NewTable(
			ctx,
			sess,
			cli.TableWriter(os.Stdout),
			cli.TableName("clusters"),
			cli.TableColumns(opts.Columns),
			cli.TableNoHeaders(opts.NoHeaders),
		)
		if err != nil {
			return err
		}

		defer table.Close()

		if err = table.WriteHeaders(); err != nil {
			return err
		}

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

			if err := table.WriteObject(cluster); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	}
}
