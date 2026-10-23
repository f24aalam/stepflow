# Stepflow

A minimal, themeable TUI wizard library for Go CLI onboarding flows. Powered by Bubble Tea and Lip Gloss.

## Installation

```sh
go get github.com/f24aalam/stepflow
```

## Quick start

```go
import stepflow "github.com/f24aalam/stepflow/pkg"

result, err := stepflow.New().WithSteps(
    stepflow.Text("name", "Project name").Default("my-project"),
    stepflow.Confirm("git", "Initialize Git?"),
    stepflow.List("license", "Choose license").Items(stepflow.Item("MIT"), stepflow.Item("Apache-2.0")),
).Run()
if err != nil { /* handle cancellation or error */ }
fmt.Println(result.Get("name"), result.Bool("git"))
```

## Steps

- `Text(key, question)` with `.Default()` and `.Placeholder()`
- `Confirm(key, question)` with `.Default("No")`
- `List(key, question)` with `.Items()`, `.MultiSelect(true)`, `.VisibleRows(n)`, `.PreSelect()` and `.Pinned()`
- `Load(key, question).Run(func(chan<- string) (string,error))` for background work; `.RunContext(func(context.Context, chan<- string)(string,error))` for cancelable work
- `Info(key, question).Body(text)`; `Success(key, question).Body(text)`; `Error(key, question).Body(text)`
- `.WithNext(func(Result) []Step)` on each built-in step to choose the remainder dynamically

Multi-select results are comma-separated strings for compatibility. For uncomplicated labels, `result.Strings(key)` splits this value. Labels containing commas are ambiguous; do not treat this method as a lossless serialization format.

`Run()` returns `(Result,error)`. `RunContext(ctx)` additionally accepts cancellation. `errors.Is(err, stepflow.ErrCancelled)` identifies interactive cancellation. Loading failures propagate their original error. Legacy `.Run(fn)` workers have no cancellation mechanism; use `.RunContext(fn)` when cancellation matters, and send status messages with a `select` that includes `ctx.Done()`.

Default terminal rendering uses the normal screen; opt in to the alternate screen with `.WithAltScreen(true)`. Colors are configured with `.WithTheme(stepflow.DefaultTheme())`. See [theme.go](pkg/theme.go).

List controls: arrows navigate, type to filter, Space toggles multi-select, Enter confirms, Esc clears filter. Ctrl+C cancels the wizard.

## Examples

- **Quick start**: `go run ./example`
- **Interactive showcase** (all 7 step types & widgets): `go run ./example/showcase`
- **Dynamic branching** (`.WithNext` conditional flow): `go run ./example/dynamic`
- **Async loader** (status channel & context cancellation): `go run ./example/loader`
- **Custom theming & alt screen** (`.WithTheme`, `.WithAltScreen`): `go run ./example/theme`

**Scope:** Stepflow intentionally provides built-in steps, not third-party custom steps: the rendering interface currently uses internal styles. A stable custom-step API is deferred until there is a compelling use case.
