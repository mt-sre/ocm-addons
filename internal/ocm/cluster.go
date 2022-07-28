package ocm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	slv1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
)

const ocmTimeFormat = "2006-01-02 15:04:05"

var errInstallationNotFound = errors.New("installation not found")

func NewCluster(cluster *cmv1.Cluster, opts ...ClusterOption) Cluster {
	c := Cluster{
		cluster: cluster,
	}

	c.cfg.Option(opts...)
	c.cfg.Default()

	return c
}

// Cluster wraps for 'ocm-sdk-go' Cluster objects.
type Cluster struct {
	cfg                ClusterConfig
	cluster            *cmv1.Cluster
	subscription       *Subscription
	AddonInstallations AddonInstallations
}

func (c *Cluster) ExternalID() string { return c.cluster.ExternalID() }
func (c *Cluster) ID() string         { return c.cluster.ID() }
func (c *Cluster) Name() string       { return c.cluster.Name() }

func (c *Cluster) ProvideRowData() map[string]interface{} {
	result := map[string]interface{}{
		"Additional Trust Bundle":          c.cluster.AdditionalTrustBundle(),
		"API Listening Method":             c.cluster.API().Listening(),
		"API URL":                          c.cluster.API().URL(),
		"Billing Model":                    c.cluster.BillingModel(),
		"CCS Disable SCP Checks":           c.cluster.CCS().DisableSCPChecks(),
		"CCS Enabled":                      c.cluster.CCS().Enabled(),
		"CCS ID":                           c.cluster.CCS().ID(),
		"CloudProvider ID":                 c.cluster.CloudProvider().ID(),
		"CloudProvider Display Name":       c.cluster.CloudProvider().DisplayName(),
		"CloudProvider Name":               c.cluster.CloudProvider().Name(),
		"Console URL":                      c.cluster.Console().URL(),
		"Creation Timestamp":               c.cluster.CreationTimestamp(),
		"Disable User Workload Monitoring": c.cluster.DisableUserWorkloadMonitoring(),
		"DNS Base Domain":                  c.cluster.DNS().BaseDomain(),
		"ETCD Encryption":                  c.cluster.EtcdEncryption(),
		"Expiration Timestamp":             c.cluster.ExpirationTimestamp(),
		"External ID":                      c.cluster.ExternalID(),
		"FIPS":                             c.cluster.FIPS(),
		"Health State":                     c.cluster.HealthState(),
		"ID":                               c.cluster.ID(),
		"Installed Addons":                 c.installedAddons(),
		"Load Balancer Qutoa":              c.cluster.LoadBalancerQuota(),
		"Managed":                          c.cluster.Managed(),
		"Multi AZ":                         c.cluster.MultiAZ(),
		"Name":                             c.cluster.Name(),
		"Network Host Prefix":              c.cluster.Network().HostPrefix(),
		"Network Machine CIDR":             c.cluster.Network().MachineCIDR(),
		"Network Pod CIDR":                 c.cluster.Network().PodCIDR(),
		"Network Service CIDR":             c.cluster.Network().ServiceCIDR(),
		"Network Type":                     c.cluster.Network().Type(),
		"OpenShift Version":                c.cluster.OpenshiftVersion(),
		"Product ID":                       c.cluster.Product().ID(),
		"Product Name":                     c.cluster.Product().Name(),
		"HTTP Proxy":                       c.cluster.Proxy().HTTPProxy(),
		"HTTPS Proxy":                      c.cluster.Proxy().HTTPSProxy(),
		"State":                            c.cluster.State(),
		"Subscription ID":                  c.cluster.Subscription().ID(),
	}

	for k, v := range c.subscription.ProvideRowData() {
		result[k] = v
	}

	return result
}

// WithSubscription attempts to retrieve a subscription object corresponding
// to the subscription ID included with the cluster. If the subscription
// cannot be retrieved an error is returned.
func (c *Cluster) WithSubscription(ctx context.Context) (*Cluster, error) {
	trace := c.cfg.Logger.
		WithFields(log.Fields{
			"cluster":      c.cluster.ID(),
			"subscription": c.cluster.Subscription().ID(),
		}).
		Trace("requesting subscription information")
	defer trace.Stop(nil)

	sub, err := c.cfg.Conn.
		AccountsMgmt().
		V1().
		Subscriptions().
		Subscription(c.cluster.Subscription().ID()).
		Get().
		Parameter("fetchAccounts", true).
		SendContext(ctx)
	if err != nil {
		return c, err
	}

	c.subscription = &Subscription{
		sub: sub.Body(),
	}

	return c, nil
}

// WithAddonInstallations attempts to retrieve and abstract information
// about installed addons on the cluster. Any failure to retrieve data
// will return an error.
func (c *Cluster) WithAddonInstallations(ctx context.Context) (*Cluster, error) {
	installs, err := c.retrieveInstallations(ctx)
	if err != nil {
		return c, err
	}

	if len(installs) == 0 {
		return c, nil
	}

	ids := make([]string, 0, len(installs))

	for _, install := range installs {
		ids = append(ids, install.ID())
	}

	addons, err := RetrieveAddons(c.cfg.Conn, c.cfg.Logger)
	if err != nil {
		return c, err
	}

	matchingAddons := addons.FindByIDs(ids...)

	c.AddonInstallations = make([]AddonInstallation, 0, len(installs))

	_ = matchingAddons.ForEach(ctx, func(addon *Addon) error {
		install, err := findInstallationByID(installs, addon.ID())
		if err != nil {
			return nil
		}

		c.AddonInstallations = append(c.AddonInstallations, NewAddonInstallation(
			install,
			WithAddon{Addon: addon},
			WithCluster{Cluster: c},
		))

		return nil
	})

	return c, nil
}

func (c *Cluster) retrieveInstallations(ctx context.Context) ([]*cmv1.AddOnInstallation, error) {
	trace := c.cfg.Logger.
		WithFields(log.Fields{
			"cluster": c.cluster.ID(),
		}).
		Trace("requesting addon installations")
	defer trace.Stop(nil)

	res, err := c.cfg.Conn.
		ClustersMgmt().
		V1().
		Clusters().
		Cluster(c.cluster.ID()).
		Addons().
		List().
		SendContext(ctx)
	if err != nil {
		return nil, err
	}

	trace.Infof("number of installations retrieved %d", res.Items().Len())

	return res.Items().Slice(), nil
}

func findInstallationByID(installs []*cmv1.AddOnInstallation, addonID string) (*cmv1.AddOnInstallation, error) {
	for _, addonInstallation := range installs {
		if addonInstallation.ID() != addonID {
			continue
		}

		return addonInstallation, nil
	}

	return nil, fmt.Errorf("finding addon with id %q: %w", addonID, errInstallationNotFound)
}

// installedAddons returns a comma-separated list of installed addons
// for the cluster and their status.
func (c *Cluster) installedAddons() string {
	displayValues := make([]string, 0, len(c.AddonInstallations))

	for _, install := range c.AddonInstallations {
		displayValues = append(
			displayValues,
			fmt.Sprintf("%s(%s)", install.ID(), install.State()),
		)
	}

	return strings.Join(displayValues, ",")
}

// ProductID returns a string indicating the product type
// (OSD, ROSA, ARO, ...) and whether the cluster is of the
// Customer Cloud Subscription (CSS) Model.
func (c *Cluster) ProductID() string {
	ccsDisplayValue := "no-ccs"

	if c.cluster.CCS().Enabled() {
		ccsDisplayValue = "ccs"
	}

	return fmt.Sprintf("%s,%s", c.cluster.Product().ID(), ccsDisplayValue)
}

func (c *Cluster) PostLog(ctx context.Context, opts ...LogEntryOption) error {
	trace := c.cfg.Logger.
		WithFields(log.Fields{
			"cluster": c.cluster.ID(),
		}).Trace("posting log entry")
	defer trace.Stop(nil)

	ent, err := NewLogEntry(c, opts...)
	if err != nil {
		return fmt.Errorf("generating log entry: %w", err)
	}

	trace.WithFields(log.Fields{
		"entrySeverity": ent.Entry.Severity(),
		"entrySummary":  ent.Entry.Summary(),
	}).Debug("generated entry")

	res, err := c.cfg.Conn.
		ServiceLogs().
		V1().
		ClusterLogs().
		Add().
		Body(ent.Entry).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("posting log entry: %w", err)
	}

	if res.Error() != nil {
		return fmt.Errorf("posting log entry failed with status %d: %w", res.Status(), res.Error())
	}

	return nil
}

func (c *Cluster) GetLogs(ctx context.Context, opts GetLogsOptions) ([]LogEntry, error) {
	query := opts.Query()

	trace := c.cfg.Logger.
		WithFields(log.Fields{
			"cluster": c.cluster.ID(),
			"query":   query,
		}).Trace("retrieving cluster log entries")
	defer trace.Stop(nil)

	res, err := c.cfg.Conn.
		ServiceLogs().
		V1().
		Clusters().
		Cluster(c.cluster.ExternalID()).
		ClusterLogs().
		List().
		Search(query).
		SendContext(ctx)
	if err != nil {
		return nil, err
	}

	entries := NewLogEntrySorter(res.Items().Len(), opts.sorter)

	res.Items().Each(func(entry *slv1.LogEntry) bool {
		entries.Append(LogEntry{Entry: entry})

		return true
	})

	sort.Sort(entries)

	return entries.Entries(), nil
}

func NewGetLogsOptions(opts ...GetLogsOption) GetLogsOptions {
	var glo GetLogsOptions

	for _, opt := range opts {
		opt(&glo)
	}

	return glo
}

type GetLogsOptions struct {
	pattern string
	lvl     LogLevel
	sorter  LogEntrySortFunc
	before  time.Time
	after   time.Time
}

func (g GetLogsOptions) Query() string {
	var predicates []string

	if g.pattern != "" {
		predicates = append(predicates, fmt.Sprintf("description like '%s'", g.pattern))
	}

	if g.lvl != LogLevelNone {
		predicates = append(predicates, fmt.Sprintf("severity = '%s'", g.lvl))
	}

	epoch := time.Time{}

	if g.after.After(epoch) {
		predicates = append(predicates, fmt.Sprintf("timestamp >= '%s'", g.after.Format(ocmTimeFormat)))
	}

	if g.before.After(epoch) {
		predicates = append(predicates, fmt.Sprintf("timestamp <= '%s'", g.before.Format(ocmTimeFormat)))
	}

	return strings.Join(predicates, " and ")
}

type GetLogsOption func(*GetLogsOptions)

func GetLogsMatchingPattern(p string) GetLogsOption {
	return func(g *GetLogsOptions) {
		g.pattern = p
	}
}

func GetLogsWithLevel(l LogLevel) GetLogsOption {
	return func(g *GetLogsOptions) {
		g.lvl = l
	}
}

func GetLogsSorted(s LogEntrySortFunc) GetLogsOption {
	return func(g *GetLogsOptions) {
		g.sorter = s
	}
}

func GetLogsBefore(t time.Time) GetLogsOption {
	return func(g *GetLogsOptions) {
		g.before = t
	}
}

func GetLogsAfter(t time.Time) GetLogsOption {
	return func(g *GetLogsOptions) {
		g.after = t
	}
}

type ClusterConfig struct {
	Conn   *sdk.Connection
	Logger log.Interface
}

func (c *ClusterConfig) Option(opts ...ClusterOption) {
	for _, opt := range opts {
		opt.ConfigureCluster(c)
	}
}

func (c *ClusterConfig) Default() {
	if c.Logger == nil {
		c.Logger = &log.Logger{
			Handler: discard.New(),
		}
	}
}

type ClusterOption interface {
	ConfigureCluster(*ClusterConfig)
}

// Subscription wraps an 'ocm-sdk-go' Subscription object.
type Subscription struct {
	sub *amv1.Subscription
}

func (s *Subscription) ProvideRowData() map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Creator ID":       s.sub.Creator().ID(),
		"Creator Email":    s.sub.Creator().Email(),
		"Creator Name":     fmt.Sprintf("%s %s", s.sub.Creator().FirstName(), s.sub.Creator().LastName()),
		"Creator Username": s.sub.Creator().Username(),
		"Organization ID":  s.sub.OrganizationID(),
		"Support Level":    s.sub.SupportLevel(),
	}
}
