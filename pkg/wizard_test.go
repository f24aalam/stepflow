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
	th := NewTheme("#06b6d4", "#a855f7", "#64748b")
	if th.Primary != "#06b6d4" || th.Accent != "#a855f7" || th.Muted != "#64748b" {
		t.Fatalf("unexpected NewTheme colors: %#v", th)
	}

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

func TestWizardRepeatedRuns(t *testing.T) {
	st := newStyles(DefaultTheme())

	textStep := Text("name", "Name").Default("initial-name")
	confirmStep := Confirm("deploy", "Deploy?").Default("no")
	listStep := List("fruit", "Pick Fruit").
		Items(Item("Apple"), Item("Banana"), Item("Cherry")).
		PreSelect("Banana")
	loadStep := Load("sync", "Syncing").Run(func(status chan<- string) (string, error) {
		return "synced", nil
	})

	w := New().WithSteps(textStep, confirmStep, listStep, loadStep)

	// Simulate Run 1 mutations:
	// 1. Text input has value edited
	textStep.Init(st)
	textStep.input.SetValue("edited-name")
	if textStep.Answer() != "edited-name" {
		t.Fatalf("expected edited-name, got %q", textStep.Answer())
	}

	// 2. Confirm cursor is moved up to "Yes"
	confirmStep.Init(st)
	confirmStep.Update(tea.KeyMsg{Type: tea.KeyUp})
	if confirmStep.Answer() != "Yes" {
		t.Fatalf("expected Yes after moving up, got %q", confirmStep.Answer())
	}

	// 3. List selections are changed to Cherry only
	listStep.Init(st)
	listStep.Update(tea.KeyMsg{Type: tea.KeyDown})
	listStep.Update(tea.KeyMsg{Type: tea.KeyDown})
	listStep.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if listStep.Answer() != "Cherry" {
		t.Fatalf("expected Cherry, got %q", listStep.Answer())
	}

	// 4. Load step finished
	loadStep.Init(st)
	loadStep.UpdateMsg(loadDoneMsg{value: "synced"})
	if !loadStep.finished || loadStep.Answer() != "synced" {
		t.Fatalf("expected finished load step, got %v / %q", loadStep.finished, loadStep.Answer())
	}

	// Now call w.Reset() (which is also called at the start of Run / RunContext)
	w.Reset()

	// Verify all steps are restored to their configured initial states:
	if ans := textStep.Answer(); ans != "initial-name" {
		t.Fatalf("expected text step reset to 'initial-name', got %q", ans)
	}

	if ans := confirmStep.Answer(); ans != "No" {
		t.Fatalf("expected confirm step reset to default 'No', got %q", ans)
	}

	if ans := listStep.Answer(); ans != "Banana" {
		t.Fatalf("expected list step reset to preselected 'Banana', got %q", ans)
	}

	if loadStep.finished || loadStep.Answer() != "" {
		t.Fatalf("expected load step reset, finished=%v, answer=%q", loadStep.finished, loadStep.Answer())
	}
}

func TestLoadStepRapidStatusAndCancellation(t *testing.T) {
	st := newStyles(DefaultTheme())
	ctx, cancel := context.WithCancel(context.Background())

	sentinelErr := errors.New("worker canceled cooperatively")
	sentCount := 0

	load := Load("rapid", "Rapid sending").
		RunContext(func(c context.Context, status chan<- string) (string, error) {
			for i := 0; i < 200; i++ {
				select {
				case <-c.Done():
					return "", sentinelErr
				case status <- "working":
					sentCount++
				}
			}
			return "done", nil
		})

	load.ctx = ctx
	cmd := load.Init(st)
	if cmd == nil {
		t.Fatal("expected Init command")
	}

	// Cancel context while work is in flight
	cancel()

	// Run worker command
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				res := c()
				if errMsg, ok := res.(loadErrMsg); ok {
					if !errors.Is(errMsg.err, sentinelErr) {
						t.Fatalf("expected %v, got %v", sentinelErr, errMsg.err)
					}
				}
			}
		}
	}
}

func TestLoadStepPreCanceledContext(t *testing.T) {
	st := newStyles(DefaultTheme())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled

	workerCalled := false
	load := Load("pre", "Pre-canceled").
		RunContext(func(c context.Context, status chan<- string) (string, error) {
			workerCalled = true
			return "ok", nil
		})

	load.ctx = ctx
	cmd := load.Init(st)
	if cmd == nil {
		t.Fatal("expected non-nil Init command")
	}

	msg := cmd()
	if errMsg, ok := msg.(loadErrMsg); !ok || !errors.Is(errMsg.err, context.Canceled) {
		t.Fatalf("expected context.Canceled error msg, got %#v", msg)
	}

	if workerCalled {
		t.Fatal("worker should not have been called when context was pre-canceled")
	}
}
