package stepflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWizardValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		steps       []Step
		errContains string
	}{
		{
			name:        "no steps",
			steps:       []Step{},
			errContains: "no steps configured",
		},
		{
			name:        "nil step",
			steps:       []Step{nil},
			errContains: "step 0 is nil",
		},
		{
			name:        "empty key",
			steps:       []Step{Text("", "No key")},
			errContains: "step 0 has empty key",
		},
		{
			name:        "duplicate key",
			steps:       []Step{Text("same", "First"), Text("same", "Second")},
			errContains: "duplicate step key \"same\"",
		},
		{
			name:        "load step without worker",
			steps:       []Step{Load("fetch", "Fetching")},
			errContains: "load step \"fetch\" has no worker",
		},
		{
			name:        "list step with zero visible rows",
			steps:       []Step{List("list", "Choose").VisibleRows(0)},
			errContains: "needs positive VisibleRows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New().WithSteps(tt.steps...)
			_, err := w.Run()
			if err == nil || !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestWizardContextPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel upfront

	w := New().WithSteps(Text("name", "Name"))
	_, err := w.RunContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestWizardOptionsConfig(t *testing.T) {
	th := DefaultTheme()
	w := New().
		WithAltScreen(true).
		WithDoneScreen(false).
		WithTheme(th).
		WithSteps(Text("k", "Q"))

	if !w.useAltScreen {
		t.Fatal("expected useAltScreen to be true")
	}
	if w.showDone {
		t.Fatal("expected showDone to be false")
	}
	if len(w.steps) != 1 {
		t.Fatalf("expected 1 step configured, got %d", len(w.steps))
	}
}

func TestWizardDoneScreenToggle(t *testing.T) {
	st := newStyles(DefaultTheme())
	step := Text("name", "Name").Default("value")

	// Model with showDone = true
	mWithDone := newWizardModel([]Step{step}, st, true)
	mWithDone.done = true
	if !strings.Contains(mWithDone.View(), "All done!") {
		t.Fatal("expected 'All done!' when showDone is true")
	}

	// Model with showDone = false
	mNoDone := newWizardModel([]Step{step}, st, false)
	mNoDone.done = true
	if strings.Contains(mNoDone.View(), "All done!") {
		t.Fatal("did not expect 'All done!' when showDone is false")
	}
}

func TestWizardFailedView(t *testing.T) {
	st := newStyles(DefaultTheme())
	step := Text("name", "Name")

	m := newWizardModel([]Step{step}, st, true)
	m.failed = true
	m.failure = errors.New("database connection refused")

	v := m.View()
	if !strings.Contains(v, "✗ database connection refused") {
		t.Fatalf("expected failure message in view, got: %q", v)
	}
}

func TestWizardCtrlCAborts(t *testing.T) {
	st := newStyles(DefaultTheme())
	step := Text("name", "Name")

	m := newWizardModel([]Step{step}, st, true)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command on Ctrl+C")
	}
}

func TestLoadStepContextExecution(t *testing.T) {
	st := newStyles(DefaultTheme())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	statusReceived := false
	load := Load("fetch", "Fetching data").
		RunContext(func(ctx context.Context, status chan<- string) (string, error) {
			status <- "connecting"
			return "done_payload", nil
		})

	load.ctx = ctx
	cmd := load.Init(st)
	if cmd == nil {
		t.Fatal("expected batch command from LoadStep Init")
	}

	// Test tick message processing
	load.statusCh <- "in progress"
	load.UpdateMsg(loadTickMsg{})
	if load.status == "in progress" {
		statusReceived = true
	}
	if !statusReceived {
		t.Fatal("expected status update to be drained and applied")
	}

	// Test completion message
	done, _ := load.UpdateMsg(loadDoneMsg{value: "finished_ok"})
	if !done {
		t.Fatal("expected loadDoneMsg to finish step")
	}
	if load.Answer() != "finished_ok" {
		t.Fatalf("expected Answer 'finished_ok', got %q", load.Answer())
	}
	if load.HasError() {
		t.Fatal("did not expect error after loadDoneMsg")
	}
}

func TestLoadStepWorkerFailure(t *testing.T) {
	st := newStyles(DefaultTheme())
	expectedErr := errors.New("timeout dialing host")

	load := Load("fetch", "Fetching").
		Run(func(status chan<- string) (string, error) {
			return "", expectedErr
		})

	m := newWizardModel([]Step{load}, st, true)

	// Simulate receiving loadErrMsg
	updated, cmd := m.Update(loadErrMsg{err: expectedErr})
	m = updated.(wizardModel)

	if !m.failed {
		t.Fatal("expected wizardModel to be marked failed")
	}
	if !errors.Is(m.failure, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, m.failure)
	}
	if cmd == nil {
		t.Fatal("expected quit command on step error")
	}
	if !load.HasError() || !errors.Is(load.StepError(), expectedErr) {
		t.Fatal("expected step error state on load")
	}
}
