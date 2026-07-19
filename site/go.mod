module github.com/pietjan/loom/site

go 1.26.4

require (
	github.com/a-h/templ v0.3.1020
	github.com/pietjan/loom v0.0.0
)

require (
	github.com/Oudwins/tailwind-merge-go v0.2.1 // indirect
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/alecthomas/chroma/v2 v2.27.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cli/browser v1.3.0 // indirect
	github.com/dlclark/regexp2/v2 v2.2.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/rif/spark v1.9.1 // indirect
	github.com/yuin/goldmark v1.8.4 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
)

replace github.com/pietjan/loom => ../

tool (
	github.com/a-h/templ/cmd/templ
	github.com/rif/spark
)
