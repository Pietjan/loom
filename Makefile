# Change these variables as necessary.
site_dir = site
addr = :8080
base = /loom/
dist = dist

# The CLIs live in their own module so that neither the published library nor
# the site inherits their dependency trees. -modfile runs them from that module
# without changing the working directory, which is what lets a tool act on the
# module it is invoked from.
tools = -modfile=tools/go.mod
site_tools = -modfile=../tools/go.mod

# Live-reload watcher. The path is the repo root rather than site/, so editing
# a component rebuilds the page that demonstrates it - and because the site's
# reference sections are parsed from the library's source at startup, a doc
# comment would otherwise stay stale until the server was restarted by hand.
watch_path = .
# templ's default pattern covers .go, .templ and _templ.txt. The site also has
# hand-written JS in pages/scripts/ (embedded into the pages, so editing one
# needs a Go rebuild) and the Tailwind source in css/.
watch_pattern = (.+\.go$$)|(.+\.templ$$)|(.+_templ\.txt$$)|(.+\.js$$)|(.+\.css$$)
# Everything the build writes back into the tree. Without these the watcher
# would see its own output and rebuild in a loop.
ignore_pattern = (_templ\.go$$)|(static/styles\.css$$)|(css/input\.css$$)|(/dist/)|(/\.git/)

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)"

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell go tool $(tools) gofumpt -l .)"
	go vet ./...
	go tool $(tools) golangci-lint run ./...
	go tool $(tools) govulncheck ./...
	cd $(site_dir) && go mod tidy -diff
	cd $(site_dir) && go vet ./...
	cd $(site_dir) && go tool $(site_tools) golangci-lint run ./...
	cd tools && go mod tidy -diff
	cd tools && go mod verify

## test: run all tests
# Generates first: without the *_templ.go files the site's pages package does
# not compile. Builds the CSS too, because the class-conflict test reads the
# compiled stylesheet and skips itself when it is missing - a test that opts
# out silently is worse than no test.
.PHONY: test
test: site/generate site/css
	go test -race -buildvcs ./...
	cd $(site_dir) && go test -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover: site/generate
	go test -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## upgradeable: list direct dependencies that have upgrades available
.PHONY: upgradeable
upgradeable:
	@go tool $(tools) go-mod-upgrade

## upgradeable/tools: list tool dependencies that have upgrades available
.PHONY: upgradeable/tools
upgradeable/tools:
	@cd tools && go tool -modfile=go.mod go-mod-upgrade

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## tidy: tidy modfiles, modernize and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	cd $(site_dir) && go mod tidy -v
	cd tools && go mod tidy -v
	go tool $(tools) modernize -test -fix ./...
	go tool $(tools) gofumpt -l -w .

## site/generate: generate Go code from .templ files
.PHONY: site/generate
site/generate:
	go tool $(tools) templ generate -path $(site_dir)

## site/css: regenerate the Tailwind entry file and compile styles.css
.PHONY: site/css
site/css:
	go run ./cmd/css -o $(site_dir)/css/input.css
	cat $(site_dir)/css/site.css >> $(site_dir)/css/input.css
	tailwindcss -i $(site_dir)/css/input.css -o $(site_dir)/static/styles.css --minify

## site/run: run the site locally on $(addr)
.PHONY: site/run
site/run: site/generate site/css
	cd $(site_dir) && go run . serve -addr $(addr)

## site/run/live: run the site with live reload (the watcher owns regeneration)
.PHONY: site/run/live
site/run/live:
	go tool $(tools) templ generate --watch \
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
	cd $(site_dir) && go run . serve -addr $(addr)

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## site/build: render the site to $(dist) with base path $(base)
.PHONY: site/build
site/build: site/generate site/css
	cd $(site_dir) && go run . build -o $(dist) -base $(base)

## site/preview: serve $(dist) under $(base) like GitHub Pages would
.PHONY: site/preview
site/preview: site/build
	mkdir -p /tmp/loom-preview && ln -sfn $(CURDIR)/$(site_dir)/$(dist) /tmp/loom-preview/loom
	@echo 'open http://localhost:8000/loom/'
	go tool $(tools) spark -port 8000 /tmp/loom-preview

## site/clean: remove generated artifacts
.PHONY: site/clean
site/clean:
	rm -rf $(site_dir)/$(dist) $(site_dir)/css/input.css $(site_dir)/static/styles.css $(site_dir)/pages/*_templ.go
