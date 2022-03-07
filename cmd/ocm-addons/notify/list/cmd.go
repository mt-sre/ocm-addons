package list

import (
	"fmt"
	"strings"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/notification"
	"github.com/mt-sre/ocm-addons/internal/output"

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

		tableOpts := []output.TableOption{
			output.WithColumns(opts.Columns),
		}

		if pager := sess.Config().Pager(); pager != "" {
			tableOpts = append(tableOpts, output.WithPager(pager))
		}

		table, err := output.NewTable(tableOpts...)
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

					if err := table.Write(&cfg, output.WithAdditionalFields(
						output.Field{
							Name:  "ID",
							Value: id,
						},
						output.Field{
							Name:  "Product",
							Value: prod,
						},
						output.Field{
							Name:  "Team",
							Value: strings.ToUpper(team),
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
