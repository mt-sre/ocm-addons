/* #nosec */

package generator

import (
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type WithAddons []*cmv1.AddOnInstallationBuilder

func (a WithAddons) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.Addons = []*cmv1.AddOnInstallationBuilder(a)
}

type WithBaseDomain string

func (b WithBaseDomain) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.BaseDomain = string(b)
}

type WithCCS bool

func (ccs WithCCS) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.CCS = bool(ccs)
}

type WithCloudProvider string

func (cp WithCloudProvider) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.CloudProvider = string(cp)
}

type WithClusterID string

func (i WithClusterID) ConfigureSubscriptionGenerator(c *SubscriptionGenerateConfig) {
	c.ClusterID = string(i)
}

type WithDescription string

func (d WithDescription) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.Description = string(d)
}

type WithDocsLink string

func (d WithDocsLink) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.DocsLink = string(d)
}

type WithEnabled bool

func (e WithEnabled) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.Enabled = bool(e)
}

type WithExternalClusterID string

func (i WithExternalClusterID) ConfigureSubscriptionGenerator(c *SubscriptionGenerateConfig) {
	c.ExternalClusterID = string(i)
}

type WithExternalID string

func (i WithExternalID) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.ExternalID = string(i)
}

type WithHasExternalResources bool

func (e WithHasExternalResources) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.HasExternalResources = bool(e)
}

type WithHidden bool

func (h WithHidden) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.Hidden = bool(h)
}

type WithID string

func (i WithID) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.ID = string(i)
}

func (i WithID) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.ID = string(i)
}

func (i WithID) ConfigureSubscriptionGenerator(c *SubscriptionGenerateConfig) {
	c.ID = string(i)
}

type WithInstallMode string

func (i WithInstallMode) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.InstallMode = string(i)
}

type WithManaged bool

func (cm WithManaged) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.Managed = bool(cm)
}

type WithMultiAZ bool

func (m WithMultiAZ) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.MultiAZ = bool(m)
}

type WithName string

func (n WithName) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.Name = string(n)
}

func (n WithName) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.Name = string(n)
}

type WithOrganizationID string

func (i WithOrganizationID) ConfigureSubscriptionGenerator(c *SubscriptionGenerateConfig) {
	c.OrganizationID = string(i)
}

type WithProductID string

func (p WithProductID) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.ProductID = string(p)
}

type WithRegionID string

func (r WithRegionID) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.RegionID = string(r)
}

type WithResourceCost float64

func (r WithResourceCost) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.ResourceCost = float64(r)
}

type WithResourceName string

func (r WithResourceName) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.ResourceName = string(r)
}

type WithState string

func (s WithState) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.State = string(s)
}

type WithSubscriptionID string

func (i WithSubscriptionID) ConfigureClusterGenerator(c *ClusterGenerateConfig) {
	c.SubscriptionID = string(i)
}

type WithVersion string

func (v WithVersion) ConfigureAddOnGenerator(c *AddOnGenerateConfig) {
	c.Version = string(v)
}
