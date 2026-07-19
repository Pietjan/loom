# Change these variables as necessary.
addr = :8080
base = /loom/
dist = dist

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## generate: generate Go code from .templ files
.PHONY: generate
generate:
	go tool templ generate

## css: regenerate the Tailwind entry file and compile styles.css
.PHONY: css
css:
	go run github.com/pietjan/loom/cmd/css -o css/input.css
	cat css/site.css >> css/input.css
	tailwindcss -i css/input.css -o static/styles.css --minify

## run: run the site locally on $(addr)
.PHONY: run
run: generate css
	go run . serve -addr $(addr)

## run/live: run the site with live reload (the watcher owns regeneration)
.PHONY: run/live
run/live:
	go tool templ generate --watch --proxy="http://127.0.0.1$(addr)" --cmd="make run/live/server" --open-browser=false -v

# run/live/server is the watcher's --cmd: build CSS and run the server, but do
# NOT run `templ generate`. In watch mode the watcher writes dev-mode text files
# under $TMPDIR and sets TEMPL_DEV_MODE=true for this command; a nested
# `templ generate` (as `make run` does) deletes those files on exit, so the
# server would then fail with "templ: failed to render template".
.PHONY: run/live/server
run/live/server: css
	go run . serve -addr $(addr)

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/static: render the site to $(dist) with base path $(base)
.PHONY: build/static
build/static: generate css
	go run . build -o $(dist) -base $(base)

## preview: serve $(dist) under $(base) like GitHub Pages would
.PHONY: preview
preview: build/static
	mkdir -p /tmp/loom-preview && ln -sfn $(CURDIR)/$(dist) /tmp/loom-preview/loom
	@echo 'open http://localhost:8000/loom/'
	go tool spark -port 8000 /tmp/loom-preview

## clean: remove generated artifacts
.PHONY: clean
clean:
	rm -rf $(dist) css/input.css
