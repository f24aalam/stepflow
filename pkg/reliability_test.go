package stepflow

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestValidateSteps(t *testing.T) {
	cases := []struct{ steps []Step }{{nil}, {[]Step{Text("", "bad")}}, {[]Step{Text("dup", "A"), Text("dup", "B")}}, {[]Step{Load("load", "Loading")}}, {[]Step{List("list", "Pick").VisibleRows(0)}}}
	for _, c := range cases {
		if err := validateSteps(c.steps); err == nil {
			t.Fatalf("expected error for %#v", c.steps)
		}
	}
}
func TestLoadFailurePreservesError(t *testing.T) {
	sentinel := errors.New("network failed")
	s := Load("fetch", "Fetch")
	m := newWizardModel([]Step{s}, newStyles(DefaultTheme()), true)
	updated, cmd := m.Update(loadErrMsg{sentinel})
	m = updated.(wizardModel)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !errors.Is(m.failure, sentinel) {
		t.Fatalf("unexpected failure: %v", m.failure)
	}
}
func TestListEmptyAndFiltering(t *testing.T) {
	s := List("items", "Items")
	s.Init(newStyles(DefaultTheme()))
	for _, key := range []tea.KeyMsg{{Type: tea.KeySpace}, {Type: tea.KeyEnter}} {
		s.Update(key)
	}
	_ = s.View(newStyles(DefaultTheme()))
	s.Items(Item("alpha"))
	s.Init(newStyles(DefaultTheme()))
	s.search = "zzz"
	s.ensureCursorValid()
	s.Update(tea.KeyMsg{Type: tea.KeySpace})
	_ = s.View(newStyles(DefaultTheme()))
}
func TestUnicodeSearchBackspace(t *testing.T) {
	s := List("key", "Q").Items(Item("éclair"))
	s.Init(newStyles(DefaultTheme()))
	s.search = "é"
	s.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if s.search != "" {
		t.Fatalf("backspace broke unicode: %q", s.search)
	}
}
func TestTerminalNextAcceptsEnter(t *testing.T) {
	s := Success("done", "Done").WithNext(func(Result) []Step { return []Step{Info("next", "Next")} })
	done, _ := s.UpdateMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if !done {
		t.Fatal("enter must advance")
	}
}
func TestDynamicReplacesTail(t *testing.T) {
	s := Confirm("flag", "Flag").WithNext(func(Result) []Step { return []Step{Text("new", "New")} })
	m := newWizardModel([]Step{s, Info("old", "Old")}, newStyles(DefaultTheme()), false)
	m.advance(s)
	if len(m.steps) != 2 || m.steps[1].Key() != "new" {
		t.Fatal("dynamic tail not replaced")
	}
}
