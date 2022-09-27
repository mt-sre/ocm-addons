// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/cluster/events"
	"github.com/mt-sre/ocm-addons/cmd/ocm-addons/cluster/info"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return generateCommand()
}

func generateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster [command]",
		Short: "retrieve cluster details",
		Long:  "Retrieves cluster details with additional information for installed add-ons.",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(info.Cmd())
	cmd.AddCommand(events.Cmd())

	return cmd
}
