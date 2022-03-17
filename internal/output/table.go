package output

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"go.uber.org/multierr"
)

type RowDataProvider interface {
	ProvideRowData() map[string]interface{}
}

func NewTable(opts ...TableOption) (*Table, error) {
	var table Table

	table.cfg.Option(opts...)
	table.cfg.Default()

	if table.cfg.PagerBin != "" {
		var err error

		table.pager, err = NewPager(table.cfg.PagerBin, table.cfg.Out)
		if err != nil {
			return nil, fmt.Errorf("starting pager: %w", err)
		}

		table.cfg.Out = table.pager
	}

	if !table.cfg.NoHeaders {
		table.writeHeaders()
	}

	return &table, nil
}

type Table struct {
	cfg   TableConfig
	data  [][]string
	pager *Pager
}

func (t *Table) writeHeaders() {
	t.data = append(t.data, t.formattedHeaders())
}

func (t *Table) formattedHeaders() []string {
	headers := make([]string, 0, len(t.cfg.Columns))

	for _, c := range t.cfg.Columns {
		headers = append(headers, t.cfg.HFormatter(c))
	}

	return headers
}

func (t *Table) Write(r RowDataProvider, mods ...RowModifier) error {
	row := NewRow(r.ProvideRowData())

	for _, mod := range mods {
		row = mod(row)
	}

	values := make([]string, 0, len(t.cfg.Columns))

	for _, col := range t.cfg.Columns {
		values = append(values, row.ValueString(normalize(col)))
	}

	t.data = append(t.data, values)

	return nil
}

func (t *Table) Flush() error {
	var errCollector error

	if err := t.flush(); err != nil {
		multierr.AppendInto(&errCollector, fmt.Errorf("flusing writer: %w", err))
	}

	if t.pager != nil {
		if err := t.pager.Close(); err != nil {
			multierr.AppendInto(&errCollector, fmt.Errorf("closing pager: %w", err))
		}
	}

	return errCollector
}

func (t *Table) flush() error {
	if t.cfg.NoColor {
		pterm.DisableColor()

		defer pterm.EnableColor()
	}

	printer := pterm.DefaultTable.WithData(t.data)

	if !t.cfg.NoHeaders {
		printer = printer.
			WithHasHeader().
			WithHeaderStyle(
				pterm.NewStyle(pterm.Bold),
			)
	}

	contents, err := printer.Srender()
	if err != nil {
		return fmt.Errorf("rendering table: %w", err)
	}

	if _, err := fmt.Fprintln(t.cfg.Out, contents); err != nil {
		return fmt.Errorf("flusing writer: %w", err)
	}

	return nil
}

type TableConfig struct {
	Out        io.Writer
	Columns    []string
	HFormatter HeaderFormatter
	NoColor    bool
	NoHeaders  bool
	PagerBin   string
}

func (c *TableConfig) Option(opts ...TableOption) {
	for _, opt := range opts {
		opt.ConfigureTable(c)
	}
}

func (c *TableConfig) Default() {
	if c.Out == nil {
		c.Out = os.Stdout
	}

	if c.HFormatter == nil {
		c.HFormatter = UpperSnake
	}
}

type TableOption interface {
	ConfigureTable(*TableConfig)
}

type WithOutput struct{ w io.Writer }

func (wo WithOutput) ConfigureTable(c *TableConfig) {
	c.Out = wo.w
}

type WithColumns string

func (wc WithColumns) ConfigureTable(c *TableConfig) {
	c.Columns = strings.Split(string(wc), ",")
}

type WithHeaderFormatter HeaderFormatter

func (wh WithHeaderFormatter) ConfigureTable(c *TableConfig) {
	c.HFormatter = HeaderFormatter(wh)
}

type WithNoColor bool

func (wn WithNoColor) ConfigureTable(c *TableConfig) {
	c.NoColor = bool(wn)
}

type WithNoHeaders bool

func (wn WithNoHeaders) ConfigureTable(c *TableConfig) {
	c.NoHeaders = bool(wn)
}

type WithPager string

func (wp WithPager) ConfigureTable(c *TableConfig) {
	c.PagerBin = string(wp)
}

func NewRow(data map[string]interface{}) Row {
	row := make(Row)

	for name, val := range data {
		row.AddField(name, val)
	}

	return row
}

type Row map[string]interface{}

func (r Row) AddField(name string, val interface{}) { r[normalize(name)] = val }

func (r Row) ValueString(name string) string {
	if val, ok := r[normalize(name)]; ok {
		return fmt.Sprint(val)
	}

	return ""
}

type RowModifier func(Row) Row

func WithAdditionalFields(fields map[string]interface{}) RowModifier {
	return func(r Row) Row {
		row := NewRow(r)

		for name, val := range fields {
			row.AddField(name, val)
		}

		return row
	}
}

type HeaderFormatter func(string) string

func UpperSnake(header string) string {
	trimmed := strings.TrimSpace(header)
	snaked := strings.Join(strings.Fields(trimmed), "_")

	return strings.ToUpper(snaked)
}

func NewPager(bin string, out io.Writer) (*Pager, error) {
	binPath, err := exec.LookPath(bin)
	if err != nil {
		return nil, fmt.Errorf("looking up pager binary: %w", err)
	}

	pipeOut, pipeIn, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("creating pipe: %w", err)
	}

	cmd := exec.Command(binPath)
	cmd.Stdin = pipeOut
	cmd.Stdout = out

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("running pager command: %w", err)
	}

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	return &Pager{
		pIn:  pipeIn,
		pOut: pipeOut,
		done: done,
	}, nil
}

type Pager struct {
	pIn  io.WriteCloser
	pOut io.ReadCloser
	done chan error
}

func (p *Pager) Write(b []byte) (n int, err error) {
	return p.pIn.Write(b)
}

func (p *Pager) Close() error {
	p.pIn.Close()
	p.pOut.Close()
	// wait for pager process to stop
	return fmt.Errorf("waiting for pager: %w", <-p.done)
}

func normalize(s string) string {
	trimmed := strings.TrimSpace(s)
	snaked := strings.Join(strings.Fields(trimmed), "_")

	return strings.ToLower(snaked)
}
