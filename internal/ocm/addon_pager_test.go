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

func TestAddonPagerIteration(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	expectedIterations := 149

	pager := setupAddonPager(expectedIterations)

	var actualIterations int

	err := pager.ForEach(context.Background(), func(addon *Addon) error {
		actualIterations++

		return nil
	})

	assert.Nil(err, "should not return an error")
	assert.Equal(expectedIterations, actualIterations, "should iterate exactly once for each addon")
}

var errAddonShortCircuit = errors.New("short-circuit")

func TestAddonPagerShortcircuit(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	expectedIterations := 25

	pager := setupAddonPager(addonPageSize - 1)

	var actualIterations int

	err := pager.ForEach(context.Background(), func(addon *Addon) error {
		if addon.Name() == fmt.Sprintf("test-addon-%d", expectedIterations) {
			return errAddonShortCircuit
		}

		actualIterations++

		return nil
	})

	assert.ErrorIs(err, errAddonShortCircuit, "should return error when short circuit condition is reached")
	assert.Equal(expectedIterations, actualIterations, "should only iterate until short circuit is reached")
}

func setupAddonPager(totalItems int) *AddonPager {
	response := &addonsListResponseMock{}

	for i := totalItems; i > 0; i -= addonPageSize {
		returnSize := addonPageSize

		if i < addonPageSize {
			returnSize = i
		}

		response.
			On("Items").
			Return(addonList(returnSize)).
			Once()
		response.
			On("Size").
			Return(returnSize).
			Once()
	}

	expectedPageRequests := int(math.Ceil(float64(totalItems) / float64(addonPageSize)))

	request := &addonsListRequestMock{}
	request.
		On("RequestPage").
		Return(response, nil).
		Times(expectedPageRequests)

	return &AddonPager{
		index:   1,
		request: request,
	}
}

func addonList(size int) *cmv1.AddOnList {
	addonList := make([]*cmv1.AddOnBuilder, size)
	for i := 0; i < size; i++ {
		addonList[i] = cmv1.
			NewAddOn().
			Name(fmt.Sprintf("test-addon-%d", i))
	}

	result, _ := cmv1.
		NewAddOnList().
		Items(addonList...).
		Build()

	return result
}

type addonsListRequestMock struct {
	mock.Mock
}

func (a *addonsListRequestMock) Search(query string) addonsListRequester {
	return a
}

func (a *addonsListRequestMock) RequestPage(ctx context.Context, page, size int) (addonsListResponser, error) {
	args := a.Called()

	return args.Get(0).(*addonsListResponseMock), args.Error(1) //nolint:forcetypeassert
}

type addonsListResponseMock struct {
	mock.Mock
}

func (a *addonsListResponseMock) Items() *cmv1.AddOnList {
	args := a.Called()

	return args.Get(0).(*cmv1.AddOnList) //nolint:forcetypeassert
}

func (a *addonsListResponseMock) Size() int {
	args := a.Called()

	return args.Int(0)
}
