package ocm

import (
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// Addon wraps an 'ocm-sdk-go' AddOn object.
type Addon struct {
	*v1.AddOn
}
