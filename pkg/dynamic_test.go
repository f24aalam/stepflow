package stepflow

import (
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

	m := newWizardModel([]Step{ds, old}, st)
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

func TestDynamicStepEmptyNextFinishesWizard(t *testing.T) {
	st := newStyles(DefaultTheme())

	ds := dummyDynamicStep{
		dummyStep: dummyStep{key: "scan", q: "Scan", ans: "done"},
		next:      nil,
	}

	m := newWizardModel([]Step{ds}, st)
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
