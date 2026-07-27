// Command themes writes every accent/base combination as an attribute-scoped
// override, to be appended to the entry file cmd/css generates:
//
//	go run github.com/pietjan/loom/cmd/css -o css/input.css
//	go run github.com/pietjan/loom/cmd/css/themes >> css/input.css
//
// The result is a stylesheet that switches theme from data-accent /
// data-base on <html>, instead of one compiled per combination. It is what
// loom's own site uses for its ?accent= and ?base= preview; cmd/css on its
// own still writes a single theme, which is what most projects want.
//
// It lives under cmd/css because that is the only place allowed to import
// the internal theme package the accent table lives in.
package main

import (
	"log"
	"os"

	"github.com/pietjan/loom/cmd/css/internal/theme"
)

func main() {
	css, err := theme.GenerateAll()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.Write(css); err != nil {
		log.Fatal(err)
	}
}
