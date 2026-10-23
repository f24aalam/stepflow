package stepflow

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestMultiSelectWithCommasAndQuotes(t *testing.T) {
	st := newStyles(DefaultTheme())

	// Test 1: Labels with commas, quotes, leading/trailing spaces, newlines, and empty labels
	list := List("items", "Select items").
		MultiSelect(true).
		Items(
			Item("Smith, John"),
			Item("Doe, Jane"),
			Item("Plain"),
			Item(`O'Connor, "Bob"`),
			Item("  leading and trailing  "),
			Item("line1\nline2"),
			Item(""),
		).
		PreSelect(
			"Smith, John",
			"Doe, Jane",
			`O'Connor, "Bob"`,
			"  leading and trailing  ",
			"line1\nline2",
			"",
		)

	list.Init(st)

	ans := list.Answer()
	r := Result{"items": ans}

	got := r.Strings("items")
	expected := []string{
		"Smith, John",
		"Doe, Jane",
		`O'Connor, "Bob"`,
		"  leading and trailing  ",
		"line1\nline2",
		"",
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d items, got %d: %#v", len(expected), len(got), got)
	}
	for i, exp := range expected {
		if got[i] != exp {
			t.Errorf("item %d: expected %q, got %q", i, exp, got[i])
		}
	}

	// Verify Result.Get returns non-empty representation
	if r.Get("items") == "" {
		t.Fatal("expected non-empty Get value")
	}

	// Test 2: Single-select with comma in label
	single := List("author", "Select author").
		Items(Item("Smith, John")).
		PreSelect("Smith, John")
	single.Init(st)
	if single.Answer() != "Smith, John" {
		t.Fatalf("expected 'Smith, John', got %q", single.Answer())
	}
	rSingle := Result{"author": single.Answer()}
	if rSingle.Get("author") != "Smith, John" {
		t.Fatalf("expected 'Smith, John', got %q", rSingle.Get("author"))
	}

	// Test 3: Backward compatibility with unquoted comma-separated string
	legacyResult := Result{"tags": "go, cli, tui"}
	tags := legacyResult.Strings("tags")
	if len(tags) != 3 || tags[0] != "go" || tags[1] != "cli" || tags[2] != "tui" {
		t.Fatalf("unexpected legacy tags split: %#v", tags)
	}

	// Test 4: Single empty label selected in multi-select
	emptyList := List("empty", "Pick empty").
		MultiSelect(true).
		Items(Item("")).
		PreSelect("")
	emptyList.Init(st)
	rEmpty := Result{"empty": emptyList.Answer()}
	emptyGot := rEmpty.Strings("empty")
	if len(emptyGot) != 1 || emptyGot[0] != "" {
		t.Fatalf("expected slice with one empty string, got %#v", emptyGot)
	}
}

func TestTerminalResizing(t *testing.T) {
	st := newStyles(DefaultTheme())

	// 1. TextStep reacts to terminal width
	text := Text("query", "Search query").Default("abc")
	text.Init(st)

	// Default width is 72
	if text.input.Width != 72 {
		t.Fatalf("expected default width 72, got %d", text.input.Width)
	}

	// Narrow terminal (width 36)
	text.UpdateMsg(tea.WindowSizeMsg{Width: 36, Height: 20})
	if text.input.Width != 34 {
		t.Fatalf("expected width to resize to 34 (36-2), got %d", text.input.Width)
	}

	// Wide terminal (width 120) caps at 72
	text.UpdateMsg(tea.WindowSizeMsg{Width: 120, Height: 20})
	if text.input.Width != 72 {
		t.Fatalf("expected width capped at 72, got %d", text.input.Width)
	}

	// Extremely narrow dimensions
	for _, w := range []int{10, 5, 2, 1, 0, -5} {
		text.UpdateMsg(tea.WindowSizeMsg{Width: w, Height: 20})
		if w > 0 && text.input.Width < 1 {
			t.Fatalf("width %d resulted in invalid text input width: %d", w, text.input.Width)
		}
		if w >= 2 && text.input.Width > w {
			t.Fatalf("terminal width %d exceeded by input width %d", w, text.input.Width)
		}
	}

	// 2. ListStep reacts to terminal width and avoids overflow
	longLabel := strings.Repeat("VeryLongOptionName", 10)
	list := List("opts", "Choose option").
		Items(Item(longLabel, "description that is also long"))
	list.Init(st)

	// Set narrow terminal width 40
	list.UpdateMsg(tea.WindowSizeMsg{Width: 40, Height: 20})
	view := list.View(st)

	// Verify lines do not exceed terminal width (excluding ANSI escape codes)
	for _, line := range strings.Split(view, "\n") {
		lineWidth := lipgloss.Width(line)
		if lineWidth > 40 {
			t.Errorf("list view line exceeded terminal width 40: width was %d for line %q", lineWidth, line)
		}
	}

	// 3. InfoStep with narrow terminal wraps rather than truncates
	longBody := "First line with important end token XYZ.\nSecond line with tail token ABC."
	info := Info("info", "Information").Body(longBody)
	info.UpdateMsg(tea.WindowSizeMsg{Width: 25, Height: 20})
	infoView := info.View(st)
	if !strings.Contains(infoView, "XYZ") || !strings.Contains(infoView, "ABC") {
		t.Fatalf("expected info view to wrap and preserve tokens XYZ and ABC, got: %q", infoView)
	}
}

func TestLongMessagesWrappedNotTruncated(t *testing.T) {
	st := newStyles(DefaultTheme())

	// 1. Long error message in LoadStep
	longErr := errors.New("database connection refused: dial tcp 192.168.1.50:5432: connection timeout after 30 seconds")
	load := Load("fetch", "Fetch data")
	load.UpdateMsg(tea.WindowSizeMsg{Width: 35, Height: 20})
	load.UpdateMsg(loadErrMsg{err: longErr})
	loadView := load.View(st)
	if !strings.Contains(loadView, "timeout after") || !strings.Contains(loadView, "seconds") {
		t.Fatalf("expected long error message to be preserved through wrapping, got: %q", loadView)
	}

	// 2. Long loading status in LoadStep
	longStatus := "Downloading massive dependency archive from remote registry (95% completed)"
	loadActive := Load("active", "Working")
	loadActive.Init(st)
	loadActive.UpdateMsg(tea.WindowSizeMsg{Width: 35, Height: 20})
	loadActive.status = longStatus
	activeView := loadActive.View(st)
	if !strings.Contains(activeView, "95%") || !strings.Contains(activeView, "completed") {
		t.Fatalf("expected long status message to be preserved through wrapping, got: %q", activeView)
	}

	// 3. Long message in TerminalStep
	longTermMsg := "Scaffolding finished completely: inspect ./my-app and execute 'npm run dev' to launch"
	term := Success("done", "Finished").
		Body(longTermMsg).
		WithNext(func(Result) []Step { return nil })
	term.UpdateMsg(tea.WindowSizeMsg{Width: 30, Height: 20})
	termView := term.View(st)
	if !strings.Contains(termView, "npm run") || !strings.Contains(termView, "dev") {
		t.Fatalf("expected terminal step message to be wrapped without truncation, got: %q", termView)
	}
}

func TestCompactSpacingAndRender(t *testing.T) {
	st := newStyles(DefaultTheme())

	// Check helper text indentation across all steps:
	// TextStep helper has 2-space indentation
	text := Text("k", "Q")
	text.Init(st)
	if !strings.Contains(text.View(st), "  "+st.helper.Render("enter to confirm")) {
		t.Fatalf("expected TextStep helper with 2-space indentation, got: %q", text.View(st))
	}

	// ConfirmStep helper has 2-space indentation
	confirm := Confirm("k", "Q")
	confirm.Init(st)
	if !strings.Contains(confirm.View(st), "  "+st.helper.Render("↑↓ move · enter confirm")) {
		t.Fatalf("expected ConfirmStep helper with 2-space indentation, got: %q", confirm.View(st))
	}

	// ListStep helper has 2-space indentation
	list := List("k", "Q").Items(Item("A"))
	list.Init(st)
	if !strings.Contains(list.View(st), "  "+st.helper.Render("↑↓ navigate · enter select · esc clear search")) {
		t.Fatalf("expected ListStep helper with 2-space indentation, got: %q", list.View(st))
	}

	// InfoStep helper has 2-space indentation
	info := Info("k", "Q")
	if !strings.Contains(info.View(st), "  "+st.helper.Render("enter to continue")) {
		t.Fatalf("expected InfoStep helper with 2-space indentation, got: %q", info.View(st))
	}

	// TerminalStep helper has 2-space indentation
	term := Success("k", "Q").WithNext(func(Result) []Step { return nil })
	if !strings.Contains(term.View(st), "  "+st.helper.Render("enter to continue")) {
		t.Fatalf("expected TerminalStep helper with 2-space indentation, got: %q", term.View(st))
	}
}
