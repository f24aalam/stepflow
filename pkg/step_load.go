package stepflow

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── internal messages ─────────────────────────────────────────────────────────

type loadTickMsg struct{}
type loadStatusMsg string
type loadDoneMsg struct{ value string }
type loadErrMsg struct{ err error }

// ── spinner ───────────────────────────────────────────────────────────────────

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
		return loadTickMsg{}
	})
}

// ── WorkFunc ──────────────────────────────────────────────────────────────────

// WorkFunc is the signature for background work passed to a LoadStep.
// Send human-readable status strings to the status channel to update the
// spinner message live. Return a result string and any error when done.
//
//	func(status chan<- string) (string, error) {
//	    status <- "resolving host…"
//	    v, err := fetchVersion()
//	    return v, err
//	}
type WorkFunc func(status chan<- string) (string, error)

// ── LoadStep ──────────────────────────────────────────────────────────────────

// LoadStep displays an animated spinner + live status message while running a
// background WorkFunc. On success it auto-advances to the next step. On error
// it freezes with a red ✗ and halts the wizard.
type LoadStep struct {
	key      string
	question string
	work     WorkFunc

	// runtime state (populated in Init)
	statusCh chan string
	frame    int
	status   string
	answer   string
	finished bool
	failed   bool
	errMsg   string
}

// Load creates a new loading step.
//
// Usage:
//
//	stepflow.Load("version", "Fetching latest release").
//	    Run(func(status chan<- string) (string, error) {
//	        status <- "contacting registry…"
//	        v, err := fetchLatestVersion()
//	        return v, err
//	    })
func Load(key, question string) *LoadStep {
	return &LoadStep{key: key, question: question}
}

// Run attaches the work function and returns the step.
func (s *LoadStep) Run(fn WorkFunc) *LoadStep {
	s.work = fn
	return s
}

func (s *LoadStep) Key() string      { return s.key }
func (s *LoadStep) Question() string { return s.question }
func (s *LoadStep) Answer() string   { return s.answer }

// Init launches the work goroutine and kicks off the spinner tick.
func (s *LoadStep) Init(st styles) tea.Cmd {
	s.status = "working…"
	s.frame = 0
	s.finished = false
	s.failed = false
	s.statusCh = make(chan string, 16)

	ch := s.statusCh

	workCmd := func() tea.Msg {
		val, err := s.work(ch)
		close(ch)
		if err != nil {
			return loadErrMsg{err}
		}
		return loadDoneMsg{val}
	}

	return tea.Batch(spinnerTick(), tea.Cmd(workCmd))
}

// UpdateMsg satisfies MessageStep — handles spinner ticks and goroutine results.
func (s *LoadStep) UpdateMsg(msg tea.Msg) (done bool, cmd tea.Cmd) {
	switch m := msg.(type) {

	case loadTickMsg:
		for {
			select {
			case txt, open := <-s.statusCh:
				if !open {
					goto drained
				}
				s.status = txt
			default:
				goto drained
			}
		}
	drained:
		if !s.finished && !s.failed {
			s.frame = (s.frame + 1) % len(spinnerFrames)
			return false, spinnerTick()
		}

	case loadDoneMsg:
		s.answer = m.value
		s.finished = true
		return true, nil

	case loadErrMsg:
		s.failed = true
		s.errMsg = m.err.Error()
		return false, nil
	}

	return false, nil
}

// Update satisfies the Step interface for key messages (no-op while loading).
func (s *LoadStep) Update(_ tea.KeyMsg) (bool, tea.Cmd) { return false, nil }

// HasError reports whether the work func returned an error.
func (s *LoadStep) HasError() bool { return s.failed }

// View renders the spinner frame + current status message.
func (s *LoadStep) View(st styles) string {
	var b strings.Builder

	if s.failed {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
		b.WriteString(fmt.Sprintf("  %s  %s\n\n",
			errStyle.Render("✗"),
			errStyle.Render(s.errMsg),
		))

		return b.String()
	}

	if s.finished {
		return ""
	}

	spinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b6b6b"))

	b.WriteString(fmt.Sprintf("  %s  %s\n\n",
		spinStyle.Render(spinnerFrames[s.frame]),
		statusStyle.Render(s.status),
	))

	return b.String()
}
