package stepflow

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Wizard is the entry point for building and running a TUI wizard.
//
// Usage:
//
//	result, err := stepflow.New().
//	    WithTheme(stepflow.DefaultTheme()).
//	    WithSteps(
//	        stepflow.Text("name", "Project name").Default("my-project"),
//	        stepflow.Confirm("guidelines", "Add guidelines?"),
//	        stepflow.List("agents", "Select agents").
//	            Items(stepflow.Item("Claude Code", ".claude/skills")).
//	            MultiSelect(true),
//	    ).
//	    Run()
type Wizard struct {
	steps []Step
	theme Theme
}

// New creates a new Wizard with sensible defaults.
func New() *Wizard {
	return &Wizard{
		theme: DefaultTheme(),
	}
}

// WithSteps sets the ordered list of steps for the wizard.
func (w *Wizard) WithSteps(steps ...Step) *Wizard {
	w.steps = steps
	return w
}

// WithTheme sets a custom color theme.
func (w *Wizard) WithTheme(t Theme) *Wizard {
	w.theme = t
	return w
}

// Run starts the interactive TUI and blocks until the user completes or
// cancels all steps. Returns the collected answers as a Result (map[string]string).
// Returns ErrCancelled if the user pressed ctrl+c before finishing.
func (w *Wizard) Run() (Result, error) {
	if len(w.steps) == 0 {
		return Result{}, errors.New("tui: no steps configured")
	}

	st := newStyles(w.theme)
	m := newWizardModel(w.steps, st)

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		return nil, err
	}

	wm, ok := final.(wizardModel)
	if !ok || !wm.done || wm.result == nil {
		return nil, ErrCancelled
	}
	return wm.result, nil
}

// ErrCancelled is returned when the user quits before completing all steps.
var ErrCancelled = errors.New("tui: wizard cancelled by user")
