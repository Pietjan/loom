// The site command serves the loom project website for development and
// renders it out as static HTML for GitHub Pages.
//
//	make dev            # generate templ + build CSS + serve on :8080
//	make build/static   # render to dist/ with base path /loom/
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-h/templ"

	"github.com/pietjan/loom"
	"github.com/pietjan/loom/site/pages"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		addr := fs.String("addr", ":8080", "listen address")
		fs.Parse(os.Args[2:])
		log.Printf("site: http://localhost%s", *addr)
		log.Fatal(http.ListenAndServe(*addr, handler()))
	case "build":
		fs := flag.NewFlagSet("build", flag.ExitOnError)
		out := fs.String("o", "dist", "output directory")
		base := fs.String("base", "", "base path prefix, e.g. /loom/")
		fs.Parse(os.Args[2:])
		if err := build(*out, *base); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: site serve [-addr :8080] | site build [-o dist] [-base /loom/]")
	os.Exit(2)
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/{$}", templ.Handler(pages.Index()))
	mux.HandleFunc("/components/{slug}/{$}", func(w http.ResponseWriter, r *http.Request) {
		p, ok := pages.Find(r.PathValue("slug"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		templ.Handler(p.Body()).ServeHTTP(w, r)
	})
	return loom.Middleware(mux)
}

func build(out, base string) error {
	render := func(rel string, c templ.Component) error {
		// A fresh loom context per page keeps element IDs deterministic and
		// per-page-scoped, matching what the middleware does in serve mode.
		ctx := pages.WithBase(loom.NewContext(context.Background()), base)
		path := filepath.Join(out, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return c.Render(ctx, f)
	}
	if err := render("index.html", pages.Index()); err != nil {
		return err
	}
	for _, p := range pages.All() {
		if err := render(filepath.Join("components", p.Slug, "index.html"), p.Body()); err != nil {
			return err
		}
	}
	// Copy the whole static tree (compiled CSS plus the self-hosted fonts).
	if err := copyDir("static", filepath.Join(out, "static")); err != nil {
		return err
	}
	// .nojekyll stops GitHub Pages from running the artifact through Jekyll.
	if err := os.WriteFile(filepath.Join(out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}
	log.Printf("site: rendered %d pages to %s", len(pages.All())+1, out)
	return nil
}

// copyDir recursively copies the src tree into dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
