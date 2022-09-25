// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type clustersListRequester interface {
	Search(string) clustersListRequester
	RequestPage(context.Context, int, int) (clustersListResponser, error)
}

type clustersListRequest struct {
	*cmv1.ClustersListRequest
}

var _ clustersListRequester = (*clustersListRequest)(nil)

func (c *clustersListRequest) Search(query string) clustersListRequester {
	c.ClustersListRequest = c.ClustersListRequest.Search(query)

	return c
}

func (c *clustersListRequest) RequestPage(ctx context.Context, page, size int) (clustersListResponser, error) {
	response, err := c.ClustersListRequest.
		Size(clusterPageSize).
		Page(page).
		SendContext(ctx)

	return &clustersListResponse{
		ClustersListResponse: response,
	}, err
}

type clustersListResponser interface {
	Items() *cmv1.ClusterList
	Size() int
}

var _ clustersListResponser = (*clustersListResponse)(nil)

type clustersListResponse struct {
	*cmv1.ClustersListResponse
}

func (a *clustersListResponse) Items() *cmv1.ClusterList {
	return a.ClustersListResponse.Items()
}
