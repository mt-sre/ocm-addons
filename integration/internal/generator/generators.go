/* #nosec */

package generator

import (
	"net/http"

	"github.com/mt-sre/ocm-addons/integration/internal/utils"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GenerateAddOn(opts ...AddOnGenerateOption) (*cmv1.AddOn, error) {
	var cfg AddOnGenerateConfig

	cfg.Option(opts...)
	cfg.Default()

	const addonsAPIPrefix = "/api/clusters_mgmt/v1/addons"

	version := cmv1.NewAddOnVersion().
		ID(cfg.Version).
		HREF(utils.SlashJoin(addonsAPIPrefix, cfg.ID, "versions", cfg.Version))

	return cmv1.NewAddOn().
		HREF(utils.SlashJoin(addonsAPIPrefix, cfg.ID)).
		ID(cfg.ID).
		Version(version).
		Name(cfg.Name).
		Description(cfg.Description).
		DocsLink(cfg.DocsLink).
		Label(utils.SlashJoin("api.openshift.com", cfg.ID+"-addon")).
		Enabled(cfg.Enabled).
		ResourceName(cfg.ResourceName).
		ResourceCost(cfg.ResourceCost).
		TargetNamespace(cfg.ID).
		InstallMode(cmv1.AddOnInstallMode(cfg.InstallMode)).
		OperatorName(cfg.ID).
		Hidden(cfg.Hidden).
		HasExternalResources(cfg.HasExternalResources).
		Build()
}

type AddOnGenerateConfig struct {
	Description          string
	DocsLink             string
	Enabled              bool
	HasExternalResources bool
	Hidden               bool
	ID                   string
	InstallMode          string
	Name                 string
	ResourceCost         float64
	ResourceName         string
	Version              string
}

func (c *AddOnGenerateConfig) Option(opts ...AddOnGenerateOption) {
	for _, opt := range opts {
		opt.ConfigureAddOnGenerator(c)
	}
}

func (c *AddOnGenerateConfig) Default() {
	if c.Description == "" {
		c.Description = "Test Addon"
	}

	if c.DocsLink == "" {
		c.DocsLink = "https://example.com"
	}

	if c.ID == "" {
		c.ID = "test-addon-0"
	}

	if c.InstallMode == "" {
		c.InstallMode = "single_namespace"
	}

	if c.Name == "" {
		c.Name = "Test Addon 0"
	}

	if c.ResourceName == "" {
		c.ResourceName = "FreeWithOSD"
	}

	if c.Version == "" {
		c.Version = "0.0.0"
	}
}

type AddOnGenerateOption interface {
	ConfigureAddOnGenerator(*AddOnGenerateConfig)
}

func GenerateCluster(opts ...ClusterGenerateOption) (*cmv1.Cluster, error) {
	var cfg ClusterGenerateConfig

	cfg.Option(opts...)
	cfg.Default()

	subscription := cmv1.NewSubscription().
		ID(cfg.SubscriptionID).
		HREF(utils.SlashJoin("/api/clusters_mgmt/v1/subscriptions", cfg.SubscriptionID))

	cloudProviderHREF := utils.SlashJoin("/api/clusters_mgmt/v1/cloud_providers", cfg.CloudProvider)
	cloudProvider := cmv1.NewCloudProvider().
		ID(cfg.CloudProvider).
		HREF(cloudProviderHREF)

	region := cmv1.NewCloudRegion().
		ID(cfg.RegionID).
		HREF(utils.SlashJoin(cloudProviderHREF, cfg.RegionID))

	dns := cmv1.NewDNS().
		BaseDomain(cfg.BaseDomain)

	console := cmv1.NewClusterConsole().
		URL(utils.DotJoin("https://console-openshift-console.apps", cfg.Name, cfg.BaseDomain))

	api := cmv1.NewClusterAPI().
		URL(utils.DotJoin("https://api", cfg.Name, cfg.BaseDomain+":6443"))

	ccs := cmv1.NewCCS().
		Enabled(cfg.CCS)

	longVersion := "openshift-v" + cfg.OpenshiftVersion
	version := cmv1.NewVersion().
		RawID(cfg.OpenshiftVersion).
		ID(longVersion).
		HREF(utils.SlashJoin("/api/clusters_mgmt/v1/versions", longVersion)).
		ChannelGroup("stable")

	addons := cmv1.NewAddOnInstallationList().
		Items(cfg.Addons...)

	product := cmv1.NewProduct().
		ID(cfg.ProductID).
		HREF(utils.SlashJoin("/api/clusters_mgmt/v1/products", cfg.ProductID))

	return cmv1.NewCluster().
		ID(cfg.ID).
		HREF(utils.SlashJoin("/api/clusters_mgmt/v1/clusters", cfg.ID)).
		Name(cfg.Name).
		DisplayName(cfg.Name).
		ExternalID(cfg.ExternalID).
		OpenshiftVersion(cfg.OpenshiftVersion).
		CloudProvider(cloudProvider).
		Region(region).
		Subscription(subscription).
		DNS(dns).
		Console(console).
		API(api).
		State(cmv1.ClusterState(cfg.State)).
		Managed(cfg.Managed).
		MultiAZ(cfg.MultiAZ).
		CCS(ccs).
		Version(version).
		Addons(addons).
		Product(product).
		Build()
}

type ClusterGenerateConfig struct {
	Addons           []*cmv1.AddOnInstallationBuilder
	BaseDomain       string
	CCS              bool
	CloudProvider    string
	ExternalID       string
	ID               string
	Managed          bool
	MultiAZ          bool
	Name             string
	OpenshiftVersion string
	ProductID        string
	RegionID         string
	State            string
	SubscriptionID   string
}

func (c *ClusterGenerateConfig) Option(opts ...ClusterGenerateOption) {
	for _, opt := range opts {
		opt.ConfigureClusterGenerator(c)
	}
}

func (c *ClusterGenerateConfig) Default() { //nolint: cyclop
	if c.ID == "" {
		c.ID = "0123456789abcdefghijklmnopqrstuv"
	}

	if c.ExternalID == "" {
		c.ExternalID = "00000000-0000-0000-0000-000000000000"
	}

	if c.Name == "" {
		c.Name = "test-cluster-0"
	}

	if c.OpenshiftVersion == "" {
		c.OpenshiftVersion = "4.9.11"
	}

	if c.CloudProvider == "" {
		c.CloudProvider = "aws"
	}

	if c.RegionID == "" && c.CloudProvider == "aws" {
		c.RegionID = "us-east-1"
	}

	if c.SubscriptionID == "" {
		c.SubscriptionID = "0123456789abcdefghijklmnopq"
	}

	if c.BaseDomain == "" {
		c.BaseDomain = "xxxx.s1.devshift.org"
	}

	if c.ProductID == "" {
		c.ProductID = "osd"
	}

	if c.State == "" {
		c.State = "ready"
	}
}

type ClusterGenerateOption interface {
	ConfigureClusterGenerator(*ClusterGenerateConfig)
}

func NewSubscription(opts ...SubscriptionGenerateOption) (*amv1.Subscription, error) {
	var cfg SubscriptionGenerateConfig

	cfg.Option(opts...)
	cfg.Default()

	return amv1.NewSubscription().
		ID(cfg.ID).
		OrganizationID(cfg.OrganizationID).
		ClusterID(cfg.ClusterID).
		ExternalClusterID(cfg.ExternalClusterID).
		Build()
}

type SubscriptionGenerateConfig struct {
	ID                string
	OrganizationID    string
	ClusterID         string
	ExternalClusterID string
}

func (c *SubscriptionGenerateConfig) Option(opts ...SubscriptionGenerateOption) {
	for _, opt := range opts {
		opt.ConfigureSubscriptionGenerator(c)
	}
}

func (c *SubscriptionGenerateConfig) Default() {
	if c.ID == "" {
		c.ID = "0123456789abcdefghijklmnopqrstu"
	}

	if c.OrganizationID == "" {
		c.OrganizationID = "0123456789abcdefghijklmnopq"
	}
}

type SubscriptionGenerateOption interface {
	ConfigureSubscriptionGenerator(*SubscriptionGenerateConfig)
}

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
