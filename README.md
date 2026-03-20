# stepflow

A minimal, themeable TUI wizard library for CLI onboarding flows in Go.  
Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss).

```
◆  Project name
│  agentsync
│
◆  Add guidelines?
│  Yes
│
◆  Fetching latest agentsync release
│
◆  Which agents do you want to install to?

── Universal (.agents/skills) (always included) ──────────
  •  Codex
  •  Cursor

── Additional ────────────────────────────────────────────
  / claude█

 ▌  ●  Claude Code       .claude/skills
    ○  Junie              .junie/skills

  [ Claude Code ]
```

---

## Install

```bash
go get github.com/f24aalam/stepflow
```

---

## Quick start

```go
import stepflow "github.com/f24aalam/stepflow/pkg"

result, err := stepflow.New().
    WithTheme(stepflow.DefaultTheme()).
    WithSteps(
        stepflow.Text("name", "Project name").Default("my-project"),
        stepflow.Confirm("guidelines", "Add guidelines?"),
        stepflow.Load("version", "Fetching latest release").
            Run(func(status chan<- string) (string, error) {
                status <- "contacting registry…"
                v, err := fetchVersion()
                return v, err
            }),
        stepflow.List("agents", "Select agents").
            Items(
                stepflow.Item("Claude Code", ".claude/skills"),
                stepflow.Item("Cursor",      ".cursor/skills"),
            ).
            MultiSelect(true),
    ).
    Run()

if err != nil { ... }

fmt.Println(result.Get("name"))        // "my-project"
fmt.Println(result.Bool("guidelines")) // true / false
fmt.Println(result.Get("version"))     // "v1.2.3"
fmt.Println(result.Get("agents"))      // "Claude Code, Cursor"
```

---

## Step types

### `Text` — free-text input

```go
stepflow.Text("key", "Question label").
    Placeholder("hint text").
    Default("fallback if empty")
```

### `Confirm` — Yes / No selector

```go
stepflow.Confirm("key", "Question label").
    Default("No") // pre-selects No; default is Yes
```

### `Load` — spinner with live status updates

Runs a background function while showing an animated spinner. The status
message updates live as the work progresses. On success it auto-advances to
the next step; on error it halts the wizard with a red `✗`.

```go
stepflow.Load("key", "Fetching remote config").
    Run(func(status chan<- string) (string, error) {
        status <- "resolving host…"
        time.Sleep(500 * time.Millisecond)

        status <- "downloading…"
        data, err := http.Get("https://example.com/config.json")
        if err != nil {
            return "", err
        }

        status <- "parsing…"
        return parseConfig(data)
    })
```

The return value is stored in `Result` under the step's key. Send as many
status strings as you like — the spinner drains the channel on every tick so
sends never block.

**What it looks like while running:**

```
◆  Fetching remote config
  ⠹  downloading…
```

**On success** — collapses into history and moves on:

```
◆  Fetching remote config
│  v0.4.2
│
```

**On error** — freezes with a red ✗ and the wizard returns an error:

```
◆  Fetching remote config
  ✗  dial tcp: connection refused
```

### `List` — searchable, scrollable item picker

```go
stepflow.List("key", "Question label").
    Items(
        stepflow.Item("Display name", "optional meta/path"),
        stepflow.Item("Another item"),
    ).
    Pinned("Section title", "note", "Item A", "Item B"). // read-only top section
    MultiSelect(true).   // space to toggle; false = single select
    VisibleRows(5).      // how many rows to show at once
    PreSelect("Item A")  // pre-check items by label
```

---

## Result

`Run()` returns `(Result, error)`.

| Method | Returns |
|---|---|
| `result.Get("key")` | Raw string answer |
| `result.Bool("key")` | `true` if answer is `"Yes"` |
| `result["key"]` | Direct map access |

List answers are returned as a comma-separated string: `"Claude Code, Cursor, Junie"`

---

## Error handling

```go
result, err := stepflow.New().WithSteps(...).Run()
if err != nil {
    if errors.Is(err, stepflow.ErrCancelled) {
        // user pressed ctrl+c
        os.Exit(0)
    }
    // a Load step's work func returned an error
    fmt.Fprintf(os.Stderr, "wizard failed: %v\n", err)
    os.Exit(1)
}
```

---

## Customisation

### Theme

```go
theme := stepflow.DefaultTheme()
theme.MarkerColor = lipgloss.Color("#ff79c6")

stepflow.New().WithTheme(theme).WithSteps(...).Run()
```

Full list of theme fields in [`theme.go`](./stepflow/theme.go).

## Keyboard controls

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `space` | Toggle selection (List multi-select) |
| `enter` | Confirm / advance |
| `backspace` | Delete search character (List search) |
| `ctrl+c` | Cancel — returns `ErrCancelled` |

> Loading steps ignore all key input while the work function is running.

---

## File structure

```
stepflow/
├── go.mod
├── README.md
├── pkg/
│   ├── wizard.go        — New(), WithSteps(), WithTheme(), Run()
│   ├── step.go          — Step + MessageStep interfaces
│   ├── step_text.go     — Text step (free-text input)
│   ├── step_confirm.go  — Confirm step (Yes / No)
│   ├── step_list.go     — List step (searchable multi/single select)
│   ├── step_load.go     — Load step (spinner + background WorkFunc)
│   ├── model.go         — bubbletea model, step progression, message routing
│   ├── theme.go         — Theme struct + DefaultTheme()
│   ├── styles.go        — lipgloss styles compiled from a Theme
│   └── result.go        — Result map + .Get() / .Bool()
└── example/
    └── main.go
```

---

## Run the example

```bash
cd example
go run main.go
```
