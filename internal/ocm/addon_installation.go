package ocm

import (
	"fmt"
	"strings"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// AddonInstallations wraps a slice of AddonInstallation objects.
type AddonInstallations []AddonInstallation

// Matching filters the addons within an AddonInstallations object by
// only including addons whose name or id matches the supplied pattern.
func (installs AddonInstallations) Matching(pattern string) AddonInstallations {
	var result AddonInstallations

	for _, install := range installs {
		if strings.Contains(install.Name(), pattern) ||
			strings.Contains(install.ID(), pattern) {
			result = append(result, install)
		}
	}

	return result
}

func NewAddonInstallation(install *cmv1.AddOnInstallation, opts ...AddonInstallationOption) AddonInstallation {
	var cfg AddonInstallationConfig

	cfg.Option(opts...)

	return AddonInstallation{
		install: install,
		cfg:     cfg,
	}
}

// AddonInstallation provides details of an AddonInstallation.
type AddonInstallation struct {
	install *cmv1.AddOnInstallation
	cfg     AddonInstallationConfig
}

func (a *AddonInstallation) ID() string    { return a.cfg.Addon.ID() }
func (a *AddonInstallation) Name() string  { return a.cfg.Addon.Name() }
func (a *AddonInstallation) State() string { return string(a.install.State()) }

func (a *AddonInstallation) ProvideRowData() map[string]interface{} {
	result := map[string]interface{}{
		"Creation Timestamp":   a.install.CreationTimestamp(),
		"Installed Version ID": a.install.AddonVersion().ID(),
		"Operator Version":     a.install.OperatorVersion(),
		"Parameters":           a.parameters(),
		"State":                a.install.State(),
		"State Description":    a.install.StateDescription(),
		"Updated Timestamp":    a.install.UpdatedTimestamp(),
	}

	for k, v := range a.cfg.Addon.ProvideRowData() {
		result["Addon "+k] = v
	}

	for k, v := range a.cfg.Cluster.ProvideRowData() {
		result["Cluster "+k] = v
	}

	return result
}

func (a *AddonInstallation) parameters() string {
	if a.install.Parameters() == nil {
		return "None"
	}

	var res []string

	a.install.Parameters().Each(func(param *cmv1.AddOnInstallationParameter) bool {
		res = append(res, fmt.Sprintf("%s: %s", param.ID(), param.Value()))

		return true
	})

	return strings.Join(res, ", ")
}

type AddonInstallationConfig struct {
	Cluster *Cluster
	Addon   *Addon
}

func (c *AddonInstallationConfig) Option(opts ...AddonInstallationOption) {
	for _, opt := range opts {
		opt.ConfigureAddonInstallation(c)
	}
}

type AddonInstallationOption interface {
	ConfigureAddonInstallation(*AddonInstallationConfig)
}

type WithCluster struct{ *Cluster }

func (cl WithCluster) ConfigureAddonInstallation(c *AddonInstallationConfig) {
	c.Cluster = cl.Cluster
}

type WithAddon struct{ *Addon }

func (a WithAddon) ConfigureAddonInstallation(c *AddonInstallationConfig) {
	c.Addon = a.Addon
}
