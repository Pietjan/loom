// Package table renders data tables:
//
//	@table.New() {
//		@table.Header() {
//			@table.Column() { Name }
//			@table.Column() { Status }
//		}
//		@table.Body() {
//			@table.Row() {
//				@table.Cell() { Ada }
//				@table.Cell() { @badge.New(badge.Green) { Active } }
//			}
//		}
//	}
//
// The wrapper scrolls horizontally on overflow. Fragment parsing happens
// against the real table ancestors, so <tr>/<td> survive — the exact
// parser trap the predecessor's AsNode fell into.
package table

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
)

// Config holds table options.
type Config struct {
	opts.Common
}

// Option configures a table part.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// New renders the scroll wrapper and <table>.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}

		tbl := dom.El(atom.Table, dom.Marker("table"))
		if err := render.Children(ctx, tbl); err != nil {
			return nil, err
		}
		dom.SetAttr(tbl, "class", tableClasses())

		wrap := dom.El(atom.Div, dom.Marker("table-wrapper"))
		wrap.AppendChild(tbl)
		cfg.Apply(wrap, wrapperClasses())
		return wrap, nil
	})
}

// Header renders <thead> with an implicit row around its Columns.
func Header(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		thead := dom.El(atom.Thead, dom.Marker("table-header"))
		tr := dom.El(atom.Tr)
		if err := render.Children(ctx, tr); err != nil {
			return nil, err
		}
		thead.AppendChild(tr)
		cfg.Apply(thead, headerClasses())
		return thead, nil
	})
}

// Body renders <tbody>.
func Body(options ...Option) templ.Component {
	return section(atom.Tbody, "table-body", bodyClasses(), options)
}

// Row renders <tr>. Inside Header its cells are Columns.
func Row(options ...Option) templ.Component {
	return section(atom.Tr, "table-row", rowClasses(), options)
}

// Column renders a <th> header cell.
func Column(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		th := dom.El(atom.Th, dom.Marker("table-column"), dom.Attr("scope", "col"))
		if err := render.Children(ctx, th); err != nil {
			return nil, err
		}
		cfg.Apply(th, columnClasses())
		return th, nil
	})
}

// Cell renders a <td>.
func Cell(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		td := dom.El(atom.Td, dom.Marker("table-cell"))
		if err := render.Children(ctx, td); err != nil {
			return nil, err
		}
		cfg.Apply(td, cellClasses())
		return td, nil
	})
}

func section(a atom.Atom, marker, classes string, options []Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		cfg := Config{}
		for _, opt := range options {
			opt(&cfg)
		}
		n := dom.El(a, dom.Marker(marker))
		if err := render.Children(ctx, n); err != nil {
			return nil, err
		}
		cfg.Apply(n, classes)
		return n, nil
	})
}
