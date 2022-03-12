package cli

import (
	"strings"
	"time"

	"github.com/mt-sre/ocm-addons/internal/ocm"
	"github.com/spf13/pflag"
)

type CommonOptions struct {
	Columns   string
	NoHeaders bool
	NoColor   bool
}

func (c *CommonOptions) DefaultColumns(cols string) {
	c.Columns = cols
}

func (c *CommonOptions) AddColumnsFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&c.Columns,
		"columns",
		c.Columns,
		"comma separated list of columns to display",
	)
}

func (c *CommonOptions) AddNoHeadersFlag(flags *pflag.FlagSet) {
	flags.BoolVar(
		&c.NoHeaders,
		"no-headers",
		c.NoHeaders,
		"omits header row",
	)
}

func (c *CommonOptions) AddNoColorFlag(flags *pflag.FlagSet) {
	flags.BoolVar(
		&c.NoColor,
		"no-color",
		c.NoColor,
		"disables colorized output",
	)
}

type SearchOptions struct {
	Search    string
	searchUsg string
}

func (s *SearchOptions) SearchUsage(usg string) {
	s.searchUsg = usg
}

func (s *SearchOptions) AddSearchFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&s.Search,
		"search",
		s.Search,
		s.searchUsg,
	)
}

type FilterOptions struct {
	Order     ocm.Order
	orderIn   string
	orderUsg  string
	Before    time.Time
	beforeIn  string
	beforeUsg string
	After     time.Time
	afterIn   string
	afterUsg  string
}

func (f *FilterOptions) OrderDefault(ord string) {
	f.orderIn = ord
}

func (f *FilterOptions) OrderUsage(usg string) {
	f.orderUsg = usg
}

func (f *FilterOptions) AddOrderFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&f.orderIn,
		"order",
		f.orderIn,
		f.orderUsg,
	)
}

func (f *FilterOptions) BeforeUsage(usg string) {
	f.beforeUsg = usg
}

func (f *FilterOptions) AddBeforeFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&f.beforeIn,
		"before",
		f.beforeIn,
		f.beforeUsg,
	)
}

func (f *FilterOptions) AfterUsage(usg string) {
	f.afterUsg = usg
}

func (f *FilterOptions) AddAfterFlag(flags *pflag.FlagSet) {
	flags.StringVar(
		&f.afterIn,
		"after",
		f.afterIn,
		f.afterUsg,
	)
}

func (f *FilterOptions) ParseFilterOptions() error {
	var err error

	f.Order = ParseOrder(f.orderIn)

	if f.beforeIn != "" {
		if f.Before, err = ParseTime(f.beforeIn); err != nil {
			return err
		}
	}

	if f.afterIn != "" {
		if f.After, err = ParseTime(f.afterIn); err != nil {
			return err
		}
	}

	return nil
}

func ParseOrder(maybeOrder string) ocm.Order {
	switch maybe := strings.ToLower(strings.TrimSpace(maybeOrder)); {
	case maybe == "asc" || maybe == "ascending":
		return ocm.OrderAsc
	case maybe == "desc" || maybe == "descending":
		return ocm.OrderDesc
	default:
		return ocm.OrderNone
	}
}

const timeFormat = "2006-01-02 15:04:05"

func ParseTime(maybeTime string) (time.Time, error) {
	return time.Parse(timeFormat, maybeTime)
}
