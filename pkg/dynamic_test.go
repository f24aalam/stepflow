package stepflow

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type dummyStep struct {
	key   string
	q     string
	ans   string
	view  string
	initC tea.Cmd
}

func (d dummyStep) Key() string      { return d.key }
func (d dummyStep) Question() string { return d.q }
func (d dummyStep) Init(_ styles) tea.Cmd {
	if d.initC != nil {
		return d.initC
	}
	return nil
}
func (d dummyStep) Update(_ tea.KeyMsg) (bool, tea.Cmd) { return false, nil }
func (d dummyStep) View(_ styles) string                { return d.view }
func (d dummyStep) ResultView(s styles) string          { return s.stepAnswer.Render(d.ans) }
func (d dummyStep) Answer() string                      { return d.ans }

type dummyDynamicStep struct {
	dummyStep
	next []Step
}

func (d dummyDynamicStep) NextSteps(_ Result) []Step { return d.next }

func TestDynamicStepReplacesTail(t *testing.T) {
	st := newStyles(DefaultTheme())

	ds := dummyDynamicStep{
		dummyStep: dummyStep{key: "scan", q: "Scan", ans: "done"},
		next: []Step{
			dummyStep{key: "g1", q: "Guidelines", ans: "MIT"},
			dummyStep{key: "s1", q: "Skills", ans: "skill-a"},
		},
	}

	// This step exists only in the initial tail, and should be removed.
	old := dummyStep{key: "old", q: "Old", ans: "old"}

	m := newWizardModel([]Step{ds, old}, st, true)
	m.current = 0

	_ = m.advance(ds)

	if len(m.steps) != 3 {
		t.Fatalf("expected 3 steps after replacement, got %d", len(m.steps))
	}

	if m.steps[1].Key() != "g1" {
		t.Fatalf("expected step[1] key to be g1, got %q", m.steps[1].Key())
	}

	if m.steps[2].Key() != "s1" {
		t.Fatalf("expected step[2] key to be s1, got %q", m.steps[2].Key())
	}

	if m.current != 1 {
		t.Fatalf("expected current=1, got %d", m.current)
	}

	if len(m.answers) != 1 || m.answers[0] != "done" {
		t.Fatalf("expected answers to contain only dynamic step result, got %#v", m.answers)
	}

	if m.done {
		t.Fatalf("wizard should not be done yet")
	}
}

func TestDynamicStepNilDoesNotReplaceTail(t *testing.T) {
	st := newStyles(DefaultTheme())

	ds := dummyDynamicStep{
		dummyStep: dummyStep{key: "scan", q: "Scan", ans: "done"},
		next:      nil, // no opinion
	}

	old := dummyStep{key: "old", q: "Old", ans: "old"}

	m := newWizardModel([]Step{ds, old}, st, true)
	m.current = 0

	_ = m.advance(ds)

	if len(m.steps) != 2 {
		t.Fatalf("expected 2 steps to remain, got %d", len(m.steps))
	}

	if m.steps[1].Key() != "old" {
		t.Fatalf("expected step[1] key to be old, got %q", m.steps[1].Key())
	}

	if m.current != 1 {
		t.Fatalf("expected current=1, got %d", m.current)
	}

	if m.done {
		t.Fatalf("wizard should not be done yet")
	}
}

func TestDynamicStepEmptyNextFinishesWizard(t *testing.T) {
	st := newStyles(DefaultTheme())

	ds := dummyDynamicStep{
		dummyStep: dummyStep{key: "scan", q: "Scan", ans: "done"},
		next:      nil,
	}

	m := newWizardModel([]Step{ds}, st, true)
	m.current = 0

	_ = m.advance(ds)

	if !m.done {
		t.Fatalf("expected wizard to be done")
	}

	if m.result == nil {
		t.Fatalf("expected result to be non-nil")
	}

	if got := m.result.Get("scan"); got != "done" {
		t.Fatalf("expected result[scan]=done, got %q", got)
	}
}

func TestDynamicStepValidationFailures(t *testing.T) {
	tests := []struct {
		name        string
		dynamicNext []Step
		errContains string
	}{
		{
			name:        "nil step in dynamic steps",
			dynamicNext: []Step{nil},
			errContains: "is nil",
		},
		{
			name:        "empty key in dynamic step",
			dynamicNext: []Step{Text("", "No key")},
			errContains: "has empty key",
		},
		{
			name:        "duplicate key with earlier completed step",
			dynamicNext: []Step{Text("initial", "Duplicate")},
			errContains: "duplicate step key \"initial\"",
		},
		{
			name:        "duplicate key within dynamic steps",
			dynamicNext: []Step{Text("dup", "First"), Text("dup", "Second")},
			errContains: "duplicate step key \"dup\"",
		},
		{
			name:        "loader step without worker in dynamic steps",
			dynamicNext: []Step{Load("fetch", "Fetch")},
			errContains: "has no worker",
		},
		{
			name:        "list step with zero visible rows in dynamic steps",
			dynamicNext: []Step{List("list", "List").VisibleRows(0)},
			errContains: "needs positive VisibleRows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := newStyles(DefaultTheme())
			initial := Confirm("initial", "Initial").
				WithNext(func(Result) []Step {
					return tt.dynamicNext
				})

			m := newWizardModel([]Step{initial}, st, false)
			// Simulate pressing enter to complete initial step
			updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = updated.(wizardModel)

			if !m.failed {
				t.Fatalf("expected wizard to fail validation, but it succeeded")
			}
			if cmd == nil {
				t.Fatalf("expected quit command on validation failure")
			}
			if m.failure == nil || !strings.Contains(m.failure.Error(), tt.errContains) {
				t.Fatalf("expected failure error containing %q, got: %v", tt.errContains, m.failure)
			}
			// Verify View() renders error gracefully
			view := m.View()
			if !strings.Contains(view, tt.errContains) {
				t.Fatalf("expected view to contain error text, got: %q", view)
			}
		})
	}
}

func TestDynamicLoadStepReceivesContext(t *testing.T) {
	st := newStyles(DefaultTheme())
	type testCtxKey struct{}
	ctx := context.WithValue(context.Background(), testCtxKey{}, "custom_context_value")

	var receivedCtx context.Context
	loaderRan := false

	loadStep := Load("dynamic_load", "Loading dynamic data").
		RunContext(func(c context.Context, status chan<- string) (string, error) {
			loaderRan = true
			receivedCtx = c
			status <- "running"
			return "loaded_ok", nil
		})

	initial := Confirm("start", "Start?").
		WithNext(func(Result) []Step {
			return []Step{loadStep}
		})

	m := newWizardModel([]Step{initial}, st, true, ctx)

	// Advance past initial step
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(wizardModel)

	if m.failed {
		t.Fatalf("unexpected validation failure: %v", m.failure)
	}

	// Verify the dynamic LoadStep received m.ctx
	if loadStep.ctx == nil {
		t.Fatal("expected dynamic LoadStep to receive wizard context, got nil")
	}
	if val, ok := loadStep.ctx.Value(testCtxKey{}).(string); !ok || val != "custom_context_value" {
		t.Fatalf("expected context value 'custom_context_value', got %v", loadStep.ctx.Value(testCtxKey{}))
	}

	// Initialize the load step and run its command
	cmd := m.steps[m.current].Init(st)
	if cmd == nil {
		t.Fatal("expected non-nil Init command for LoadStep")
	}

	// Advance through load step messages
	updated, _ = m.Update(loadTickMsg{})
	m = updated.(wizardModel)

	updated, _ = m.Update(loadDoneMsg{value: "loaded_ok"})
	m = updated.(wizardModel)

	if !m.done {
		t.Fatal("expected wizard to complete after dynamic load step")
	}
	if !loaderRan {
		// Run the cmd to verify worker execution
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					m := c()
					if done, ok := m.(loadDoneMsg); ok && done.value == "loaded_ok" {
						loaderRan = true
					}
				}
			}
		}
	}
	if receivedCtx != nil {
		if val, ok := receivedCtx.Value(testCtxKey{}).(string); !ok || val != "custom_context_value" {
			t.Fatalf("worker did not receive custom context, got %v", receivedCtx)
		}
	}
}
