package stepflow

import tea "github.com/charmbracelet/bubbletea"

// Step represents one interactive or informational wizard page.
// Custom external steps are intentionally not supported yet: rendering styles are internal.
type Step interface {
	Key() string
	Question() string
	Init(styles) tea.Cmd
	Update(tea.KeyMsg) (bool, tea.Cmd)
	View(styles) string
	ResultView(styles) string
	Answer() string
}
type MessageStep interface {
	Step
	UpdateMsg(tea.Msg) (bool, tea.Cmd)
	HasError() bool
}
type DynamicStep interface {
	Step
	NextSteps(Result) []Step
}

type resetter interface {
	reset()
}
