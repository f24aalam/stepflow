package stepflow

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type TextStep struct {
	key, question, placeholder, defaultVal string
	input                                  textinput.Model
	next                                   func(Result) []Step
}

func Text(key, question string) *TextStep          { return &TextStep{key: key, question: question} }
func (s *TextStep) Placeholder(p string) *TextStep { s.placeholder = p; return s }
func (s *TextStep) Default(d string) *TextStep {
	s.defaultVal = d
	if s.placeholder == "" {
		s.placeholder = d
	}
	return s
}
func (s *TextStep) WithNext(f func(Result) []Step) *TextStep { s.next = f; return s }
func (s *TextStep) Key() string                              { return s.key }
func (s *TextStep) Question() string                         { return s.question }
func (s *TextStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}
func (s *TextStep) Init(st styles) tea.Cmd {
	ti := textinput.New()
	ti.Placeholder = s.placeholder
	ti.CharLimit = 2048
	ti.Width = 72
	ti.PromptStyle = ti.PromptStyle.Foreground(st.marker.GetForeground())
	ti.TextStyle = ti.TextStyle.Foreground(st.stepAnswer.GetForeground())
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
func (s *TextStep) UpdateOther(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}
func (s *TextStep) View(st styles) string {
	return s.input.View() + "\n" + st.helper.Render("enter to confirm") + "\n"
}
func (s *TextStep) ResultView(st styles) string { return st.stepAnswer.Render(s.Answer()) }
func (s *TextStep) Answer() string {
	v := strings.TrimSpace(s.input.Value())
	if v == "" {
		return s.defaultVal
	}
	return v
}
