package ocm

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// RetrieveAddons initializes an Iterator which will request addons from OCM with a fixed page size.
func RetrieveAddons(conn *sdk.Connection, logger log.Interface) (*AddonIterator, error) {
	return &AddonIterator{
		conn:    conn,
		logger:  logger,
		request: conn.ClustersMgmt().V1().Addons().List(),
	}, nil
}

type AddonIterator struct {
	conn    *sdk.Connection
	logger  log.Interface
	request *cmv1.AddOnsListRequest
}

// SearchByNameOrID filters the addons requested for those
// whose 'name' or 'id' matches the supplied pattern.
func (i *AddonIterator) SearchByNameOrID(pattern string) *AddonIterator {
	if pattern == "" {
		return i
	}

	query := fmt.Sprintf(
		"name like '%s' or id like '%s'",
		pattern,
		pattern,
	)

	return i.Search(query)
}

// FindByIDs uses the supplied addon IDs to filter the request to OCM
// and return only the addons which match.
func (i *AddonIterator) FindByIDs(ids ...string) *AddonIterator {
	if len(ids) == 0 {
		return i
	}

	quotedIDs := make([]string, 0, len(ids))

	for _, id := range ids {
		quotedIDs = append(quotedIDs, fmt.Sprintf("'%s'", id))
	}

	query := fmt.Sprintf("id in (%s)", strings.Join(quotedIDs, ","))

	return i.Search(query)
}

// Search filters the addons requested by a generic query string.
// See 'ocm-sdk-go' for more information on the SQL-like strings that
// are accepted.
func (i *AddonIterator) Search(query string) *AddonIterator {
	return &AddonIterator{
		conn:    i.conn,
		logger:  i.logger,
		request: i.request.Search(query),
	}
}

// ForEach iterates over the addons requested applying the provided function.
// If the iteration will stop with the first error returned by the provided function.
func (i *AddonIterator) ForEach(ctx context.Context, applyFunc func(*Addon) error) error {
	res, err := i.request.SendContext(ctx)
	if err != nil {
		return err
	}

	var finalErr error

	res.Items().Each(func(a *cmv1.AddOn) bool {
		addon := NewAddon(a, WithConnection{i.conn}, WithLogger{i.logger})

		if err := applyFunc(&addon); err != nil {
			finalErr = err

			return false
		}

		return true
	})

	return finalErr
}
