package ocm

import (
	"strings"

	"github.com/mt-sre/ocm-addons/internal/output"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// AddonInstallations wraps a slice of AddonInstallation objects.
type AddonInstallations []AddonInstallation

// Matching filters the addons within an AddonInstallations object by
// only including addons whose name or id matches the supplied pattern.
func (installations AddonInstallations) Matching(pattern string) AddonInstallations {
	var result AddonInstallations

	for _, install := range installations {
		if strings.Contains(install.addon.Name(), pattern) ||
			strings.Contains(install.addon.ID(), pattern) {
			result = append(result, install)
		}
	}

	return result
}

// AddonInstallation provides details of an AddonInstallation.
type AddonInstallation struct {
	*cmv1.AddOnInstallation
	cluster *Cluster
	addon   *Addon
}

// ToRowObject presents addon installation data as a record with static fields.
// This is currently used to make the 'ocm-cli' TableWriter happy as the
// logic which resolves column names to fields using reflection has trouble
// with AddonInstallations.
func (a *AddonInstallation) ToRowObject() *AddonInstallationRowObject {
	return &AddonInstallationRowObject{
		AddonID:     a.addon.ID(),
		AddonName:   a.addon.Name(),
		ClusterID:   a.cluster.ID(),
		ClusterName: a.cluster.Name(),
		State:       string(a.State()),
	}
}

type AddonInstallationRowObject struct {
	AddonID     string
	AddonName   string
	ClusterID   string
	ClusterName string
	State       string
}

func (a *AddonInstallation) ToRow() output.Row {
	return output.Row{
		{
			Name:  "Addon ID",
			Value: a.addon.ID(),
		}, {
			Name:  "Addon Name",
			Value: a.addon.Name(),
		}, {
			Name:  "Cluster ID",
			Value: a.cluster.ID(),
		}, {
			Name:  "Cluster Name",
			Value: a.cluster.Name(),
		}, {
			Name:  "State",
			Value: a.State(),
		},
	}
}
