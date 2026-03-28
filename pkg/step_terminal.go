package stepflow

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type terminalKind int

const (
	terminalKindSuccess terminalKind = iota
	terminalKindError
)

// TerminalStep renders a final success/error message and exits on Enter.
type TerminalStep struct {
	key      string
	question string
	body     string
	kind     terminalKind
	next     func(Result) []Step
}

// Success renders a final success message in primary/done color.
func Success(key, question string) *TerminalStep {
	return &TerminalStep{key: key, question: question, kind: terminalKindSuccess}
}

// Error renders a final error message in red.
func Error(key, question string) *TerminalStep {
	return &TerminalStep{key: key, question: question, kind: terminalKindError}
}

func (s *TerminalStep) Body(text string) *TerminalStep {
	s.body = text
	return s
}

// WithNext attaches a dynamic step planner.
func (s *TerminalStep) WithNext(fn func(Result) []Step) *TerminalStep {
	s.next = fn
	return s
}

func (s *TerminalStep) Key() string      { return s.key }
func (s *TerminalStep) Question() string { return s.question }

// NextSteps satisfies the DynamicStep interface.
func (s *TerminalStep) NextSteps(completed Result) []Step {
	if s.next != nil {
		return s.next(completed)
	}
	return nil
}

type termDoneMsg struct{}

func (s *TerminalStep) Init(_ styles) tea.Cmd {
	// If no next step is provided, we auto-advance (finish) instantly.
	if s.next == nil {
		return func() tea.Msg { return termDoneMsg{} }
	}
	return nil
}

// UpdateMsg satisfies MessageStep — allows auto-completion.
func (s *TerminalStep) UpdateMsg(msg tea.Msg) (bool, tea.Cmd) {
	if _, ok := msg.(termDoneMsg); ok {
		return true, nil
	}
	return false, nil
}

func (s *TerminalStep) HasError() bool { return false }

func (s *TerminalStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.String() == "enter" {
		return true, nil
	}

	return false, nil
}

func (s *TerminalStep) View(st styles) string {
	if s.next == nil {
		// When auto-completing, we render nothing in active view;
		// it immediately collapses to history view which uses Answer().
		return ""
	}

	body := strings.TrimSpace(s.body)
	if body == "" {
		body = "Done."
	}

	lines := strings.Split(body, "\n")
	var b strings.Builder

	for _, line := range lines {
		if s.kind == terminalKindError {
			b.WriteString("   " + st.error.Render(line) + "\n")
		} else {
			b.WriteString("   " + st.done.Render(line) + "\n")
		}
	}
	b.WriteString("\n" + st.helper.Render("enter to exit") + "\n")

	return b.String()
}

func (s *TerminalStep) ResultView(st styles) string {
	body := strings.TrimSpace(s.body)
	if body == "" {
		body = "Done."
	}

	if s.kind == terminalKindError {
		return st.error.Render(body)
	}

	return st.done.Render(body)
}

func (s *TerminalStep) Answer() string {
	if s.body != "" {
		return s.body
	}

	return "done"
}
