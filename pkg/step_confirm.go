package stepflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmStep is a two-option Yes / No selection step.
type ConfirmStep struct {
	key      string
	question string
	cursor   int // 0 = Yes, 1 = No
	defVal   string
}

// Confirm creates a new yes/no confirmation step.
func Confirm(key, question string) *ConfirmStep {
	return &ConfirmStep{key: key, question: question, defVal: "Yes"}
}

// Default sets which option is pre-selected ("Yes" or "No").
func (s *ConfirmStep) Default(d string) *ConfirmStep {
	if strings.EqualFold(d, "no") {
		s.cursor = 1
	} else {
		s.cursor = 0
	}
	s.defVal = d
	return s
}

func (s *ConfirmStep) Key() string           { return s.key }
func (s *ConfirmStep) Question() string      { return s.question }
func (s *ConfirmStep) Init(_ styles) tea.Cmd { return nil }

func (s *ConfirmStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
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
	var b strings.Builder
	for i, opt := range []string{"Yes", "No"} {
		if i == s.cursor {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				st.focusBar.Render("›"),
				st.focusLabel.Render(opt),
			))
		} else {
			b.WriteString(fmt.Sprintf("     %s\n", st.stepAnswer.Render(opt)))
		}
	}
	b.WriteString(st.helper.Render("↑↓ move · enter confirm") + "\n")

	return b.String()
}

func (s *ConfirmStep) Answer() string {
	if s.cursor == 0 {
		return "Yes"
	}

	return "No"
}
