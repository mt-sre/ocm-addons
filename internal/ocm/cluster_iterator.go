package ocm

import (
	"context"
	"fmt"

	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// RetrieveClusters initializes a ClusterIterator which will request clusters from OCM with a fixed page size.
func RetrieveClusters(conn *sdk.Connection, logger log.Interface) (*ClusterIterator, error) {
	request := conn.ClustersMgmt().V1().Clusters().List()

	return &ClusterIterator{
		conn:    conn,
		logger:  logger,
		request: request,
	}, nil
}

// ClusterIterator retains state for paged cluster requests and maintains a buffer
// of the last page of objects.
type ClusterIterator struct {
	conn    *sdk.Connection
	logger  log.Interface
	request *cmv1.ClustersListRequest
}

// SearchByNameOrID filters the clusters requested by an ClusterIterator for those
// whose 'name', 'id', or 'external_id' matches the supplied pattern.
func (i *ClusterIterator) SearchByNameOrID(pattern string) *ClusterIterator {
	if pattern == "" {
		return i
	}

	query := fmt.Sprintf(
		"name like '%s' or id = '%s' or external_id = '%s'",
		pattern,
		pattern,
		pattern,
	)

	return i.Search(query)
}

// Search filters the clusters requested by a generic query string.
// See 'ocm-sdk-go' for more information on the SQL-like strings that
// are accepted.
func (i *ClusterIterator) Search(query string) *ClusterIterator {
	return &ClusterIterator{
		conn:    i.conn,
		logger:  i.logger,
		request: i.request.Search(query),
	}
}

// ForEach iterates over the clusters requested by a ClusterIterator applying
// the provided function. If the iteration will stop with the first
// error returned by the provided function.
func (i *ClusterIterator) ForEach(ctx context.Context, applyFunc func(*Cluster) error) error {
	res, err := i.request.SendContext(ctx)
	if err != nil {
		return err
	}

	var finalErr error

	res.Items().Each(func(c *cmv1.Cluster) bool {
		cluster := NewCluster(c, WithConnection{i.conn}, WithLogger{i.logger})

		if err := applyFunc(&cluster); err != nil {
			finalErr = err

			return false
		}

		return true
	})

	return finalErr
}
