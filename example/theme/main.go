package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	// Customize colors based on DefaultTheme()
	cyberpunk := stepflow.DefaultTheme()
	cyberpunk.MarkerColor = lipgloss.Color("#06b6d4")         // Cyan
	cyberpunk.PipeColor = lipgloss.Color("#0891b2")           // Dark Cyan
	cyberpunk.FocusBarColor = lipgloss.Color("#a855f7")       // Purple bar
	cyberpunk.FocusLabelColor = lipgloss.Color("#c084fc")     // Bright lavender
	cyberpunk.StepLabelColor = lipgloss.Color("#38bdf8")      // Sky blue
	cyberpunk.ActiveQuestionColor = lipgloss.Color("#f0fdf4") // Light text
	cyberpunk.PillItemColor = lipgloss.Color("#a855f7")       // Purple pills
	cyberpunk.DoneColor = lipgloss.Color("#22d3ee")           // Cyan checkmarks

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
