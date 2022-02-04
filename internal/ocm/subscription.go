package ocm

import amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"

// Subscription wraps an 'ocm-sdk-go' Subscription object.
type Subscription struct {
	*amv1.Subscription
}
