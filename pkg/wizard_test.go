package stepflow

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestLoadStepRapidStatusNormalCompletion(t *testing.T) {
	st := newStyles(DefaultTheme())
	sentCount := 0

	load := Load("rapid_ok", "Rapid Normal").
		Run(func(status chan<- string) (string, error) {
			for i := 0; i < 200; i++ {
				status <- "progress"
				sentCount++
			}
			return "completed_ok", nil
		})

	cmd := load.Init(st)
	if cmd == nil {
		t.Fatal("expected Init command")
	}

	// Execute batch command
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				res := c()
				if doneMsg, ok := res.(loadDoneMsg); ok {
					if doneMsg.value != "completed_ok" {
						t.Fatalf("expected 'completed_ok', got %q", doneMsg.value)
					}
				}
			}
		}
	}

	if sentCount != 200 {
		t.Fatalf("expected 200 sends, got %d", sentCount)
	}

	// Drain status channel via tick
	load.UpdateMsg(loadTickMsg{})
	if load.status != "progress" {
		t.Fatalf("expected status 'progress', got %q", load.status)
	}
}

var ansiSeq = regexp.MustCompile("\x1b\\[[0-9;]*m")

// completedView answers every step with Enter at the given terminal width and
// returns the final rendered history.
func completedView(t *testing.T, width int, steps ...Step) string {
	t.Helper()
	m := newWizardModel(steps, newStyles(DefaultTheme()), true)
	m.Init()
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: width, Height: 24})
	for i := 0; i < len(steps); i++ {
		tm, _ = tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	wm := tm.(wizardModel)
	if !wm.done {
		t.Fatalf("wizard did not complete")
	}
	return wm.View()
}

// historyStep holds the wrapped lines of one completed step, prefixes removed.
type historyStep struct{ question, answer []string }

// parseHistory checks the compact history layout line by line and returns the
// completed steps. Every line must fit the width, question continuation lines
// must be indented under the question, answer lines must carry the "│  "
// connector, and no blank lines may appear.
func parseHistory(t *testing.T, view string, width int) []historyStep {
	t.Helper()
	var out []historyStep
	lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
	prev := "" // "q", "a", "sep"
	for i, raw := range lines {
		if w := lipgloss.Width(raw); w > width {
			t.Errorf("line %d width %d exceeds terminal width %d: %q", i, w, width, raw)
		}
		line := ansiSeq.ReplaceAllString(raw, "")
		switch {
		case strings.HasPrefix(line, "◆  "):
			if prev != "" && prev != "sep" {
				t.Errorf("line %d: question must follow a separator: %q", i, line)
			}
			out = append(out, historyStep{question: []string{line[len("◆  "):]}})
			prev = "q"
		case strings.HasPrefix(line, "   ") && prev == "q":
			s := &out[len(out)-1]
			s.question = append(s.question, line[3:])
		case strings.HasPrefix(line, "│  "):
			if prev != "q" && prev != "a" {
				t.Errorf("line %d: answer line must follow its question: %q", i, line)
			}
			s := &out[len(out)-1]
			s.answer = append(s.answer, line[len("│  "):])
			prev = "a"
		case line == "│":
			if prev != "a" {
				t.Errorf("line %d: separator must follow an answer: %q", i, line)
			}
			prev = "sep"
		case line == "✓  All done!" && i == len(lines)-1:
		default:
			t.Errorf("line %d: unexpected line %q", i, line)
		}
	}
	return out
}

func squash(s string) string { return strings.Join(strings.Fields(s), "") }

func TestCompletedLongAnswerWraps(t *testing.T) {
	path := "/home/user/projects/a-very-long-project-directory/example/with/many/nested/segments"
	view := completedView(t, 40, Text("dir", "Project directory").Default(path))
	steps := parseHistory(t, view, 40)
	if len(steps) != 1 {
		t.Fatalf("expected 1 completed step, got %d: %q", len(steps), view)
	}
	if len(steps[0].answer) < 2 {
		t.Fatalf("expected long answer to wrap onto continuation lines, got %q", steps[0].answer)
	}
	if got := strings.Join(steps[0].answer, ""); got != path {
		t.Fatalf("answer clipped or altered:\nwant %q\ngot  %q", path, got)
	}
}

func TestCompletedLongQuestionWraps(t *testing.T) {
	q := "Which deployment environment should receive the generated configuration files today?"
	view := completedView(t, 40, Confirm("deploy", q))
	steps := parseHistory(t, view, 40)
	if len(steps) != 1 {
		t.Fatalf("expected 1 completed step, got %d: %q", len(steps), view)
	}
	if len(steps[0].question) < 2 {
		t.Fatalf("expected long question to wrap, got %q", steps[0].question)
	}
	if squash(strings.Join(steps[0].question, " ")) != squash(q) {
		t.Fatalf("question content altered: %q", steps[0].question)
	}
	if len(steps[0].answer) != 1 || steps[0].answer[0] != "Yes" {
		t.Fatalf("answer should stay on its own line below the question, got %q", steps[0].answer)
	}
}

func TestCompletedShortAnswerUnchanged(t *testing.T) {
	st := newStyles(DefaultTheme())
	want := st.marker.Render("◆") + "  " + st.stepLabel.Render("Name") + "\n" +
		st.pipe.Render("│") + "  " + st.stepAnswer.Render("Bob") + "\n" +
		st.pipe.Render("│") + "\n" +
		st.done.Render("✓  All done!") + "\n"
	for _, w := range []int{0, 20, 40, 80} {
		if got := completedView(t, w, Text("name", "Name").Default("Bob")); got != want {
			t.Errorf("width %d: short history changed:\nwant %q\ngot  %q", w, want, got)
		}
	}
}

func TestCompletedMultipleSteps(t *testing.T) {
	long := "a long answer that definitely needs more than one line at narrow widths"
	for _, w := range []int{20, 40, 80} {
		view := completedView(t, w,
			Text("name", "Name").Default("Bob"),
			Text("desc", "Description").Default(long),
			Confirm("git", "Initialize Git?"),
			Success("end", "Finished").Body("Run 'make install' in ./my-app to complete the setup"),
		)
		steps := parseHistory(t, view, w)
		if len(steps) != 4 {
			t.Fatalf("width %d: expected 4 completed steps, got %d: %q", w, len(steps), view)
		}
		want := []string{"Bob", long, "Yes", "Run 'make install' in ./my-app to complete the setup"}
		for i, s := range steps {
			if squash(strings.Join(s.answer, " ")) != squash(want[i]) {
				t.Errorf("width %d step %d: answer altered: %q", w, i, s.answer)
			}
		}
		if len(steps[0].answer) != 1 || len(steps[2].answer) != 1 {
			t.Errorf("width %d: short answers should not wrap: %q / %q", w, steps[0].answer, steps[2].answer)
		}
	}
}

func TestRenderingAtNarrowWidths(t *testing.T) {
	st := newStyles(DefaultTheme())
	for _, w := range []int{20, 40, 80} {
		checkLines := func(name, view string) {
			for _, line := range strings.Split(view, "\n") {
				if lw := lipgloss.Width(line); lw > w {
					t.Errorf("width %d %s: line width %d exceeds terminal: %q", w, name, lw, line)
				}
			}
		}

		text := Text("q", "Query")
		text.Init(st)
		text.UpdateMsg(tea.WindowSizeMsg{Width: w, Height: 24})
		if text.input.Width < 1 || text.input.Width > w-2 {
			t.Errorf("width %d: text input width %d outside available space", w, text.input.Width)
		}

		list := List("l", "Pick").Items(Item(strings.Repeat("LongOption", 10), "meta"), Item("A"))
		list.Init(st)
		list.UpdateMsg(tea.WindowSizeMsg{Width: w, Height: 24})
		checkLines("list", list.View(st))

		info := Info("i", "Info").Body("Configuration was written successfully to ./config/settings.yaml END")
		info.UpdateMsg(tea.WindowSizeMsg{Width: w, Height: 24})
		iv := info.View(st)
		checkLines("info", iv)
		if !strings.Contains(iv, "END") {
			t.Errorf("width %d: info body truncated: %q", w, iv)
		}

		m := newWizardModel([]Step{Text("n", "Name")}, st, true)
		m.width = w
		m.failed = true
		m.failure = errors.New("database connection refused: dial tcp 192.168.1.50:5432 timeout END")
		fv := m.View()
		checkLines("failure", fv)
		if !strings.Contains(squash(ansiSeq.ReplaceAllString(fv, "")), squash(m.failure.Error())) {
			t.Errorf("width %d: failure message truncated: %q", w, fv)
		}

		// Short content must not be padded out to the terminal width.
		short := completedView(t, w, Confirm("ok", "OK?"))
		for _, line := range strings.Split(strings.TrimSuffix(short, "\n"), "\n") {
			if strings.HasSuffix(line, " ") {
				t.Errorf("width %d: short history line padded: %q", w, line)
			}
		}
	}
}
