package ocm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	slv1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
)

const ocmTimeFormat = "2006-01-02 15:04:05"

var errInstallationNotFound = errors.New("installation not found")

// Cluster wraps for 'ocm-sdk-go' Cluster objects.
type Cluster struct {
	*cmv1.Cluster
	conn               *sdk.Connection
	logger             log.Interface
	subscription       *Subscription
	AddonInstallations AddonInstallations
}

// WithSubscription attempts to retrieve a subscription object corresponding
// to the subscription ID included with the cluster. If the subscription
// cannot be retrieved an error is returned.
func (c *Cluster) WithSubscription(ctx context.Context) (*Cluster, error) {
	trace := c.logger.
		WithFields(log.Fields{
			"cluster":      c.ID(),
			"subscription": c.Subscription().ID(),
		}).
		Trace("requesting subscription information")
	defer trace.Stop(nil)

	sub, err := c.conn.
		AccountsMgmt().
		V1().
		Subscriptions().
		Subscription(c.Subscription().ID()).
		Get().
		SendContext(ctx)
	if err != nil {
		return c, err
	}

	c.subscription = &Subscription{
		Subscription: sub.Body(),
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

	addons, err := RetrieveAddons(c.conn, c.logger)
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

		c.AddonInstallations = append(c.AddonInstallations, AddonInstallation{
			AddOnInstallation: install,
			addon:             addon,
			cluster:           c,
		})

		return nil
	})

	return c, nil
}

func (c *Cluster) retrieveInstallations(ctx context.Context) ([]*cmv1.AddOnInstallation, error) {
	trace := c.logger.
		WithFields(log.Fields{
			"cluster": c.ID(),
		}).
		Trace("requesting addon installations")
	defer trace.Stop(nil)

	res, err := c.conn.
		ClustersMgmt().
		V1().
		Clusters().
		Cluster(c.ID()).
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

// Domain returns the 'BaseDomain' of the cluster.
func (c *Cluster) Domain() string {
	return c.DNS().BaseDomain()
}

// InstalledAddons returns a comma-separated list of installed addons
// for the cluster and their status.
func (c *Cluster) InstalledAddons() string {
	displayValues := make([]string, 0, len(c.AddonInstallations))

	for _, install := range c.AddonInstallations {
		displayValues = append(
			displayValues,
			fmt.Sprintf("%s(%s)", install.addon.ID(), install.State()),
		)
	}

	return strings.Join(displayValues, ",")
}

// Organization returns the Organization ID associated with the
// subscription for this cluster if the subscription is
// retrievable and populated.
func (c *Cluster) Organization() string {
	if c.subscription == nil {
		return ""
	}

	return c.subscription.OrganizationID()
}

// ProductID returns a string indicating the product type
// (OSD, ROSA, ARO, ...) and whether the cluster is of the
// Customer Cloud Subscription (CSS) Model.
func (c *Cluster) ProductID() string {
	ccsDisplayValue := "no-ccs"

	if c.CCS().Enabled() {
		ccsDisplayValue = "ccs"
	}

	return fmt.Sprintf("%s,%s", c.Product().ID(), ccsDisplayValue)
}

func (c *Cluster) PostLog(ctx context.Context, opts ...LogEntryOption) error {
	trace := c.logger.
		WithFields(log.Fields{
			"cluster": c.ID(),
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

	res, err := c.conn.
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

	trace := c.logger.
		WithFields(log.Fields{
			"cluster": c.ID(),
			"query":   query,
		}).Trace("retrieving cluster log entries")
	defer trace.Stop(nil)

	res, err := c.conn.
		ServiceLogs().
		V1().
		Clusters().
		Cluster(c.ExternalID()).
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
