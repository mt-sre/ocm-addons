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
		Use:   "cluster [CLUSTER_ID|EXTERNAL_ID|CLUSTER_NAME|CLUSTER_NAME_SEARCH]",
		Short: "retrieve cluster details",
		Long:  "Retrieves cluster details with additional information for installed add-ons.",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(info.Cmd())
	cmd.AddCommand(events.Cmd())

	return cmd
}
