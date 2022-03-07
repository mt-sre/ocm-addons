package output

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

const tabWidth = 8

func NewTable(opts ...TableOption) (*Table, error) {
	var table Table

	table.cfg.Option(opts...)
	table.cfg.Default()

	out := table.cfg.Out

	if table.cfg.PagerBin != "" {
		var err error

		table.pager, err = NewPager(table.cfg.PagerBin, table.cfg.Out)
		if err != nil {
			return nil, err
		}

		out = table.pager
	}

	table.writer = tabwriter.NewWriter(out, 0, tabWidth, 1, '\t', 0)

	if !table.cfg.NoHeaders {
		if err := table.writeHeaders(); err != nil {
			return nil, err
		}
	}

	return &table, nil
}

type Table struct {
	cfg    TableConfig
	pager  *Pager
	writer *tabwriter.Writer
}

func (t *Table) writeHeaders() error {
	row := strings.Join(t.formattedHeaders(), "\t")

	_, err := t.writer.Write([]byte(row + "\t\n"))

	return err
}

func (t *Table) formattedHeaders() []string {
	headers := make([]string, 0, len(t.cfg.Columns))

	for _, c := range t.cfg.Columns {
		headers = append(headers, t.cfg.HFormatter(c))
	}

	return headers
}

func (t *Table) Write(r ToRower, mods ...RowModifier) error {
	row := r.ToRow()

	for _, mod := range mods {
		row = mod(row)
	}

	processedRow := t.cfg.Selector(row)

	_, err := t.writer.Write([]byte(processedRow.Format()))

	return err
}

func (t *Table) Flush() error {
	var finalErr error

	if err := t.writer.Flush(); err != nil {
		finalErr = fmt.Errorf("flusing writer: %w", finalErr)
	}

	if t.pager != nil {
		if err := t.pager.Close(); err != nil {
			finalErr = fmt.Errorf("closing pager: %w", finalErr)
		}
	}

	return finalErr
}

type TableConfig struct {
	Out        io.Writer
	Columns    []string
	Selector   FieldSelector
	HFormatter HeaderFormatter
	NoHeaders  bool
	PagerBin   string
}

func (c *TableConfig) Option(opts ...TableOption) {
	for _, opt := range opts {
		opt.ApplyToTableConfig(c)
	}
}

func (c *TableConfig) Default() {
	if c.Out == nil {
		c.Out = os.Stdout
	}

	if c.Selector == nil {
		c.Selector = ByName(c.Columns...)
	}

	if c.HFormatter == nil {
		c.HFormatter = UpperSnake
	}
}

type TableOption interface {
	ApplyToTableConfig(*TableConfig)
}

type WithOutput struct{ w io.Writer }

func (wo WithOutput) ApplyToTableConfig(c *TableConfig) {
	c.Out = wo.w
}

type WithColumns string

func (wc WithColumns) ApplyToTableConfig(c *TableConfig) {
	c.Columns = strings.Split(string(wc), ",")
}

type WithFieldSelector FieldSelector

func (wf WithFieldSelector) ApplyToTableConfig(c *TableConfig) {
	c.Selector = FieldSelector(wf)
}

type WithHeaderFormatter HeaderFormatter

func (wh WithHeaderFormatter) ApplyToTableConfig(c *TableConfig) {
	c.HFormatter = HeaderFormatter(wh)
}

type WithNoHeaders bool

func (wn WithNoHeaders) ApplyToTableConfig(c *TableConfig) {
	c.NoHeaders = bool(wn)
}

type WithPager string

func (wp WithPager) ApplyToTableConfig(c *TableConfig) {
	c.PagerBin = string(wp)
}

type ToRower interface {
	ToRow() Row
}

type Row []Field

func (r Row) Format() string {
	fields := make([]string, 0, len(r))

	for _, f := range r {
		fields = append(fields, f.ValueString())
	}

	res := strings.Join(fields, "\t")

	return res + "\t\n"
}

func (r Row) GetField(name string) Field {
	for _, f := range r {
		if normalizedEquals(f.Name, name) {
			return f
		}
	}

	return Field{}
}

func normalizedEquals(s1, s2 string) bool {
	return normalizeString(s1) == normalizeString(s2)
}

func normalizeString(s string) string {
	trimmed := strings.TrimSpace(s)
	snaked := strings.Join(strings.Fields(trimmed), "_")

	return strings.ToLower(snaked)
}

type RowModifier func(Row) Row

func WithAdditionalFields(fs ...Field) RowModifier {
	return func(r Row) Row {
		return append(r, fs...)
	}
}

type Field struct {
	Name  string
	Value interface{}
}

func (f *Field) ValueString() string {
	return fmt.Sprint(f.Value)
}

type FieldSelector func(Row) Row

func ByName(fieldNames ...string) FieldSelector {
	return func(r Row) Row {
		res := make(Row, 0, len(r))

		for _, name := range fieldNames {
			res = append(res, r.GetField(name))
		}

		return res
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
		return nil, err
	}

	pipeOut, pipeIn, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binPath)
	cmd.Stdin = pipeOut
	cmd.Stdout = out

	if err := cmd.Start(); err != nil {
		return nil, err
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
	return <-p.done
}
