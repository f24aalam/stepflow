package stepflow

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// InfoStep renders read-only details and continues on Enter.
type InfoStep struct {
	key      string
	question string
	body     string
	next     func(Result) []Step
}

// Info creates a non-interactive details step.
func Info(key, question string) *InfoStep {
	return &InfoStep{key: key, question: question}
}

// Body sets the details text shown in the step.
func (s *InfoStep) Body(text string) *InfoStep {
	s.body = text
	return s
}

// WithNext attaches a dynamic step planner.
func (s *InfoStep) WithNext(fn func(Result) []Step) *InfoStep {
	s.next = fn
	return s
}

func (s *InfoStep) Key() string           { return s.key }
func (s *InfoStep) Question() string      { return s.question }
func (s *InfoStep) Init(_ styles) tea.Cmd { return nil }

// NextSteps satisfies the DynamicStep interface.
func (s *InfoStep) NextSteps(completed Result) []Step {
	if s.next != nil {
		return s.next(completed)
	}
	return nil
}

func (s *InfoStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.String() == "enter" {
		return true, nil
	}

	return false, nil
}

func (s *InfoStep) View(st styles) string {
	body := strings.TrimSpace(s.body)
	if body == "" {
		body = "No details."
	}

	lines := strings.Split(body, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString("   " + st.stepAnswer.Render(line) + "\n")
	}

	b.WriteString("\n" + st.helper.Render("enter to continue") + "\n")

	return b.String()
}

func (s *InfoStep) ResultView(st styles) string {
	return st.stepAnswer.Render(s.Answer())
}

func (s *InfoStep) Answer() string {
	return "shown"
}
