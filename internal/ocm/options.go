package ocm

import (
	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

type WithConnection struct{ *sdk.Connection }

func (wc WithConnection) ConfigureAddon(c *AddonConfig) {
	c.Conn = wc.Connection
}

func (wc WithConnection) ConfigureCluster(c *ClusterConfig) {
	c.Conn = wc.Connection
}

type WithLogger struct {
	Logger log.Interface
}

func (wl WithLogger) ConfigureAddon(c *AddonConfig) {
	c.Logger = wl.Logger
}

func (wl WithLogger) ConfigureCluster(c *ClusterConfig) {
	c.Logger = wl.Logger
}
