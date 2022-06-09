package ocm

import (
	"context"
	"fmt"

	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

const (
	clusterPageSize = 50
)

// RetrieveClusters initializes a ClusterPager which will request clusters from OCM with a fixed page size.
func RetrieveClusters(conn *sdk.Connection, logger log.Interface) (*ClusterPager, error) {
	request := &clustersListRequest{
		conn.ClustersMgmt().V1().Clusters().List(),
	}

	return &ClusterPager{
		conn:    conn,
		index:   1,
		logger:  logger,
		request: request,
	}, nil
}

// ClusterPager retains state for paged cluster requests and maintains a buffer
// of the last page of objects.
type ClusterPager struct {
	buffer    []Cluster
	conn      *sdk.Connection
	finalPage bool
	index     int
	logger    log.Interface
	request   clustersListRequester
}

// SearchByNameOrID filters the clusters requested by an ClusterPager for those
// whose 'name', 'id', or 'external_id' matches the supplied pattern.
func (p *ClusterPager) SearchByNameOrID(pattern string) *ClusterPager {
	if pattern == "" {
		return p
	}

	query := fmt.Sprintf(
		"name like '%s' or id = '%s' or external_id = '%s'",
		pattern,
		pattern,
		pattern,
	)

	return p.Search(query)
}

// Search filters the clusters requested by a generic query string.
// See 'ocm-sdk-go' for more information on the SQL-like strings that
// are accepted.
func (p *ClusterPager) Search(query string) *ClusterPager {
	return &ClusterPager{
		conn:    p.conn,
		logger:  p.logger,
		index:   1,
		request: p.request.Search(query),
	}
}

// ForEach iterates over the clusters requested by an ClusterPager applying
// the provided function. If the iteration will stop with the first
// error returned by the provided function.
func (p *ClusterPager) ForEach(ctx context.Context, applyFunc func(*Cluster) error) error {
	for {
		clusters, hasMorePages, err := p.NextPage(ctx)
		if err != nil {
			return err
		}

		if !hasMorePages {
			return nil
		}

		for i := range clusters {
			err = applyFunc(&clusters[i])
			if err != nil {
				return err
			}
		}
	}
}

// NextPage returns the next page of requested clusters if there are any remaining.
// If no clusters remain the second return value will be 'false'.
func (p *ClusterPager) NextPage(ctx context.Context) ([]Cluster, bool, error) {
	if p.finalPage {
		return nil, false, nil
	}

	if p.buffer == nil {
		p.buffer = make([]Cluster, clusterPageSize)
	}

	p.buffer = p.buffer[:0]

	res, err := p.request.RequestPage(ctx, p.index, clusterPageSize)
	if err != nil {
		return nil, false, err
	}

	for _, cluster := range res.Items().Slice() {
		p.buffer = append(p.buffer, NewCluster(cluster,
			WithConnection{Connection: p.conn},
			WithLogger{Logger: p.logger},
		))
	}

	if res.Size() < clusterPageSize {
		p.finalPage = true
	}

	p.index++

	return p.buffer, true, nil
}
