# Change these variables as necessary.
site_dir = site
addr = :8080
base = /loom/
dist = dist

# Live-reload watcher. The path is the repo root rather than site/, so editing
# a component rebuilds the page that demonstrates it — and because the site's
# reference sections are parsed from the library's source at startup, a doc
# comment would otherwise stay stale until the server was restarted by hand.
#
# Absolute on purpose. The templ CLI belongs to the site module, so it runs
# under `go -C site`, which puts the process's working directory in site/; a
# relative path would quietly watch that subtree instead of the repo, and only
# the site's own files would trigger a rebuild.
watch_path = $(CURDIR)
# templ's default pattern covers .go, .templ and _templ.txt. The site also has
# hand-written JS in static/ and the Tailwind source in css/.
watch_pattern = (.+\.go$$)|(.+\.templ$$)|(.+_templ\.txt$$)|(.+\.js$$)|(.+\.css$$)
# Everything the build writes back into the tree. Without these the watcher
# would see its own output and rebuild in a loop.
ignore_pattern = (_templ\.go$$)|(static/styles\.css$$)|(css/input\.css$$)|(/dist/)|(/\.git/)

# The templ and spark CLIs are tools of the site module, not the library —
# keeping them out of the root go.mod is what stops a published consumer of
# loom from inheriting them. `go -C` runs them in the module that declares them.
site_go = go -C $(site_dir)

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## confirm: prompt before continuing
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## no-dirty: fail if the working tree has uncommitted changes
.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)" || (echo 'working tree is dirty'; exit 1)

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all code and tidy both modules
.PHONY: tidy
tidy:
	go mod tidy -v
	$(site_go) mod tidy -v
	gofmt -w .

## audit: run quality control checks over both modules
# Generates first: without the *_templ.go files the site package imports none
# of the components, and `go mod tidy -diff` reports the whole dependency set
# as removable.
.PHONY: audit
audit: site/generate
	go mod tidy -diff
	go mod verify
	@test -z "$$(gofmt -l .)" || (echo 'unformatted files, run make tidy:'; gofmt -l .; exit 1)
	go vet ./...
	$(site_go) mod tidy -diff
	$(site_go) vet ./...
	$(MAKE) test

# lint is deliberately not part of audit: it currently reports pre-existing
# findings (see `make lint`), and audit is meant to stay green so a failure
# means something new broke.
## lint: run golangci-lint over both modules
.PHONY: lint
lint:
	golangci-lint run ./...
	cd $(site_dir) && golangci-lint run ./...

## test: run the library and site test suites
.PHONY: test
test: site/generate
	go test -race -buildvcs ./...
	$(site_go) test -race -buildvcs ./...

## test/cover: run the tests and open the coverage profile
.PHONY: test/cover
test/cover: site/generate
	go test -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## site/generate: generate Go code from .templ files
.PHONY: site/generate
site/generate:
	$(site_go) tool templ generate

## site/css: regenerate the Tailwind entry file and compile styles.css
.PHONY: site/css
site/css:
	go run ./cmd/css -o $(site_dir)/css/input.css
	cat $(site_dir)/css/site.css >> $(site_dir)/css/input.css
	tailwindcss -i $(site_dir)/css/input.css -o $(site_dir)/static/styles.css --minify

## site/run: run the site locally on $(addr)
.PHONY: site/run
site/run: site/generate site/css
	$(site_go) run . serve -addr $(addr)

## site/run/live: run the site with live reload (the watcher owns regeneration)
.PHONY: site/run/live
site/run/live:
	$(site_go) tool templ generate --watch \
		-path "$(watch_path)" \
		-watch-pattern '$(watch_pattern)' \
		-ignore-pattern '$(ignore_pattern)' \
		--proxy="http://127.0.0.1$(addr)" \
		--cmd="$(MAKE) -C $(CURDIR) site/run/live/server" \
		--open-browser=false -v

# site/run/live/server is the watcher's --cmd, invoked with -C because templ
# runs the command from the watched path. Build CSS and run the server, but do
# NOT run `templ generate`. In watch mode the watcher writes dev-mode text files
# under $TMPDIR and sets TEMPL_DEV_MODE=true for this command; a nested
# `templ generate` (as site/run does) deletes those files on exit, so the
# server would then fail with "templ: failed to render template".
.PHONY: site/run/live/server
site/run/live/server: site/css
	$(site_go) run . serve -addr $(addr)

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## site/build: render the site to $(dist) with base path $(base)
.PHONY: site/build
site/build: site/generate site/css
	$(site_go) run . build -o $(dist) -base $(base)

## site/preview: serve $(dist) under $(base) like GitHub Pages would
.PHONY: site/preview
site/preview: site/build
	mkdir -p /tmp/loom-preview && ln -sfn $(CURDIR)/$(site_dir)/$(dist) /tmp/loom-preview/loom
	@echo 'open http://localhost:8000/loom/'
	$(site_go) tool spark -port 8000 /tmp/loom-preview

## site/clean: remove generated artifacts
.PHONY: site/clean
site/clean:
	rm -rf $(site_dir)/$(dist) $(site_dir)/css/input.css $(site_dir)/static/styles.css $(site_dir)/pages/*_templ.go
