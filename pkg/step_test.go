package stepflow

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTextStep(t *testing.T) {
	st := newStyles(DefaultTheme())

	s := Text("username", "Enter username").
		Placeholder("e.g. alice").
		Default("guest")

	if s.Key() != "username" || s.Question() != "Enter username" {
		t.Fatalf("unexpected key or question: %s / %s", s.Key(), s.Question())
	}

	cmd := s.Init(st)
	if cmd == nil {
		t.Fatal("expected Init to return focus command")
	}

	// Verify default answer if enter pressed immediately
	if ans := s.Answer(); ans != "guest" {
		t.Fatalf("expected default answer 'guest', got %q", ans)
	}

	// Press Enter to confirm
	done, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !done {
		t.Fatal("expected enter to advance text step")
	}

	// View rendering checks
	view := s.View(st)
	if !strings.Contains(view, "enter to confirm") {
		t.Fatalf("expected view to contain helper text, got %q", view)
	}

	resView := s.ResultView(st)
	if !strings.Contains(resView, "guest") {
		t.Fatalf("expected result view to contain 'guest', got %q", resView)
	}

	// Test WithNext chaining
	s.WithNext(func(r Result) []Step {
		return []Step{Confirm("ok", "OK?")}
	})
	next := s.NextSteps(Result{"username": "alice"})
	if len(next) != 1 || next[0].Key() != "ok" {
		t.Fatalf("expected dynamic next step 'ok', got %#v", next)
	}
}

func TestConfirmStep(t *testing.T) {
	st := newStyles(DefaultTheme())

	// Default No
	cNo := Confirm("proceed", "Proceed?").Default("no")
	if cNo.Answer() != "No" {
		t.Fatalf("expected default No, got %q", cNo.Answer())
	}

	// Move Up with "up" key
	cNo.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cNo.Answer() != "Yes" {
		t.Fatalf("expected Yes after moving up, got %q", cNo.Answer())
	}

	// Move Down with vim "j" key
	cNo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cNo.Answer() != "No" {
		t.Fatalf("expected No after pressing j, got %q", cNo.Answer())
	}

	// Move Up with vim "k" key
	cNo.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cNo.Answer() != "Yes" {
		t.Fatalf("expected Yes after pressing k, got %q", cNo.Answer())
	}

	// Confirm with enter
	done, _ := cNo.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !done {
		t.Fatal("expected enter to confirm")
	}

	// View rendering
	v := cNo.View(st)
	if !strings.Contains(v, "Yes") || !strings.Contains(v, "No") {
		t.Fatalf("view missing options: %q", v)
	}

	rv := cNo.ResultView(st)
	if !strings.Contains(rv, "Yes") {
		t.Fatalf("result view missing answer: %q", rv)
	}

	// WithNext
	cNo.WithNext(func(r Result) []Step { return nil })
	if cNo.NextSteps(Result{}) != nil {
		t.Fatal("expected nil next steps")
	}
}

func TestListStepSingleSelect(t *testing.T) {
	st := newStyles(DefaultTheme())

	list := List("fruit", "Pick fruit").
		Items(
			Item("Apple", "crisp"),
			Item("Banana", "sweet"),
			Item("Cherry", "tart"),
		).
		PreSelect("Banana")

	list.Init(st)

	// Banana was preselected
	if list.Answer() != "Banana" {
		t.Fatalf("expected preselected Banana, got %q", list.Answer())
	}

	// Move cursor down twice: from 0 (Apple) -> 1 (Banana) -> 2 (Cherry)
	list.Update(tea.KeyMsg{Type: tea.KeyDown})
	list.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter in single-select mode
	done, _ := list.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !done {
		t.Fatal("expected Enter to advance list step")
	}
	if list.Answer() != "Cherry" {
		t.Fatalf("expected Cherry after moving down and selecting, got %q", list.Answer())
	}

	view := list.View(st)
	if !strings.Contains(view, "Cherry") || !strings.Contains(view, "tart") {
		t.Fatalf("expected view to contain item and metadata, got: %q", view)
	}
}

func TestListStepMultiSelectAndPinned(t *testing.T) {
	st := newStyles(DefaultTheme())

	list := List("plugins", "Select plugins").
		MultiSelect(true).
		VisibleRows(2).
		Pinned("Core", "Always included", "Base").
		Items(
			Item("Plugin A"),
			Item("Plugin B"),
			Item("Plugin C"),
			Item("Plugin D"),
		).
		PreSelect("Plugin A")

	list.Init(st)

	// Initial selection
	if list.Answer() != "Plugin A" {
		t.Fatalf("expected Plugin A, got %q", list.Answer())
	}

	// Space on current item toggles it off
	list.Update(tea.KeyMsg{Type: tea.KeySpace})
	if list.Answer() != "" {
		t.Fatalf("expected empty selection after toggle off, got %q", list.Answer())
	}

	// Move down to Plugin B and toggle it on
	list.Update(tea.KeyMsg{Type: tea.KeyDown})
	list.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Move down to Plugin C and toggle it on
	list.Update(tea.KeyMsg{Type: tea.KeyDown})
	list.Update(tea.KeyMsg{Type: tea.KeySpace})

	ans := list.Answer()
	if !strings.Contains(ans, "Plugin B") || !strings.Contains(ans, "Plugin C") {
		t.Fatalf("expected B and C selected, got %q", ans)
	}

	// View rendering checks: pinned section, pagination, search slash
	view := list.View(st)
	if !strings.Contains(view, "Core") || !strings.Contains(view, "Always included") {
		t.Fatalf("expected pinned title and note in view, got: %q", view)
	}
	if !strings.Contains(view, "Base") {
		t.Fatalf("expected pinned item 'Base' in view, got: %q", view)
	}
	// Pagination indicator (3 / 4)
	if !strings.Contains(view, "/ 4") {
		t.Fatalf("expected pagination indicator in view, got: %q", view)
	}

	// Test Esc key with empty search does not panic
	list.Update(tea.KeyMsg{Type: tea.KeyEsc})
}

func TestListStepSearchAndFiltering(t *testing.T) {
	st := newStyles(DefaultTheme())

	list := List("langs", "Choose languages").
		Items(
			Item("Go"),
			Item("Rust"),
			Item("Python"),
		)

	list.Init(st)

	// Type 'r'
	list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if len(list.filtered()) != 1 || list.items[list.filtered()[0]].item.Label != "Rust" {
		t.Fatalf("expected only Rust to match 'r', got %#v", list.filtered())
	}

	// Backspace to clear 'r'
	list.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if len(list.filtered()) != 3 {
		t.Fatalf("expected all 3 items after backspace, got %d", len(list.filtered()))
	}

	// Search non-existent
	list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z', 'z', 'z'}})
	if len(list.filtered()) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(list.filtered()))
	}

	// Enter on empty filter shouldn't finish
	done, _ := list.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if done {
		t.Fatal("should not complete when no items match search")
	}

	// Esc clears search
	list.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if len(list.filtered()) != 3 {
		t.Fatalf("expected 3 items after Esc cleared search, got %d", len(list.filtered()))
	}
}

func TestInfoStep(t *testing.T) {
	st := newStyles(DefaultTheme())

	info := Info("welcome", "Welcome").Body("Line 1\nLine 2")
	if info.Key() != "welcome" || info.Question() != "Welcome" {
		t.Fatalf("unexpected info key/question")
	}

	if cmd := info.Init(st); cmd != nil {
		t.Fatal("expected nil Init cmd for InfoStep")
	}

	if ans := info.Answer(); ans != "shown" {
		t.Fatalf("expected answer 'shown', got %q", ans)
	}

	v := info.View(st)
	if !strings.Contains(v, "Line 1") || !strings.Contains(v, "enter to continue") {
		t.Fatalf("unexpected info view: %q", v)
	}

	rv := info.ResultView(st)
	if !strings.Contains(rv, "shown") {
		t.Fatalf("unexpected result view: %q", rv)
	}

	// Default body when empty
	emptyInfo := Info("empty", "Empty")
	if !strings.Contains(emptyInfo.View(st), "No details.") {
		t.Fatal("expected 'No details.' fallback for empty body")
	}

	done, _ := info.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !done {
		t.Fatal("expected enter to advance InfoStep")
	}

	// Dynamic next
	info.WithNext(func(Result) []Step { return []Step{Confirm("c", "Q")} })
	if len(info.NextSteps(Result{})) != 1 {
		t.Fatal("expected dynamic steps from info")
	}
}

func TestTerminalStep(t *testing.T) {
	st := newStyles(DefaultTheme())

	// Standalone Success terminal step without WithNext auto-advances
	succ := Success("done", "All Completed").Body("Great job!")
	if succ.HasError() {
		t.Fatal("terminal step should not have error")
	}

	cmd := succ.Init(st)
	if cmd == nil {
		t.Fatal("expected termDoneMsg generator cmd for auto-advancing terminal step")
	}
	msg := cmd()
	done, _ := succ.UpdateMsg(msg)
	if !done {
		t.Fatal("expected termDoneMsg to finish terminal step")
	}
	if succ.Answer() != "Great job!" {
		t.Fatalf("expected body as answer, got %q", succ.Answer())
	}

	// Terminal step with WithNext waits for Enter
	interactive := Error("oops", "Failed").
		Body("Something broke").
		WithNext(func(Result) []Step {
			return []Step{Confirm("retry", "Retry?")}
		})

	if interactive.Init(st) != nil {
		t.Fatal("interactive terminal step should return nil Init cmd")
	}
	v := interactive.View(st)
	if !strings.Contains(v, "Something broke") || !strings.Contains(v, "enter to continue") {
		t.Fatalf("unexpected interactive terminal view: %q", v)
	}
	doneEnter, _ := interactive.UpdateMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if !doneEnter {
		t.Fatal("expected Enter to advance interactive terminal step")
	}
	if len(interactive.NextSteps(Result{})) != 1 {
		t.Fatal("expected next steps to be returned")
	}
}

func TestResultHelpers(t *testing.T) {
	r := Result{
		"text":        "hello world",
		"flag_yes":    "Yes",
		"flag_no":     "No",
		"flag_lower":  "yes",
		"multi":       "item1, item2, item3",
		"empty_multi": "   ",
	}

	// Get
	if r.Get("text") != "hello world" {
		t.Fatalf("unexpected Get value: %s", r.Get("text"))
	}
	if r.Get("missing") != "" {
		t.Fatalf("expected empty string for missing key, got %q", r.Get("missing"))
	}

	// Bool
	if !r.Bool("flag_yes") || !r.Bool("flag_lower") {
		t.Fatal("expected true for 'Yes' and 'yes'")
	}
	if r.Bool("flag_no") || r.Bool("missing") {
		t.Fatal("expected false for 'No' and missing keys")
	}

	// Strings
	items := r.Strings("multi")
	if len(items) != 3 || items[0] != "item1" || items[1] != "item2" || items[2] != "item3" {
		t.Fatalf("unexpected Strings split: %#v", items)
	}
	if r.Strings("empty_multi") != nil || r.Strings("nonexistent") != nil {
		t.Fatal("expected nil slice for empty/missing Strings key")
	}
}
