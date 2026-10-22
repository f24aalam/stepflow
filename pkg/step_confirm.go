package stepflow

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type ConfirmStep struct {
	key, question string
	cursor        int
	next          func(Result) []Step
}

func Confirm(key, question string) *ConfirmStep { return &ConfirmStep{key: key, question: question} }
func (s *ConfirmStep) Default(d string) *ConfirmStep {
	if strings.EqualFold(d, "no") {
		s.cursor = 1
	} else {
		s.cursor = 0
	}
	return s
}
func (s *ConfirmStep) WithNext(f func(Result) []Step) *ConfirmStep { s.next = f; return s }
func (s *ConfirmStep) Key() string                                 { return s.key }
func (s *ConfirmStep) Question() string                            { return s.question }
func (s *ConfirmStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}
func (s *ConfirmStep) Init(styles) tea.Cmd { return nil }
func (s *ConfirmStep) Update(k tea.KeyMsg) (bool, tea.Cmd) {
	switch k.String() {
	case "up", "k":
		s.cursor = 0
	case "down", "j":
		s.cursor = 1
	case "enter":
		return true, nil
	}
	return false, nil
}
func (s *ConfirmStep) View(st styles) string {
	v := ""
	for i, label := range []string{"Yes", "No"} {
		if i == s.cursor {
			v += "  " + st.focusBar.Render("›") + "  " + st.focusLabel.Render(label) + "\n"
		} else {
			v += "     " + st.stepAnswer.Render(label) + "\n"
		}
	}
	return v + st.helper.Render("↑↓ move · enter confirm") + "\n"
}
func (s *ConfirmStep) ResultView(st styles) string { return st.stepAnswer.Render(s.Answer()) }
func (s *ConfirmStep) Answer() string {
	if s.cursor == 0 {
		return "Yes"
	}
	return "No"
}
