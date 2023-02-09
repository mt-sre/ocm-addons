// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package notify

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/notify/list"
	"github.com/mt-sre/ocm-addons/internal/cli"
	"github.com/mt-sre/ocm-addons/internal/notification"
	"github.com/mt-sre/ocm-addons/internal/ocm"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return generateCommand(run)
}

const _numArgs = 2

const _example = `
# Sending a notification with the following details:
# Team:                   "example-team"
# Product:                "example-product"
# Notification Config ID: "example-notification"
# Cluster                 "example-cluster"
  ocm addons notify example-cluster example-team/example-product/example-notification
`

func generateCommand(run func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"notification", "notifications"},
		Use:     "notify [CLUSTER_ID|EXTERNAL_ID|CLUSTER_NAME|CLUSTER_NAME_SEARCH] NOTIFICATION_ID",
		Example: _example,
		Short:   "post customer notifications",
		Long:    "Post add-on related notification to cluster service_logs for customer to view.",
		Args:    cobra.MinimumNArgs(_numArgs),
		RunE:    run,
	}

	cmd.AddCommand(list.Cmd())

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	var (
		ctx = cmd.Context()
		in  = cmd.InOrStdin()
		out = cmd.OutOrStdout()
	)

	sess, err := cli.NewSession()
	if err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	defer sess.End()

	search := args[0]
	notificationID := args[1]

	trace := sess.Logger().
		WithFields(log.Fields{
			"command":        "notify",
			"search":         search,
			"notificationID": notificationID,
		}).
		Trace("running command")
	defer trace.Stop(nil)

	cfg, err := getNotificationConfig(notificationID)
	if err != nil {
		return fmt.Errorf("getting notification %q: %w", notificationID, err)
	}

	entryOpts := []ocm.LogEntryOption{
		ocm.LogEntryDescription(cfg.Description),
		ocm.LogEntryServiceName(cfg.ServiceName),
		ocm.LogEntrySeverity(cfg.Severity),
		ocm.LogEntrySummary(cfg.Summary),
	}

	if cfg.InternalOnly {
		entryOpts = append(entryOpts, ocm.LogEntryInternalOnly)
	}

	pager, err := ocm.RetrieveClusters(sess.Conn(), trace)
	if err != nil {
		return fmt.Errorf("retrieving clusters: %w", err)
	}

	matchingClusters := pager.SearchByNameOrID(search)

	return matchingClusters.ForEach(ctx, func(c *ocm.Cluster) error {
		fmt.Fprintf(out, "Cluster External ID: %s\n", c.ExternalID())
		fmt.Fprintf(out, "Description: %s\n", cfg.Description)
		fmt.Fprintf(out, "Service Name: %s\n", cfg.ServiceName)
		fmt.Fprintf(out, "Severity: %s\n", cfg.Severity)
		fmt.Fprintf(out, "Summary: %s\n", cfg.Summary)
		fmt.Fprintf(out, "Internal Only: %s\n", fmt.Sprint(cfg.InternalOnly))

		if !cli.PromptYesOrNo(out, in, "Please confirm before sending this notification") {
			fmt.Fprintln(out, "notification cancelled")

			return nil
		}

		fmt.Fprintln(out, "sending notification...")

		if err := c.PostLog(ctx, entryOpts...); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		fmt.Fprintln(out, "notification sent successfully")

		return nil
	})
}

var (
	errInvalidNotificationID = errors.New("invalid notificstion ID")
	errNotificationNotFound  = errors.New("notification not found")
)

func getNotificationConfig(rawID string) (notification.Config, error) {
	const numParts = 3

	parsed := strings.SplitN(rawID, "/", numParts)

	if len(parsed) < numParts {
		return notification.Config{}, errInvalidNotificationID
	}

	team, product, id := parsed[0], parsed[1], parsed[2]

	cfg, ok := notification.GetNotification(team, product, id)
	if !ok {
		return notification.Config{}, errNotificationNotFound
	}

	return cfg, nil
}
