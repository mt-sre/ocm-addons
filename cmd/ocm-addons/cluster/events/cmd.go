// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"strings"

	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/ocm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Cmd() *cobra.Command {
	var opts options

	opts.DefaultColumns("timestamp, cluster_uuid, severity, summary")

	opts.SearchUsage("returns log events whose summary matches the given pattern")

	opts.OrderDefault("descending")
	opts.OrderUsage("selects whether logs are displayed in 'ascending' or 'descending' order by time")

	opts.BeforeUsage("returns log events which occurred before the specified time (YYYY-MM-DD HH:mm:ss)")

	opts.AfterUsage("returns log events which occurred after the specified time (YYYY-MM-DD HH:mm:ss)")

	return generateCommand(&opts, run(&opts))
}

type options struct {
	cli.CommonOptions
	cli.SearchOptions
	cli.FilterOptions
	Level   ocm.LogLevel
	levelIn string
}

func (o *options) AddLevelFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&o.levelIn,
		"level",
		o.levelIn,
		"Selects the level of the logs to display.",
	)
}

func (o *options) ParseOptions() {
	o.Level = parseLogLevel(o.levelIn)
}

func parseLogLevel(maybeLvl string) ocm.LogLevel {
	usTitler := cases.Title(language.AmericanEnglish)

	switch usTitler.String(strings.ToLower(maybeLvl)) {
	case ocm.LogLevelDebug:
		return ocm.LogLevelDebug
	case ocm.LogLevelInfo:
		return ocm.LogLevelInfo
	case ocm.LogLevelWarning:
		return ocm.LogLevelWarning
	case ocm.LogLevelError:
		return ocm.LogLevelError
	case ocm.LogLevelFatal:
		return ocm.LogLevelFatal
	default:
		return ocm.LogLevelNone
	}
}

const longDesc = `Retrieve add-on related cluster logs describing installs, uninstalls, removals,
and failures to install or remove.`

func generateCommand(opts *options, run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events [CLUSTER_ID|EXTERNAL_ID|CLUSTER_NAME|CLUSTER_NAME_SEARCH]",
		Short: "retrieve add-on related cluster logs",
		Long:  longDesc,
		Args:  cobra.MinimumNArgs(1),
		RunE:  run,
	}

	flags := cmd.Flags()

	opts.AddColumnsFlag(flags)
	opts.AddNoColorFlag(flags)
	opts.AddNoHeadersFlag(flags)
	opts.AddOrderFlag(flags)
	opts.AddLevelFlag(flags)
	opts.AddBeforeFlag(flags)
	opts.AddAfterFlag(flags)
	opts.AddSearchFlag(flags)

	return cmd
}

func run(opts *options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		opts.ParseOptions()

		sess, err := cli.NewSession()
		if err != nil {
			return err
		}

		defer sess.End()

		table, err := cli.NewTable(
			cli.WithColumns(opts.Columns),
			cli.WithNoHeaders(opts.NoHeaders),
			cli.WithNoColor(opts.NoColor),
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
				"command": "cluster events",
				"search":  search,
			}).
			Trace("running command")
		defer trace.Stop(nil)

		pager, err := ocm.RetrieveClusters(sess.Conn(), trace)
		if err != nil {
			return err
		}

		matchingClusters := pager.SearchByNameOrID(search)

		options, err := commandOptsToGetLogsOpts(opts)
		if err != nil {
			return err
		}

		return matchingClusters.ForEach(ctx, func(c *ocm.Cluster) error {
			logs, err := c.GetLogs(ctx, options)
			if err != nil {
				return err
			}

			for i := range logs {
				if err := table.Write(&logs[i]); err != nil {
					return err
				}
			}

			return nil
		})
	}
}

func commandOptsToGetLogsOpts(opts *options) (ocm.GetLogsOptions, error) {
	pattern := ""

	if opts.Search != "" {
		pattern += fmt.Sprintf("%%%s%%", opts.Search)
	}

	if err := opts.ParseFilterOptions(); err != nil {
		return ocm.GetLogsOptions{}, err
	}

	return ocm.NewGetLogsOptions(
		ocm.GetLogsMatchingPattern(pattern),
		ocm.GetLogsWithLevel(opts.Level),
		ocm.GetLogsSorted(ocm.LogEntryByTime(opts.Order)),
		ocm.GetLogsBefore(opts.Before),
		ocm.GetLogsAfter(opts.After),
	), nil
}
