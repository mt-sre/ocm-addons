/* #nosec */

package integration

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

const (
	pageSize = 50
)

func NewAddOnListPager(addons ...*cmv1.AddOn) *AddOnListJSONPager {
	var gen AddOnListJSONPager

	gen.items = addons
	gen.pageIndex = 1
	gen.pageSize = pageSize

	return &gen
}

type AddOnListJSONPager struct {
	pageSize  int
	pageIndex int
	items     []*cmv1.AddOn
}

func (p *AddOnListJSONPager) ToRoutes() ([]Route, error) {
	pages := make(map[int]string)

	for i := 0; i < p.Pages(); i++ {
		index := p.pageIndex

		page, err := p.NextPage()
		if err != nil {
			return nil, err
		}

		pages[index] = page
	}

	handler := func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		params := r.URL.Query()

		pageIndex, err := strconv.Atoi(params.Get("page"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		sdktesting.RespondWithJSON(http.StatusOK, pages[pageIndex])(w, r)
	}

	routes := []Route{
		{
			Method:  "GET",
			Path:    "/api/clusters_mgmt/v1/addons",
			Handler: handler,
		},
	}

	for _, addon := range p.items {
		var buf bytes.Buffer

		_ = cmv1.MarshalAddOnVersion(addon.Version(), &buf)

		routes = append(routes, Route{
			Method:  "GET",
			Path:    addon.Version().HREF(),
			Handler: sdktesting.RespondWithJSON(http.StatusOK, buf.String()),
		})
	}

	return routes, nil
}

func (p *AddOnListJSONPager) Pages() int {
	return int(math.Ceil(float64(len(p.items)) / float64(p.pageSize)))
}

func (p *AddOnListJSONPager) NextPage() (string, error) { //nolint
	type addonListJSON struct {
		Kind  string    `json:"kind"`
		Page  int       `json:"page"`
		Size  int       `json:"size"`
		Total int       `json:"total"`
		Items addOnList `json:"items"`
	}

	start := ((p.pageIndex - 1) * p.pageSize)
	end := int(math.Min(float64(p.pageIndex*p.pageSize), float64(len(p.items))))

	list := addonListJSON{
		Kind:  "AddOnList",
		Page:  p.pageIndex,
		Total: len(p.items),
	}

	list.Size = len(p.items[start:end])
	list.Items = addOnList(p.items[start:end])

	if start >= len(p.items) {
		list.Size = 0
		list.Items = addOnList([]*cmv1.AddOn{})
	}

	buf, err := json.Marshal(list)
	if err != nil {
		return "", err
	}

	p.pageIndex++

	return string(buf), nil
}

type addOnList []*cmv1.AddOn

func (l addOnList) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	err := cmv1.MarshalAddOnList([]*cmv1.AddOn(l), &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateAddOn(options ...AddOnGenerateOption) (*cmv1.AddOn, error) {
	var gen AddOnGenerator

	for _, opt := range options {
		if err := opt(&gen); err != nil {
			return nil, err
		}
	}

	if gen.description == "" {
		gen.description = "Test Addon"
	}

	if gen.docsLink == "" {
		gen.docsLink = "https://example.com"
	}

	if gen.id == "" {
		gen.id = "test-addon-0"
	}

	if gen.installMode == "" {
		gen.installMode = "single_namespace"
	}

	if gen.name == "" {
		gen.name = "Test Addon 0"
	}

	if gen.resourceName == "" {
		gen.resourceName = "FreeWithOSD"
	}

	if gen.version == "" {
		gen.version = "0.0.0"
	}

	const addonsAPIPrefix = "/api/clusters_mgmt/v1/addons"

	version := cmv1.NewAddOnVersion().
		ID(gen.version).
		HREF(slashJoin(addonsAPIPrefix, gen.id, "versions", gen.version))

	return cmv1.NewAddOn().
		HREF(slashJoin(addonsAPIPrefix, gen.id)).
		ID(gen.id).
		Version(version).
		Name(gen.name).
		Description(gen.description).
		DocsLink(gen.docsLink).
		Label(slashJoin("api.openshift.com", gen.id+"-addon")).
		Enabled(gen.enabled).
		ResourceName(gen.resourceName).
		ResourceCost(gen.resourceCost).
		TargetNamespace(gen.id).
		InstallMode(cmv1.AddOnInstallMode(gen.installMode)).
		OperatorName(gen.id).
		Hidden(gen.hidden).
		HasExternalResources(gen.hasExternalResources).
		Build()
}

type AddOnGenerator struct {
	description          string
	docsLink             string
	enabled              bool
	hasExternalResources bool
	hidden               bool
	id                   string
	installMode          string
	name                 string
	resourceCost         float64
	resourceName         string
	version              string
}

type AddOnGenerateOption func(g *AddOnGenerator) error

func AddOnDescription(desc string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.description = desc

		return nil
	}
}

func AddOnDocsLink(docsLink string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.docsLink = docsLink

		return nil
	}
}

func AddOnEnabled(g *AddOnGenerator) error {
	g.enabled = true

	return nil
}

func AddOnHasExternalResources(g *AddOnGenerator) error {
	g.hasExternalResources = true

	return nil
}

func AddOnHidden(g *AddOnGenerator) error {
	g.hidden = true

	return nil
}

func AddOnID(id string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.id = id

		return nil
	}
}

func AddOnInstallMode(installMode string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.installMode = installMode

		return nil
	}
}

func AddOnName(name string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.name = name

		return nil
	}
}

func AddOnResourceCost(resourceCost float64) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.resourceCost = resourceCost

		return nil
	}
}

func AddOnResourceName(resourceName string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.resourceName = resourceName

		return nil
	}
}

func AddOnVersion(version string) AddOnGenerateOption {
	return func(g *AddOnGenerator) error {
		g.version = version

		return nil
	}
}

func NewClusterListPager(clusters ...*cmv1.Cluster) *ClusterListJSONPager {
	var gen ClusterListJSONPager

	gen.pageIndex = 1
	gen.pageSize = pageSize
	gen.items = clusters

	return &gen
}

type ClusterListJSONPager struct {
	pageSize  int
	pageIndex int
	items     []*cmv1.Cluster
}

func (p *ClusterListJSONPager) ToRoutes() ([]Route, error) {
	pages := make(map[int]string)

	for i := 0; i < p.Pages(); i++ {
		index := p.pageIndex

		page, err := p.NextPage()
		if err != nil {
			return nil, err
		}

		pages[index] = page
	}

	handler := func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		params := r.URL.Query()

		pageIndex, err := strconv.Atoi(params.Get("page"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		sdktesting.RespondWithJSON(http.StatusOK, pages[pageIndex])(w, r)
	}

	var routes []Route

	routes = append(routes, Route{
		Method:  "GET",
		Path:    "/api/clusters_mgmt/v1/clusters",
		Handler: handler,
	})

	addons := make(map[*cmv1.AddOn]struct{})

	for _, cluster := range p.items {
		cRoutes, err := clusterRoute(cluster, addons)
		if err != nil {
			return nil, err
		}

		routes = append(routes, cRoutes...)
	}

	addonsList := make([]*cmv1.AddOn, 0, len(addons))

	for addon := range addons {
		addonsList = append(addonsList, addon)
	}

	addonRoutes, err := NewAddOnListPager(addonsList...).ToRoutes()
	if err != nil {
		return nil, err
	}

	routes = append(routes, addonRoutes...)

	return routes, nil
}

func clusterRoute(cluster *cmv1.Cluster, addons map[*cmv1.AddOn]struct{}) ([]Route, error) {
	var buf bytes.Buffer

	sub, err := NewSubscription(
		SubscriptionID(cluster.Subscription().ID()),
		SubscriptionClusterID(cluster.ID()),
		SubscriptionExternalClusterID(cluster.ExternalID()),
	)
	if err != nil {
		return nil, err
	}

	err = amv1.MarshalSubscription(sub, &buf)
	if err != nil {
		return nil, err
	}

	subRoute := Route{
		Method:  "GET",
		Path:    slashJoin("/api/accounts_mgmt/v1/subscriptions", cluster.Subscription().ID()),
		Handler: sdktesting.RespondWithJSON(http.StatusOK, buf.String()),
	}

	buf.Reset()

	for _, addon := range cluster.Addons().Slice() {
		addons[addon.Addon()] = struct{}{}
	}

	err = cmv1.MarshalAddOnInstallationList(cluster.Addons().Slice(), &buf)
	if err != nil {
		return nil, err
	}

	addonsRoute := Route{
		Method:  "GET",
		Path:    slashJoin("/api/clusters_mgmt/v1/clusters", cluster.ID(), "addons"),
		Handler: sdktesting.RespondWithJSON(http.StatusOK, buf.String()),
	}

	return []Route{
		subRoute,
		addonsRoute,
	}, nil
}

func (p *ClusterListJSONPager) Pages() int {
	return int(math.Ceil(float64(len(p.items)) / float64(p.pageSize)))
}

func (p *ClusterListJSONPager) NextPage() (string, error) { //nolint
	type clusterListJSON struct {
		Kind  string      `json:"kind"`
		Page  int         `json:"page"`
		Size  int         `json:"size"`
		Total int         `json:"total"`
		Items clusterList `json:"items"`
	}

	start := ((p.pageIndex - 1) * p.pageSize)
	end := int(math.Min(float64(p.pageIndex*p.pageSize), float64(len(p.items))))

	list := clusterListJSON{
		Kind:  "ClusterList",
		Page:  p.pageIndex,
		Total: len(p.items),
	}

	list.Size = len(p.items[start:end])
	list.Items = clusterList(p.items[start:end])

	if start >= len(p.items) {
		list.Size = 0
		list.Items = clusterList([]*cmv1.Cluster{})
	}

	buf, err := json.Marshal(list)
	if err != nil {
		return "", err
	}

	p.pageIndex++

	return string(buf), nil
}

type clusterList []*cmv1.Cluster

func (l clusterList) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	err := cmv1.MarshalClusterList([]*cmv1.Cluster(l), &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateCluster(options ...ClusterGenerateOption) (*cmv1.Cluster, error) { //nolint:cyclop
	var gen ClusterGenerator

	for _, opt := range options {
		if err := opt(&gen); err != nil {
			return nil, err
		}
	}

	if gen.id == "" {
		gen.id = "0123456789abcdefghijklmnopqrstuv"
	}

	if gen.externalID == "" {
		gen.externalID = "00000000-0000-0000-0000-000000000000"
	}

	if gen.name == "" {
		gen.name = "test-cluster-0"
	}

	if gen.openshiftVersion == "" {
		gen.openshiftVersion = "4.9.11"
	}

	if gen.cloudProvider == "" {
		gen.cloudProvider = "aws"
	}

	if gen.regionID == "" && gen.cloudProvider == "aws" {
		gen.regionID = "us-east-1"
	}

	if gen.subscriptionID == "" {
		gen.subscriptionID = "0123456789abcdefghijklmnopq"
	}

	if gen.baseDomain == "" {
		gen.baseDomain = "xxxx.s1.devshift.org"
	}

	if gen.productID == "" {
		gen.productID = "osd"
	}

	if gen.state == "" {
		gen.state = "ready"
	}

	subscription := cmv1.NewSubscription().
		ID(gen.subscriptionID).
		HREF(slashJoin("/api/clusters_mgmt/v1/subscriptions", gen.subscriptionID))

	cloudProviderHREF := slashJoin("/api/clusters_mgmt/v1/cloud_providers", gen.cloudProvider)
	cloudProvider := cmv1.NewCloudProvider().
		ID(gen.cloudProvider).
		HREF(cloudProviderHREF)

	region := cmv1.NewCloudRegion().
		ID(gen.regionID).
		HREF(slashJoin(cloudProviderHREF, gen.regionID))

	dns := cmv1.NewDNS().
		BaseDomain(gen.baseDomain)

	console := cmv1.NewClusterConsole().
		URL(dotJoin("https://console-openshift-console.apps", gen.name, gen.baseDomain))

	api := cmv1.NewClusterAPI().
		URL(dotJoin("https://api", gen.name, gen.baseDomain+":6443"))

	ccs := cmv1.NewCCS().
		Enabled(gen.ccs)

	longVersion := "openshift-v" + gen.openshiftVersion
	version := cmv1.NewVersion().
		RawID(gen.openshiftVersion).
		ID(longVersion).
		HREF(slashJoin("/api/clusters_mgmt/v1/versions", longVersion)).
		ChannelGroup("stable")

	addons := cmv1.NewAddOnInstallationList().
		Items(gen.addons...)

	product := cmv1.NewProduct().
		ID(gen.productID).
		HREF(slashJoin("/api/clusters_mgmt/v1/products", gen.productID))

	return cmv1.NewCluster().
		ID(gen.id).
		HREF(slashJoin("/api/clusters_mgmt/v1/clusters", gen.id)).
		Name(gen.name).
		ExternalID(gen.externalID).
		OpenshiftVersion(gen.openshiftVersion).
		CloudProvider(cloudProvider).
		Region(region).
		Subscription(subscription).
		DNS(dns).
		Console(console).
		API(api).
		State(cmv1.ClusterState(gen.state)).
		Managed(gen.managed).
		MultiAZ(gen.multiAZ).
		CCS(ccs).
		Version(version).
		Addons(addons).
		Product(product).
		Build()
}

type ClusterGenerator struct {
	addons           []*cmv1.AddOnInstallationBuilder
	baseDomain       string
	ccs              bool
	cloudProvider    string
	externalID       string
	id               string
	managed          bool
	multiAZ          bool
	name             string
	openshiftVersion string
	productID        string
	regionID         string
	state            string
	subscriptionID   string
}

type ClusterGenerateOption func(g *ClusterGenerator) error

func ClusterAddons(addons []*cmv1.AddOnInstallationBuilder) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.addons = addons

		return nil
	}
}

func ClusterBaseDomain(baseDomain string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.baseDomain = baseDomain

		return nil
	}
}

func ClusterCCS(g *ClusterGenerator) error {
	g.ccs = true

	return nil
}

func ClusterCloudProvider(cloudProvider string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.cloudProvider = cloudProvider

		return nil
	}
}

func ClusterExternalID(externalID string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.externalID = externalID

		return nil
	}
}

func ClusterID(id string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.id = id

		return nil
	}
}

func ClusterManaged(g *ClusterGenerator) error {
	g.managed = true

	return nil
}

func ClusterMultiAZ(g *ClusterGenerator) error {
	g.multiAZ = true

	return nil
}

func ClusterName(name string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.name = name

		return nil
	}
}

func ClusterOpenshiftVersion(openshiftVersion string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.openshiftVersion = openshiftVersion

		return nil
	}
}

func ClusterProductID(productID string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.productID = productID

		return nil
	}
}

func ClusterRegionID(regionID string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.regionID = regionID

		return nil
	}
}

func ClusterState(state string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.state = state

		return nil
	}
}

func ClusterSubscriptionID(subscriptionID string) ClusterGenerateOption {
	return func(g *ClusterGenerator) error {
		g.subscriptionID = subscriptionID

		return nil
	}
}

func NewSubscription(opts ...SubscriptionGenerateOption) (*amv1.Subscription, error) {
	var gen SubscriptionGenerator

	for _, opt := range opts {
		if err := opt(&gen); err != nil {
			return nil, err
		}
	}

	if gen.id == "" {
		gen.id = "0123456789abcdefghijklmnopqrstu"
	}

	if gen.organizationID == "" {
		gen.organizationID = "0123456789abcdefghijklmnopq"
	}

	return amv1.NewSubscription().
		ID(gen.id).
		OrganizationID(gen.organizationID).
		ClusterID(gen.clusterID).
		ExternalClusterID(gen.externalClusterID).
		Build()
}

type SubscriptionGenerator struct {
	id                string
	organizationID    string
	clusterID         string
	externalClusterID string
}

type SubscriptionGenerateOption func(g *SubscriptionGenerator) error

func SubscriptionID(id string) SubscriptionGenerateOption {
	return func(g *SubscriptionGenerator) error {
		g.id = id

		return nil
	}
}

func SubscriptionOrganizationID(orgID string) SubscriptionGenerateOption {
	return func(g *SubscriptionGenerator) error {
		g.organizationID = orgID

		return nil
	}
}

func SubscriptionClusterID(clusterID string) SubscriptionGenerateOption {
	return func(g *SubscriptionGenerator) error {
		g.clusterID = clusterID

		return nil
	}
}

func SubscriptionExternalClusterID(externalClusterID string) SubscriptionGenerateOption {
	return func(g *SubscriptionGenerator) error {
		g.externalClusterID = externalClusterID

		return nil
	}
}

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
