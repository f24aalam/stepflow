package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	result, err := stepflow.New().
		WithTheme(stepflow.DefaultTheme()).
		WithSteps(

			// ── Text: free-text input ─────────────────────────────────────────
			//
			// .Placeholder()  hint shown when empty
			// .Default()      value used if user hits enter without typing
			stepflow.Text("project_name", "Project name").
				Placeholder("my-awesome-cli").
				Default("my-awesome-cli"),

			// ── Confirm: Yes / No selector ────────────────────────────────────
			//
			// .Default("No")  pre-selects No (default is Yes)
			stepflow.Confirm("open_source", "Is this open source?"),

			stepflow.Confirm("add_ci", "Add GitHub Actions CI?").
				Default("No"),

			// ── Load: background task with live status updates ────────────────
			//
			// .Run(fn)  fn runs in a goroutine; send strings to the status
			//           channel to update the spinner message live.
			//           Return (resultString, error).
			//           On success → auto-advances to next step.
			//           On error   → halts wizard, Run() returns the error.
			stepflow.Load("go_version", "Checking latest Go version").
				Run(func(status chan<- string) (string, error) {
					status <- "contacting go.dev…"
					time.Sleep(600 * time.Millisecond)

					resp, err := http.Get("https://go.dev/VERSION?m=text")
					if err != nil {
						return "go1.22", nil // fallback
					}
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)

					// go.dev/VERSION returns "go1.x.y\ntime <unix>" — first line only
					version := strings.SplitN(strings.TrimSpace(string(body)), "\n", 2)[0]

					status <- "done"
					return version, nil
				}),

			// ── List: searchable, scrollable picker ───────────────────────────
			//
			// .Items(...)        selectable entries; Item("label", "optional meta")
			// .MultiSelect(true) space to toggle multiple; false = pick one
			// .VisibleRows(n)    how many rows show at once (default 5)
			// .PreSelect("x")    pre-check items by label
			// .Pinned(...)       read-only section pinned above the list
			stepflow.List("license", "Choose a license").
				Items(
					stepflow.Item("MIT"),
					stepflow.Item("Apache 2.0"),
					stepflow.Item("GPL 3.0"),
					stepflow.Item("BSD 2-Clause"),
					stepflow.Item("BSD 3-Clause"),
					stepflow.Item("MPL 2.0"),
					stepflow.Item("AGPL 3.0"),
					stepflow.Item("Unlicense"),
				).
				MultiSelect(false).
				VisibleRows(5).
				PreSelect("MIT"),

			// List with multi-select + pinned universal section
			stepflow.List("notify", "Notification channels").
				Pinned(
					"Always active", "built-in",
					"Terminal", "Log file",
				).
				Items(
					stepflow.Item("Slack", "#builds channel"),
					stepflow.Item("Email", "team@example.com"),
					stepflow.Item("PagerDuty", "on-call rotation"),
					stepflow.Item("Discord", "#ci-alerts"),
				).
				MultiSelect(true).
				VisibleRows(4),
		).
		Run()

	// ── Handle errors ─────────────────────────────────────────────────────────

	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) {
			fmt.Println("\ncancelled.")
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "wizard error: %v\n", err)
		os.Exit(1)
	}

	// ── Use the result ────────────────────────────────────────────────────────
	//
	// result.Get("key")  → raw string answer
	// result.Bool("key") → true if answer is "Yes"
	// result["key"]      → direct map access

	fmt.Println()
	fmt.Printf("  Project    : %s\n", result.Get("project_name"))
	fmt.Printf("  Open source: %s\n", result.Get("open_source"))
	fmt.Printf("  CI         : %s\n", result.Get("add_ci"))
	fmt.Printf("  Go version : %s\n", result.Get("go_version"))
	fmt.Printf("  License    : %s\n", result.Get("license"))
	fmt.Printf("  Notify     : %s\n", result.Get("notify"))

	fmt.Println()
	if result.Bool("open_source") {
		fmt.Println("  → will create LICENSE file")
	}
	if result.Bool("add_ci") {
		fmt.Println("  → will scaffold .github/workflows/ci.yml")
	}
}
