package stepflow

import (
	"context"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type loadTickMsg struct{}
type loadDoneMsg struct{ value string }
type loadErrMsg struct{ err error }

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return loadTickMsg{} })
}

type WorkFunc func(status chan<- string) (string, error)
type ContextWorkFunc func(context.Context, chan<- string) (string, error)
type LoadStep struct {
	key, question  string
	work           WorkFunc
	contextWork    ContextWorkFunc
	ctx            context.Context
	statusCh       chan string
	frame          int
	status, answer string
	finished       bool
	failure        error
	next           func(Result) []Step
}

func Load(key, question string) *LoadStep { return &LoadStep{key: key, question: question} }

// Run attaches a legacy work function. Legacy workers lack context cancellation support
// and cannot be forcibly terminated if the work function does not exit on its own.
// Status updates are handled asynchronously so workers will not stall on full buffers,
// but RunContext is strongly recommended for cancellation-aware background operations.
func (s *LoadStep) Run(fn WorkFunc) *LoadStep { s.work = fn; s.contextWork = nil; return s }

// RunContext attaches a context-aware work function. The worker receives the wizard's context.Context
// and should monitor ctx.Done() for cooperative cancellation. Status updates sent to the channel
// will not block even if sends occur rapidly.
func (s *LoadStep) RunContext(fn ContextWorkFunc) *LoadStep {
	s.contextWork = fn
	s.work = nil
	return s
}
func (s *LoadStep) WithNext(f func(Result) []Step) *LoadStep { s.next = f; return s }
func (s *LoadStep) Key() string                              { return s.key }
func (s *LoadStep) Question() string                         { return s.question }
func (s *LoadStep) Answer() string                           { return s.answer }
func (s *LoadStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}
func (s *LoadStep) reset() {
	s.frame = 0
	s.status = "working…"
	s.answer = ""
	s.finished = false
	s.failure = nil
}
func (s *LoadStep) Init(styles) tea.Cmd {
	s.reset()
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		s.failure = err
		return func() tea.Msg { return loadErrMsg{err} }
	}

	workerCh := make(chan string)
	statusCh := make(chan string, 1)
	s.statusCh = statusCh

	go func() {
		for msg := range workerCh {
			select {
			case statusCh <- msg:
			default:
				select {
				case <-statusCh:
				default:
				}
				select {
				case statusCh <- msg:
				default:
				}
			}
		}
		close(statusCh)
	}()

	return tea.Batch(spinnerTick(), func() tea.Msg {
		var v string
		var err error
		if s.contextWork != nil {
			v, err = s.contextWork(ctx, workerCh)
		} else if s.work != nil {
			v, err = s.work(workerCh)
		} else {
			err = errors.New("stepflow: missing loading worker")
		}
		close(workerCh)
		if err != nil {
			return loadErrMsg{err}
		}
		return loadDoneMsg{v}
	})
}
func (s *LoadStep) UpdateMsg(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case loadTickMsg:
		for {
			select {
			case v, ok := <-s.statusCh:
				if !ok {
					goto drained
				}
				s.status = v
			default:
				goto drained
			}
		}
	drained:
		s.frame = (s.frame + 1) % len(spinnerFrames)
		if s.ctx != nil && s.ctx.Err() != nil && !s.finished && s.failure == nil {
			s.failure = s.ctx.Err()
			return false, nil
		}
		if !s.finished && s.failure == nil {
			return false, spinnerTick()
		}
	case loadDoneMsg:
		s.answer = m.value
		s.finished = true
		return true, nil
	case loadErrMsg:
		s.failure = m.err
	}
	return false, nil
}
func (s *LoadStep) Update(tea.KeyMsg) (bool, tea.Cmd) { return false, nil }
func (s *LoadStep) HasError() bool                    { return s.failure != nil }
func (s *LoadStep) StepError() error                  { return s.failure }
func (s *LoadStep) View(st styles) string {
	if s.failure != nil {
		return "  " + st.error.Render("✗ "+s.failure.Error()) + "\n"
	}
	if s.finished {
		return ""
	}
	return fmt.Sprintf("  %s  %s\n", st.marker.Render(spinnerFrames[s.frame]), st.helper.Render(s.status))
}
func (s *LoadStep) ResultView(st styles) string { return st.stepAnswer.Render(s.answer) }
