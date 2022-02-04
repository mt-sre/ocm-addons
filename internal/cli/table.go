package cli

import (
	"context"
	"io"

	"github.com/openshift-online/ocm-cli/pkg/output"
)

// NewTable returns a Table object which is used to format tabular output
// from the requested data for a particular cli command.
func NewTable(ctx context.Context, sess Session, options ...TableOption) (*Table, error) {
	var table Table

	for _, opt := range options {
		err := opt(&table)
		if err != nil {
			return nil, err
		}
	}

	var err error

	table.printer, err = output.NewPrinter().
		Writer(table.writer).
		Pager(sess.Config().Pager()).
		Build(ctx)
	if err != nil {
		return nil, err
	}

	table.Table, err = table.printer.NewTable().
		Name(table.name).
		Columns(table.columns).
		Build(ctx)
	if err != nil {
		return nil, err
	}

	return &table, err
}

type Table struct {
	columns   string
	name      string
	noHeaders bool
	writer    io.Writer
	printer   *output.Printer
	*output.Table
}

// WriteHeaders writes the table column headers when noHeaders is false.
func (t *Table) WriteHeaders() error {
	if t.noHeaders {
		return nil
	}

	return t.Table.WriteHeaders()
}

// Close flushes the output stream and releases any resources in use.
func (t *Table) Close() error {
	if err := t.printer.Close(); err != nil {
		return err
	}

	if err := t.Table.Close(); err != nil {
		return err
	}

	return nil
}

type TableOption func(*Table) error

// TableColumns selects the columns to be written by the table.
func TableColumns(columns string) TableOption {
	return func(t *Table) error {
		t.columns = columns

		return nil
	}
}

// TableName selects the name of the table file to reference for
// this table instance.
func TableName(name string) TableOption {
	return func(t *Table) error {
		t.name = name

		return nil
	}
}

// TableNoHeaders toggles whether this table will print headers.
func TableNoHeaders(noHeaders bool) TableOption {
	return func(t *Table) error {
		t.noHeaders = noHeaders

		return nil
	}
}

// TableWriter sets the writer instance where output will be written.
func TableWriter(writer io.Writer) TableOption {
	return func(t *Table) error {
		t.writer = writer

		return nil
	}
}
