// The showcase renders every loom component for visual verification.
//
//	make dev   # generate templ + build CSS + run on :8080
package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"

	"github.com/pietjan/loom"
	"github.com/pietjan/loom/showcase/templates"
)

func main() {
	log.Println("showcase: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler()))
}

// handler builds the full showcase app; the chromedp smoke test mounts it
// on an httptest server.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/{$}", templ.Handler(templates.Index()))
	return loom.Middleware(mux)
}
