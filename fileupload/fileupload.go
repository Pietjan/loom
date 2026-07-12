// Package fileupload renders a styled native file input:
//
//	@fileupload.New(fileupload.Name("avatar"), fileupload.Accept("image/*"))
//	@fileupload.New(fileupload.Name("docs"), fileupload.Multiple())
//
// The browser draws the "choose file" button and the selected file name;
// loom styles the button (::file-selector-button in cmd/css/loom.css).
// Showing a rich drop-zone preview would need JS — out of scope; the
// native control shows the file name for free. Inside a field it adopts
// the field's id and disabled state.
package fileupload

import (
	"context"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/internal/scope"
)

// Config holds file-upload options.
type Config struct {
	opts.Common
}

// Option configures a file upload.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Name sets the form field name.
func Name(name string) Option { return Attr("name", name) }

// Accept sets the accepted file types (e.g. "image/*", ".pdf,.docx").
func Accept(types string) Option { return Attr("accept", types) }

// Multiple allows selecting more than one file.
func Multiple() Option { return Attr("multiple", "") }

// Disabled disables the input.
func Disabled() Option { return Attr("disabled", "") }

// New renders a file input as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the <input type=file> node.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{}
	for _, opt := range options {
		opt(&cfg)
	}

	n := dom.El(atom.Input, dom.Marker("file"), dom.Attr("type", "file"))
	if sc, ok := scope.From[field.Scope](ctx); ok {
		if sc.Required {
			dom.SetAttr(n, "required", "")
		}
		if sc.Disabled {
			dom.SetAttr(n, "disabled", "")
		}
	}
	cfg.Apply(n, classes())
	return n, nil
}
