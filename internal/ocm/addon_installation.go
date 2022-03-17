package ocm

import (
	"strings"

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

type AddonInstallationRowObject struct {
	AddonID     string
	AddonName   string
	ClusterID   string
	ClusterName string
	State       string
}

func (a *AddonInstallation) ProvideRowData() map[string]interface{} {
	result := map[string]interface{}{
		"State": a.State(),
	}

	for k, v := range a.addon.ProvideRowData() {
		result["Addon "+k] = v
	}

	for k, v := range a.cluster.ProvideRowData() {
		result["Cluster "+k] = v
	}

	return result
}
