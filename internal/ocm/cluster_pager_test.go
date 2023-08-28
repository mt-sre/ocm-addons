// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClusterPagerIteration(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	expectedIterations := 149

	pager := setupClusterPager(expectedIterations)

	var actualIterations int

	err := pager.ForEach(context.Background(), func(cluster *Cluster) error {
		actualIterations++

		return nil
	})

	assert.Nil(err, "should not return an error")
	assert.Equal(expectedIterations, actualIterations, "should iterate exactly once for each cluster")
}

var errClusterShortCircuit = errors.New("short-circuit")

func TestClusterPagerShortcircuit(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	expectedIterations := 25

	pager := setupClusterPager(clusterPageSize - 1)

	var actualIterations int

	err := pager.ForEach(context.Background(), func(cluster *Cluster) error {
		if cluster.Name() == fmt.Sprintf("test-cluster-%d", expectedIterations) {
			return errClusterShortCircuit
		}

		actualIterations++

		return nil
	})

	assert.ErrorIs(err, errClusterShortCircuit, "should return error when short circuit condition is reached")
	assert.Equal(expectedIterations, actualIterations, "should only iterate until short circuit is reached")
}

func setupClusterPager(totalItems int) *ClusterPager {
	response := &clustersListResponseMock{}

	for i := totalItems; i > 0; i -= clusterPageSize {
		returnSize := clusterPageSize

		if i < clusterPageSize {
			returnSize = i
		}

		response.
			On("Items").
			Return(clusterList(returnSize)).
			Once()
		response.
			On("Size").
			Return(returnSize).
			Once()
	}

	expectedPageRequests := int(math.Ceil(float64(totalItems) / float64(clusterPageSize)))

	request := &clustersListRequestMock{}
	request.
		On("RequestPage").
		Return(response, nil).
		Times(expectedPageRequests)

	return &ClusterPager{
		index:   1,
		request: request,
	}
}

func clusterList(size int) *cmv1.ClusterList {
	clusterList := make([]*cmv1.ClusterBuilder, size)
	for i := 0; i < size; i++ {
		clusterList[i] = cmv1.
			NewCluster().
			Name(fmt.Sprintf("test-cluster-%d", i))
	}

	result, _ := cmv1.
		NewClusterList().
		Items(clusterList...).
		Build()

	return result
}

type clustersListRequestMock struct {
	mock.Mock
}

var _ clustersListRequester = (*clustersListRequest)(nil)

func (a *clustersListRequestMock) Search(string) clustersListRequester {
	return a
}

func (a *clustersListRequestMock) RequestPage(context.Context, int, int) (clustersListResponser, error) {
	args := a.Called()

	return args.Get(0).(*clustersListResponseMock), args.Error(1) //nolint:forcetypeassert
}

var _ clustersListResponser = (*clustersListResponseMock)(nil)

type clustersListResponseMock struct {
	mock.Mock
}

func (a *clustersListResponseMock) Items() *cmv1.ClusterList {
	args := a.Called()

	return args.Get(0).(*cmv1.ClusterList) //nolint:forcetypeassert
}

func (a *clustersListResponseMock) Size() int {
	args := a.Called()

	return args.Int(0)
}
