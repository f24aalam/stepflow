package stepflow

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type InfoStep struct {
	key, question, body string
	next                func(Result) []Step
}

func Info(key, question string) *InfoStep                    { return &InfoStep{key: key, question: question} }
func (s *InfoStep) Body(v string) *InfoStep                  { s.body = v; return s }
func (s *InfoStep) WithNext(f func(Result) []Step) *InfoStep { s.next = f; return s }
func (s *InfoStep) Key() string                              { return s.key }
func (s *InfoStep) Question() string                         { return s.question }
func (s *InfoStep) Init(styles) tea.Cmd                      { return nil }
func (s *InfoStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}
func (s *InfoStep) Update(k tea.KeyMsg) (bool, tea.Cmd) { return k.String() == "enter", nil }
func (s *InfoStep) View(st styles) string {
	body := strings.TrimSpace(s.body)
	if body == "" {
		body = "No details."
	}
	v := ""
	for _, line := range strings.Split(body, "\n") {
		v += "   " + st.stepAnswer.Render(line) + "\n"
	}
	return v + st.helper.Render("enter to continue") + "\n"
}
func (s *InfoStep) ResultView(st styles) string { return st.stepAnswer.Render(s.Answer()) }
func (s *InfoStep) Answer() string              { return "shown" }
