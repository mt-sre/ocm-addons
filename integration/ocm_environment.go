package integration

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/ghttp"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	sdktesting "github.com/openshift-online/ocm-sdk-go/testing"
)

const (
	tokenDuration = 15 * time.Minute
)

func NewOCMEnvironment(opts ...OCMEnvironmentOption) (*OCMEnvironment, error) {
	var (
		env OCMEnvironment
		err error
	)

	env.tmpDir, err = os.MkdirTemp("", "ocm-test-environment-*.d")
	if err != nil {
		return nil, err
	}

	env.apiServer = sdktesting.MakeTCPServer()

	for _, opt := range opts {
		err = env.Option(opt)
		if err != nil {
			return nil, err
		}
	}

	if env.ssoServer == nil {
		env.ssoServer = sdktesting.MakeTCPServer()

		token := sdktesting.MakeTokenString("Bearer", tokenDuration)

		env.ssoServer.AppendHandlers(
			sdktesting.RespondWithAccessToken(token),
		)
	}

	return &env, nil
}

type OCMEnvironment struct {
	apiServer     *ghttp.Server
	ssoServer     *ghttp.Server
	addons        []*cmv1.AddOn
	clusters      []*cmv1.Cluster
	subscriptions []*amv1.Subscription
	tmpDir        string
}

func (e *OCMEnvironment) Option(opt OCMEnvironmentOption) error {
	return opt(e)
}

func (e *OCMEnvironment) Addons() []*cmv1.AddOn {
	return append(make([]*cmv1.AddOn, 0, len(e.addons)), e.addons...)
}

func (e *OCMEnvironment) AddAddonRoutes() error {
	addonRoutes, err := NewAddOnListPager(e.addons...).ToRoutes()
	if err != nil {
		return err
	}

	for _, r := range addonRoutes {
		e.apiServer.RouteToHandler(r.Method, r.Path, r.Handler)
	}

	return nil
}

func (e *OCMEnvironment) Clusters() []*cmv1.Cluster {
	return append(make([]*cmv1.Cluster, 0, len(e.clusters)), e.clusters...)
}

func (e *OCMEnvironment) AddClusterRoutes() error {
	clusterRoutes, err := NewClusterListPager(e.clusters...).ToRoutes()
	if err != nil {
		return err
	}

	for _, r := range clusterRoutes {
		e.apiServer.RouteToHandler(r.Method, r.Path, r.Handler)
	}

	return nil
}

func (e *OCMEnvironment) Subscriptions() []*amv1.Subscription {
	return append(make([]*amv1.Subscription, 0, len(e.subscriptions)), e.subscriptions...)
}

func (e *OCMEnvironment) AddSubscriptionRoutes() error {
	var buf bytes.Buffer

	for _, sub := range e.subscriptions {
		err := amv1.MarshalSubscription(sub, &buf)
		if err != nil {
			return err
		}

		e.apiServer.RouteToHandler(
			"GET",
			slashJoin("/api/accounts_mgmt/v1/subscriptions", sub.ID()),
			sdktesting.RespondWithJSON(http.StatusOK, buf.String()),
		)

		buf.Reset()
	}

	return nil
}

func (e *OCMEnvironment) APIServerURL() string {
	return e.apiServer.URL()
}

func (e *OCMEnvironment) Config() string {
	return filepath.Join(e.tmpDir, ".ocm.json")
}

func (e *OCMEnvironment) SSOServerURL() string {
	return e.ssoServer.URL()
}

func (e *OCMEnvironment) CleanUp() error {
	if e.apiServer != nil {
		e.apiServer.Close()
	}

	if e.ssoServer != nil {
		e.ssoServer.Close()
	}

	if e.tmpDir != "" {
		err := os.RemoveAll(e.tmpDir)
		if err != nil {
			return err
		}
	}

	return nil
}

type OCMEnvironmentOption func(e *OCMEnvironment) error

func OCMEnvironmentAddons(addons ...*cmv1.AddOn) OCMEnvironmentOption {
	return func(e *OCMEnvironment) error {
		e.addons = append(e.addons, addons...)

		return e.AddAddonRoutes()
	}
}

func OCMEnvironmentClusters(clusters ...*cmv1.Cluster) OCMEnvironmentOption {
	return func(e *OCMEnvironment) error {
		e.clusters = append(e.clusters, clusters...)

		return e.AddClusterRoutes()
	}
}

func OCMEnvironmentSSOServer(serv *ghttp.Server) OCMEnvironmentOption {
	return func(e *OCMEnvironment) error {
		e.ssoServer = serv

		return nil
	}
}

func OCMEnvironmentSubscriptions(subs ...*amv1.Subscription) OCMEnvironmentOption {
	return func(e *OCMEnvironment) error {
		e.subscriptions = append(e.subscriptions, subs...)

		return e.AddSubscriptionRoutes()
	}
}
