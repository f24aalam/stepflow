package stepflow

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextStep is a free-text input step.
type TextStep struct {
	key         string
	question    string
	placeholder string
	defaultVal  string
	input       textinput.Model
}

// Text creates a new free-text step.
func Text(key, question string) *TextStep {
	return &TextStep{key: key, question: question}
}

// Placeholder sets the input placeholder text.
func (s *TextStep) Placeholder(p string) *TextStep {
	s.placeholder = p
	return s
}

// Default sets the value used when the user submits an empty input.
func (s *TextStep) Default(d string) *TextStep {
	s.defaultVal = d
	if s.placeholder == "" {
		s.placeholder = d
	}
	return s
}

func (s *TextStep) Key() string      { return s.key }
func (s *TextStep) Question() string { return s.question }

func (s *TextStep) Init(st styles) tea.Cmd {
	ti := textinput.New()
	ti.Placeholder = s.placeholder
	ti.CharLimit = 2048
	ti.Width = 72
	ti.PromptStyle = lipgloss.NewStyle().Foreground(st.marker.GetForeground())
	ti.TextStyle = lipgloss.NewStyle().Foreground(st.stepAnswer.GetForeground())
	ti.SetValue(s.defaultVal)
	s.input = ti

	return tea.Batch(textinput.Blink, s.input.Focus())
}

func (s *TextStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.String() == "enter" {
		return true, nil
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	return false, cmd
}

func (s *TextStep) View(st styles) string {
	return s.input.View() + "\n" +
		"\n" + st.helper.Render("enter to confirm") + "\n"
}

func (s *TextStep) Answer() string {
	val := strings.TrimSpace(s.input.Value())
	if val == "" {
		return s.defaultVal
	}

	return val
}
