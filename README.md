# loom

Pre-styled, accessible UI components for Go + [templ](https://templ.guide),
inspired by [Flux UI](https://fluxui.dev). Import a component, drop it in
your template, done:

```templ
import "github.com/pietjan/loom/button"

templ Page() {
	@button.New(button.Primary) { Save }
}
```

Interactivity comes from **web platform primitives, not JavaScript**:
modals are `<dialog>` driven by [invoker commands](https://developer.mozilla.org/en-US/docs/Web/API/Invoker_Commands_API)
(`commandfor`/`command`, Baseline 2025), dropdowns are popovers with CSS
anchor positioning, accordions and nav groups are `<details name>`, selects
are [customizable selects](https://developer.mozilla.org/en-US/docs/Learn_web_development/Extensions/Forms/Customizable_select),
and the responsive sidebar is a popover that CSS pins open on wide
viewports. Where a platform feature isn't Baseline yet, the CSS degrades
gracefully — never breaks.

## Setup

```sh
go get github.com/pietjan/loom
```

Generate your Tailwind entry file. It's self-contained: the Tailwind
import, an `@source` pointing at loom in the module cache (so the CLI sees
the class strings baked into the components), your theme variables, and
loom's structural CSS — all written into one file:

```sh
go run github.com/pietjan/loom/cmd/css -accent indigo -o assets/css/input.css
tailwindcss -i assets/css/input.css -o assets/static/styles.css
```

Re-run `cmd/css` after upgrading loom — the `@source` path is
version-pinned and the structural CSS is snapshotted into the file, so
both refresh together. A `//go:generate` line is a natural home for it.

Install the per-request ID middleware so generated element IDs are
deterministic and unique per page:

```go
http.ListenAndServe(":8080", loom.Middleware(mux))
```

## Components

**Foundation** — button, badge, icon (phosphor, generated constants),
heading, text, link, separator, card, callout, chart.

**Data display** — avatar (photo/initials, status, stacking), breadcrumbs,
progress (determinate + indeterminate), skeleton, pagination (link-based),
stat (KPI tile — compose a delta badge or sparkline as children),
timeline (composable indicators, statuses, horizontal mode), kanban
(board/column/card — static markup; reordering is your app's concern),
description (`<dl>` term/detail), kbd. All pure markup + CSS.

Charts are server-rendered SVG — line/area (straight or smooth), grouped
bars, and sparklines, with nice ticks, gridlines, a legend, and per-point
hover values shown in the ordinary `tooltip` component (an HTML overlay
positioned in percentages over the scaling SVG — reused, not reinvented,
still no JS):

```templ
@chart.New(
	chart.Title("Visitors per month"),
	chart.Labels("Jan", "Feb", "Mar", "Apr", "May", "Jun"),
	chart.Series("Visitors", []float64{120, 190, 170, 220, 300, 260}),
	chart.Series("Signups", []float64{40, 60, 55, 90, 120, 100}, chart.Colored(chart.Emerald)),
	chart.Area(), chart.Smooth(), chart.Legend(),
)
```

A synced crosshair cursor with a live multi-value tooltip would need JS —
out of scope by policy; per-point tooltips are the offering.

**Forms** — field, input, textarea, checkbox, radio, toggle, picker,
fieldset, slider (native range), fileupload (native file input), inputgroup
(input with leading/trailing addons). The field composite wires everything
by itself — including a control wrapped in an inputgroup:

```templ
@field.Root(field.Error(msg), field.Required()) {
	@field.Label() { Email }
	@input.New(input.Type("email"), input.Name("email"))
	@field.Description() { We never share it. }
}
```

renders label `for` ↔ control `id`, `aria-describedby` listing exactly the
parts that rendered, `aria-invalid` + error styling, mirrored
`required`/`disabled` — zero manual plumbing.

**Layout** — two application shells. A **sidebar** layout (left `sidebar`
+ vertical `navlist`, with the responsive popover trick) and a **header**
layout (`header` top bar + horizontal `navbar`/`navbar.Item`, `header.Main`
for content):

```templ
@header.New(header.Sticky()) {
	@link.New("/") { Acme Inc. }
	@navbar.New(navbar.Label("Main")) {
		@navbar.Item("/", navbar.Current()) { @icon.New(icon.Home, icon.Mini) Home }
		@navbar.Item("/inbox", navbar.Badge("12")) { @icon.New(icon.Inbox, icon.Mini) Inbox }
	}
	@header.Spacer()
	@button.New(button.Ghost) { Sign out }
}
```

**Overlays & navigation** — modal, dropdown, popover (free-form anchored
panel), tooltip, accordion, navlist, sidebar, navbar, header, table, tabs,
carousel (CSS scroll-snap), flash (server-rendered dismissible alert;
checkbox-hack dismiss + CSS auto-hide, no JS):

```templ
@modal.Root() {
	@modal.Trigger() { @button.New(button.Danger) { Delete } }
	@modal.Content() {
		@modal.Title() { Are you sure? }
		@modal.Close() { @button.New() { Cancel } }
	}
}
```

Root generates the pairing id; Trigger/Close stamp `command`/`commandfor`
on the button in their block. A trigger far from its dialog pairs by name:
`modal.Trigger(modal.For("confirm"))` ↔ `modal.Content(modal.Name("confirm"))`.
Missing pairing is a **render error**, never a dead button.

## Architecture

Components are pure Go: they build `*html.Node` trees
(`golang.org/x/net/html`) and implement `templ.Component`. Two composition
mechanisms, with strict rules:

1. **Context scopes** (state flowing down): a parent installs a typed
   scope before rendering its children; children adapt themselves
   (`field.Scope` carries ids + invalid/required/disabled; controls read
   it). Works because templ child blocks receive the render-time context —
   pinned by a test in `internal/render`.
2. **Post-passes** (sibling relations): after children materialize, the
   parent runs typed queries (`dom.FindShallow` — stops at other
   components' `data-ui` roots) and sets attributes only. Post-passes
   never restructure the tree.

Every component root carries `data-ui="<name>"` — the hook for structural
CSS, post-processing, tests, and your own CSS.

Styling is Tailwind class strings in Go (`style.go` per component,
complete literals only — the scanner must see them), merged with user
classes via tailwind-merge (user wins), sorted canonically for stable
output. Structural CSS that utilities can't express (popover fallbacks,
`::picker(select)`, sidebar media rules, dialog transitions) lives in
`cmd/css/loom.css` — embedded into the generated entry file by `cmd/css`,
keyed on `data-ui` markers, inside `@layer components` so your CSS always
overrides it.

### Tabs

`tabs.New`/`tabs.Section`: a `<details name>` disclosure group laid out
as tabs via `display: contents` + `::details-content` (Baseline 2025) —
all panels ship in the page and the platform switches them. Older
browsers fall back to a vertical accordion via `@supports`; announced as
disclosure (which it is), not ARIA tabs. For URL-addressable sections,
compose `navlist`-style links and render the active content
server-side.

### JavaScript policy

loom ships zero JavaScript — no exceptions. Patterns that honestly
require JS (the full ARIA tabs pattern with arrow-key roving,
combobox/autocomplete) are out of scope rather than faked.

## Development

```sh
go test ./...                    # unit + golden + contract tests
LOOM_UPDATE=1 go test ./...      # rewrite golden files

cd site
make run                          # templ generate + CSS build + serve :8080
make run/live                     # ...with live reload (templ watch + proxy)
```

The contract harness (`contract_test.go`) checks every composite render:
all id references (`for`, `aria-describedby`, `commandfor`, …) resolve to
an element in the same document, ids are unique, and every form control
has an accessible name. New composites must register themselves.

The import-graph test (`arch_test.go`) enforces layering: primitives
import only `internal/`; composites cooperate through scopes along
explicitly allowed edges — never by calling another component's `Node()`.
