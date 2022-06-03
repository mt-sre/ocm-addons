/* #nosec */

package generator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/mt-sre/ocm-addons/integration/internal/utils"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

func NewAddOnListEncoder(addons ...*cmv1.AddOn) *AddOnListEncoder {
	return &AddOnListEncoder{
		Kind:  "AddOnList",
		Page:  1,
		Items: addons,
		Size:  len(addons),
		Total: len(addons),
	}
}

type AddOnListEncoder struct {
	Kind  string    `json:"kind"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
	Total int       `json:"total"`
	Items addOnList `json:"items"`
}

func (p *AddOnListEncoder) ToRoutes() ([]Route, error) {
	buf, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	handler := func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		sdktesting.RespondWithJSON(http.StatusOK, string(buf))(w, r)
	}

	routes := []Route{
		{
			Method:  "GET",
			Path:    "/api/clusters_mgmt/v1/addons",
			Handler: handler,
		},
	}

	for _, addon := range p.Items {
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

type addOnList []*cmv1.AddOn

func (l addOnList) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	err := cmv1.MarshalAddOnList([]*cmv1.AddOn(l), &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewClusterListJSONEncoder(clusters ...*cmv1.Cluster) *ClusterListJSONEncoder {
	return &ClusterListJSONEncoder{
		Kind:  "ClusterList",
		Items: clusters,
		Page:  1,
		Size:  len(clusters),
		Total: len(clusters),
	}
}

type ClusterListJSONEncoder struct {
	Kind  string      `json:"kind"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int         `json:"total"`
	Items clusterList `json:"items"`
}

func (p *ClusterListJSONEncoder) ToRoutes() ([]Route, error) {
	buf, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	handler := func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		sdktesting.RespondWithJSON(http.StatusOK, string(buf))(w, r)
	}

	var routes []Route

	routes = append(routes, Route{
		Method:  "GET",
		Path:    regexp.MustCompile(`/api/clusters_mgmt/v1/clusters.*`),
		Handler: handler,
	})

	addons := make(map[*cmv1.AddOn]struct{})

	for _, cluster := range p.Items {
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

	addonRoutes, err := NewAddOnListEncoder(addonsList...).ToRoutes()
	if err != nil {
		return nil, err
	}

	routes = append(routes, addonRoutes...)

	return routes, nil
}

func clusterRoute(cluster *cmv1.Cluster, addons map[*cmv1.AddOn]struct{}) ([]Route, error) {
	var buf bytes.Buffer

	sub, err := NewSubscription(
		WithID(cluster.Subscription().ID()),
		WithClusterID(cluster.ID()),
		WithExternalClusterID(cluster.ExternalID()),
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
		Path:    utils.SlashJoin("/api/accounts_mgmt/v1/subscriptions", cluster.Subscription().ID()),
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
		Path:    utils.SlashJoin("/api/clusters_mgmt/v1/clusters", cluster.ID(), "addons"),
		Handler: sdktesting.RespondWithJSON(http.StatusOK, buf.String()),
	}

	return []Route{
		subRoute,
		addonsRoute,
	}, nil
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
