package stepflow

import tea "github.com/charmbracelet/bubbletea"

// Step is the common interface every wizard step must implement.
type Step interface {
	// Key is the map key in the final Result.
	Key() string
	// Question is the prompt shown to the user.
	Question() string
	// Init is called once before the step becomes active.
	Init(s styles) tea.Cmd
	// Update handles a key message while this step is active.
	// Returns (done, cmd): done=true means the step is complete.
	Update(msg tea.KeyMsg) (done bool, cmd tea.Cmd)
	// View renders the interactive portion of the step.
	View(s styles) string
	// Answer returns the final string value after the step completes.
	Answer() string
}

// MessageStep is implemented by steps that need to handle arbitrary tea.Msg
// values beyond key presses (e.g. spinner ticks, goroutine completion signals).
// The wizardModel checks for this interface and routes all messages to UpdateMsg.
type MessageStep interface {
	Step
	UpdateMsg(msg tea.Msg) (done bool, cmd tea.Cmd)
	HasError() bool
}

// DynamicStep allows a step to decide which steps should run next.
// When a DynamicStep completes, the wizard replaces the remaining step tail
// with the slice returned by NextSteps.
type DynamicStep interface {
	Step
	NextSteps(completed Result) []Step
}
