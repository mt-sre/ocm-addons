package list

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
	opts.AddNoHeadersFlag(flags)
	opts.AddSearchFlag(flags)

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
			cli.TableName("addons"),
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

		trace := sess.Logger().
			WithFields(log.Fields{
				"command": "list",
			}).
			Trace("running command")
		defer trace.Stop(nil)

		addons, err := ocm.RetrieveAddons(sess.Conn(), trace)
		if err != nil {
			return err
		}

		matchingAddons := addons.SearchByNameOrID(opts.Search)

		err = matchingAddons.ForEach(ctx, func(a *ocm.Addon) error {
			if err := table.WriteObject(a); err != nil {
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
