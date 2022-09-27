// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type addonsListRequester interface {
	Search(string) addonsListRequester
	RequestPage(context.Context, int, int) (addonsListResponser, error)
}

type addonsListRequest struct {
	*cmv1.AddOnsListRequest
}

func (a *addonsListRequest) Search(query string) addonsListRequester {
	a.AddOnsListRequest = a.AddOnsListRequest.Search(query)

	return a
}

func (a *addonsListRequest) RequestPage(ctx context.Context, page, size int) (addonsListResponser, error) {
	response, err := a.AddOnsListRequest.
		Size(addonPageSize).
		Page(page).
		SendContext(ctx)

	return &addonsListResponse{
		AddOnsListResponse: response,
	}, err
}

type addonsListResponser interface {
	Items() *cmv1.AddOnList
	Size() int
}

type addonsListResponse struct {
	*cmv1.AddOnsListResponse
}

func (a *addonsListResponse) Items() *cmv1.AddOnList {
	return a.AddOnsListResponse.Items()
}
