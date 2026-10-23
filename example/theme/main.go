package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	// Build a cohesive Cyberpunk palette with 3 semantic colors:
	// Primary: Cyan, Accent: Purple, Muted: Slate
	cyberpunk := stepflow.NewTheme(
		lipgloss.Color("#06b6d4"), // Primary (Cyan markers, question labels, search slash)
		lipgloss.Color("#a855f7"), // Accent (Purple focus bars, selected dots, pills)
		lipgloss.Color("#64748b"), // Muted (Slate pipes, unselected circles, meta descriptions, hints)
	)

	wizard := stepflow.New().
		WithTheme(cyberpunk).
		WithAltScreen(true).  // Render in alternate terminal buffer (clean screen on exit)
		WithDoneScreen(true). // Keep completion screen visible before exiting
		WithSteps(
			stepflow.Info("brand", "Cyberpunk CLI Theme Demo").
				Body("This example uses a custom Cyan/Purple theme in an alternate terminal screen buffer."),
			stepflow.Text("tagline", "CLI Tagline").
				Default("Next-gen terminal experience"),
			stepflow.List("flavor", "Select color scheme variant").
				Items(
					stepflow.Item("Neon City", "High contrast cyan and magenta"),
					stepflow.Item("Night Drive", "Deep navy and amber highlights"),
					stepflow.Item("Matrix Rain", "Phosphor green on terminal black"),
				).
				PreSelect("Neon City"),
			stepflow.Confirm("save", "Apply this theme?").
				Default("Yes"),
		)

	result, err := wizard.Run()
	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Alt screen exits cleanly back to caller terminal
	fmt.Println("Custom theme applied successfully!")
	fmt.Printf("Tagline: %s | Flavor: %s | Saved: %t\n",
		result.Get("tagline"), result.Get("flavor"), result.Bool("save"))
}
