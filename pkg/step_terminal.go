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

func (s *TerminalStep) Key() string           { return s.key }
func (s *TerminalStep) Question() string      { return s.question }
func (s *TerminalStep) Init(_ styles) tea.Cmd { return nil }

func (s *TerminalStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.String() == "enter" {
		return true, nil
	}

	return false, nil
}

func (s *TerminalStep) View(st styles) string {
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

func (s *TerminalStep) Answer() string {
	return "acknowledged"
}
