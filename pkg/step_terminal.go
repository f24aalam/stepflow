package stepflow

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type terminalKind int

const (
	terminalKindSuccess terminalKind = iota
	terminalKindError
)

type TerminalStep struct {
	key, question, body string
	kind                terminalKind
	next                func(Result) []Step
	width               int
}

func Success(key, q string) *TerminalStep { return &TerminalStep{key: key, question: q} }
func Error(key, q string) *TerminalStep {
	return &TerminalStep{key: key, question: q, kind: terminalKindError}
}
func (s *TerminalStep) Body(v string) *TerminalStep                  { s.body = v; return s }
func (s *TerminalStep) WithNext(f func(Result) []Step) *TerminalStep { s.next = f; return s }
func (s *TerminalStep) Key() string                                  { return s.key }
func (s *TerminalStep) Question() string                             { return s.question }
func (s *TerminalStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}

type termDoneMsg struct{}

func (s *TerminalStep) Init(styles) tea.Cmd {
	if s.next == nil {
		return func() tea.Msg { return termDoneMsg{} }
	}
	return nil
}
func (s *TerminalStep) UpdateMsg(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case termDoneMsg:
		return true, nil
	case tea.KeyMsg:
		if m.String() == "enter" {
			return true, nil
		}
	case tea.WindowSizeMsg:
		s.width = m.Width
		return false, nil
	}
	return false, nil
}
func (s *TerminalStep) HasError() bool                      { return false }
func (s *TerminalStep) Update(k tea.KeyMsg) (bool, tea.Cmd) { return k.String() == "enter", nil }
func (s *TerminalStep) View(st styles) string {
	if s.next == nil {
		return ""
	}
	v := ""
	body := strings.TrimSpace(s.body)
	if body == "" {
		body = "Done."
	}
	for _, line := range strings.Split(body, "\n") {
		if s.width > 6 {
			line = lipgloss.NewStyle().Width(s.width - 4).Render(line)
		}
		for _, sub := range strings.Split(line, "\n") {
			if s.kind == terminalKindError {
				v += "  " + st.error.Render(sub) + "\n"
			} else {
				v += "  " + st.done.Render(sub) + "\n"
			}
		}
	}
	return v + "  " + st.helper.Render("enter to continue") + "\n"
}
func (s *TerminalStep) ResultView(st styles) string {
	if s.kind == terminalKindError {
		return st.error.Render(s.Answer())
	}
	return st.done.Render(s.Answer())
}
func (s *TerminalStep) Answer() string {
	if s.body != "" {
		return s.body
	}
	return "done"
}
