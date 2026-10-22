package stepflow

import (
	"context"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrCancelled = errors.New("stepflow: wizard cancelled by user")

type Wizard struct {
	steps                  []Step
	theme                  Theme
	useAltScreen, showDone bool
}

func New() *Wizard                                    { return &Wizard{theme: DefaultTheme(), useAltScreen: false, showDone: true} }
func (w *Wizard) WithAltScreen(enabled bool) *Wizard  { w.useAltScreen = enabled; return w }
func (w *Wizard) WithDoneScreen(enabled bool) *Wizard { w.showDone = enabled; return w }
func (w *Wizard) WithSteps(steps ...Step) *Wizard     { w.steps = steps; return w }
func (w *Wizard) WithTheme(t Theme) *Wizard           { w.theme = t; return w }
func (w *Wizard) Run() (Result, error)                { return w.RunContext(context.Background()) }
func (w *Wizard) RunContext(ctx context.Context) (Result, error) {
	if err := validateSteps(w.steps); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, s := range w.steps {
		if l, ok := s.(*LoadStep); ok {
			l.ctx = ctx
		}
	}
	m := newWizardModel(w.steps, newStyles(w.theme), w.showDone)
	opts := []tea.ProgramOption{tea.WithContext(ctx)}
	if w.useAltScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	final, err := tea.NewProgram(m, opts...).Run()
	if err != nil {
		return nil, err
	}
	f, ok := final.(wizardModel)
	if !ok {
		return nil, errors.New("stepflow: unexpected program model")
	}
	if f.failure != nil {
		return nil, f.failure
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !f.done || f.result == nil {
		return nil, ErrCancelled
	}
	return f.result, nil
}
func validateSteps(steps []Step) error {
	if len(steps) == 0 {
		return errors.New("stepflow: no steps configured")
	}
	seen := map[string]bool{}
	for i, s := range steps {
		if s == nil {
			return fmt.Errorf("stepflow: step %d is nil", i)
		}
		k := s.Key()
		if k == "" {
			return fmt.Errorf("stepflow: step %d has empty key", i)
		}
		if seen[k] {
			return fmt.Errorf("stepflow: duplicate step key %q", k)
		}
		seen[k] = true
		if l, ok := s.(*LoadStep); ok && l.work == nil && l.contextWork == nil {
			return fmt.Errorf("stepflow: load step %q has no worker", k)
		}
		if l, ok := s.(*ListStep); ok && l.visibleRows <= 0 {
			return fmt.Errorf("stepflow: list step %q needs positive VisibleRows", k)
		}
	}
	return nil
}
