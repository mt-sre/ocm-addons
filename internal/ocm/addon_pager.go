package ocm

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

const (
	addonPageSize = 50
)

// RetrieveAddons initializes a Pager which will request addons from OCM with a fixed page size.
func RetrieveAddons(conn *sdk.Connection, logger log.Interface) (*AddonPager, error) {
	request := &addonsListRequest{
		conn.ClustersMgmt().V1().Addons().List(),
	}

	return &AddonPager{
		conn:    conn,
		index:   1,
		logger:  logger,
		request: request,
	}, nil
}

// AddonPager retains state for paged addon requests and maintains a buffer
// of the last page of objects.
type AddonPager struct {
	buffer    []Addon
	finalPage bool
	index     int
	conn      *sdk.Connection
	logger    log.Interface
	request   addonsListRequester
}

// SearchByNameOrID filters the addons requested by a Pager for those
// whose 'name' or 'id' matches the supplied pattern.
func (p *AddonPager) SearchByNameOrID(pattern string) *AddonPager {
	if pattern == "" {
		return p
	}

	query := fmt.Sprintf(
		"name like '%s' or id like '%s'",
		pattern,
		pattern,
	)

	return p.Search(query)
}

// FindByIDs uses the supplied addon IDs to filter the request to OCM
// and return only the addons which match.
func (p *AddonPager) FindByIDs(ids ...string) *AddonPager {
	if len(ids) == 0 {
		return p
	}

	quotedIDs := make([]string, 0, len(ids))

	for _, id := range ids {
		quotedIDs = append(quotedIDs, fmt.Sprintf("'%s'", id))
	}

	query := fmt.Sprintf("id in (%s)", strings.Join(quotedIDs, ","))

	return p.Search(query)
}

// Search filters the addons requested by a generic query string.
// See 'ocm-sdk-go' for more information on the SQL-like strings that
// are accepted.
func (p *AddonPager) Search(query string) *AddonPager {
	return &AddonPager{
		conn:    p.conn,
		index:   1,
		logger:  p.logger,
		request: p.request.Search(query),
	}
}

// ForEach iterates over the addons requested by an Pager applying
// the provided function. If the iteration will stop with the first
// error returned by the provided function.
func (p *AddonPager) ForEach(ctx context.Context, applyFunc func(*Addon) error) error {
	for {
		addons, hasMorePages, err := p.NextPage(ctx)
		if err != nil {
			return err
		}

		if !hasMorePages {
			return nil
		}

		for i := range addons {
			err = applyFunc(&addons[i])
			if err != nil {
				return err
			}
		}
	}
}

// NextPage returns the next page of requested addons if there are any remaining.
// If no addons remain the second return value will be 'false'.
func (p *AddonPager) NextPage(ctx context.Context) ([]Addon, bool, error) {
	if p.finalPage {
		return nil, false, nil
	}

	if p.buffer == nil {
		p.buffer = make([]Addon, addonPageSize)
	}

	p.buffer = p.buffer[:0]

	res, err := p.request.RequestPage(ctx, p.index, addonPageSize)
	if err != nil {
		return nil, false, err
	}

	for _, addon := range res.Items().Slice() {
		p.buffer = append(p.buffer,
			NewAddon(addon,
				WithConnection{Connection: p.conn},
				WithLogger{Logger: p.logger},
			),
		)
	}

	if res.Size() < addonPageSize {
		p.finalPage = true
	}

	p.index++

	return p.buffer, true, nil
}
