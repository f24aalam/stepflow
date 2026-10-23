# Stepflow

**A minimal, themeable TUI wizard library for Go.**

Build beautiful, interactive CLI onboarding flows with a simple, fluent API. Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

Stepflow focuses on the essentials: **simple to use, easy to customize, and reliable by default.**

<!-- Add a screenshot captured from `go run ./example/showcase` here. -->
<!-- ![Stepflow terminal preview](docs/stepflow.png) -->

## Features

- **Simple API** — Build multi-step wizards with readable, chainable Go code.
- **Built-in steps** — Text input, confirmation, single/multi-select lists, loading, information, and success/error messages.
- **Dynamic flows** — Conditionally show steps based on previous answers using `WithNext()`.
- **Async operations** — Run background tasks with live status updates and context-aware cancellation.
- **Custom themes** — Customize colors without dealing with complex styling configurations.
- **Responsive terminal UI** — Compact output with wrapping for longer questions and answers.
- **Lightweight** — No configuration files, plugin systems, or unnecessary abstractions.

## Installation

Requires **Go 1.22 or later**.

```bash
go get github.com/f24aalam/stepflow
```

## Quick Start

```go
package main

import (
    "errors"
    "fmt"
    "log"

    stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
    result, err := stepflow.New().
        WithSteps(
            stepflow.Text("name", "Project name").
                Default("my-project"),

            stepflow.Confirm("git", "Initialize Git?"),

            stepflow.List("license", "Choose a license").
                Items(
                    stepflow.Item("MIT"),
                    stepflow.Item("Apache-2.0"),
                ),
        ).
        Run()

    if errors.Is(err, stepflow.ErrCancelled) {
        return
    }
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Project:", result.Get("name"))
    fmt.Println("Git:", result.Bool("git"))
    fmt.Println("License:", result.Get("license"))
}
```

The wizard collects user input step by step and returns the answers as a `Result`.

## Built-in Steps

| Step | Purpose |
|---|---|
| `Text` | Free-text input with placeholders and defaults |
| `Confirm` | Yes/No confirmation |
| `List` | Searchable single-select or multi-select |
| `Load` | Background tasks with live progress messages |
| `Info` | Display information and continue |
| `Success` | Display a success message |
| `Error` | Display an error message |

### Dynamic Steps

Create conditional flows with `WithNext()`:

```go
stepflow.Confirm("advanced", "Configure advanced options?").
    WithNext(func(r stepflow.Result) []stepflow.Step {
        if r.Bool("advanced") {
            return []stepflow.Step{
                stepflow.Text("region", "Deployment region").
                    Default("us-east-1"),
            }
        }
        return []stepflow.Step{}
    })
```

A non-nil slice replaces the remaining steps. Return `nil` to keep the existing sequence.

### Background Tasks

Run operations with live status updates:

```go
stepflow.Load("setup", "Preparing project").
    RunContext(func(ctx context.Context, status chan<- string) (string, error) {
        select {
        case status <- "Setting up...":
        case <-ctx.Done():
            return "", ctx.Err()
        }

        // Perform your actual setup work here,
        // checking ctx.Done() during long operations.

        return "Ready", nil
    })
```

Use `Wizard.RunContext(ctx)` to supply a cancellation context. Legacy `Load().Run(fn)` remains available, but cannot cooperatively cancel work through a context.

## Themes

Stepflow includes a default theme and supports custom colors.

```go
theme := stepflow.NewTheme(
    "#4ade80", // Primary
    "#22c55e", // Accent
    "#6b7280", // Muted
)

result, err := stepflow.New().
    WithTheme(theme).
    WithSteps(...).
    Run()
```

Normal terminal rendering is the default. Enable the alternate screen with `WithAltScreen(true)`.

## Examples

Explore the working examples included in the repository:

| Example | Run |
|---|---|
| Quick start | `go run ./example` |
| Full showcase | `go run ./example/showcase` |
| Dynamic branching | `go run ./example/dynamic` |
| Async loading | `go run ./example/loader` |
| Custom themes | `go run ./example/theme` |

## Keyboard Controls

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate options |
| `Space` | Toggle items in multi-select lists |
| `Enter` | Confirm and continue |
| `Backspace` | Delete a search character |
| `Esc` | Clear list search |
| `Ctrl+C` | Cancel wizard |

## API Notes

**Results:** `Run()` returns `(Result, error)`. Use `result.Get(key)` for strings, `result.Bool(key)` for confirmations, and `result.Strings(key)` for multi-select values.

**CSV compatibility:** Multi-select values use CSV-compatible encoding. Legacy unquoted comma-separated strings can be ambiguous, and CRLF sequences may be normalized to LF during parsing.

**Cancellation:** `RunContext(ctx)` supports context cancellation. Background workers must cooperate by checking their context; Go cannot forcibly stop an uncooperative goroutine.

**Wizard reuse:** Wizards can run multiple times and restore configured defaults between runs.

**Custom steps:** Stepflow currently supports its built-in step types. Third-party custom steps aren't part of the public extension API.

**Versioning:** `v0.1.0` is the first public release. The API may evolve before `v1.0.0`.

## Contributing

Contributions and bug reports are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Stepflow is open source under the [MIT License](LICENSE).

---

**The best implementation is the simplest one that reliably solves the real problem.**

— ChatGPT (OpenAI), from a Stepflow engineering discussion, 2026
