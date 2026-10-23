package stepflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doneMsg struct{}
type wizardModel struct {
	steps                  []Step
	current                int
	answers                []string
	st                     styles
	showDone, done, failed bool
	failure                error
	result                 Result
	ctx                    context.Context
	width, height          int
}

func newWizardModel(steps []Step, st styles, showDone bool, ctx ...context.Context) wizardModel {
	var c context.Context
	if len(ctx) > 0 && ctx[0] != nil {
		c = ctx[0]
	} else {
		c = context.Background()
	}
	return wizardModel{steps: steps, st: st, showDone: showDone, ctx: c}
}
func (m wizardModel) Init() tea.Cmd {
	if len(m.steps) == 0 {
		return func() tea.Msg { return doneMsg{} }
	}
	return m.steps[0].Init(m.st)
}
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		if m.current < len(m.steps) {
			if ms, ok := m.steps[m.current].(MessageStep); ok {
				_, cmd := ms.UpdateMsg(msg)
				return m, cmd
			}
		}
		return m, nil
	}
	if _, ok := msg.(doneMsg); ok {
		m.done = true
		return m, tea.Quit
	}
	if m.done || m.failed {
		return m, nil
	}
	step := m.steps[m.current]
	if ms, ok := step.(MessageStep); ok {
		done, cmd := ms.UpdateMsg(msg)
		if ms.HasError() {
			m.failed = true
			if e, ok := step.(interface{ StepError() error }); ok {
				m.failure = e.StepError()
			}
			if m.failure == nil {
				m.failure = errors.New("stepflow: step failed")
			}
			return m, tea.Quit
		}
		if done {
			cmd := m.advance(step)
			return m, cmd
		}
		return m, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		done, cmd := step.Update(key)
		if done {
			cmd := m.advance(step)
			return m, cmd
		}
		return m, cmd
	}
	if ts, ok := step.(*TextStep); ok {
		return m, ts.UpdateOther(msg)
	}
	return m, nil
}
func (m *wizardModel) advance(step Step) tea.Cmd {
	idx := m.current
	m.answers = append(m.answers, step.Answer())
	completed := make(Result, idx+1)
	for i := 0; i <= idx; i++ {
		completed[m.steps[i].Key()] = m.answers[i]
	}
	if ds, ok := step.(DynamicStep); ok {
		if next := ds.NextSteps(completed); next != nil {
			seen := make(map[string]bool, idx+1+len(next))
			for i := 0; i <= idx; i++ {
				seen[m.steps[i].Key()] = true
			}
			if err := validateStepList(next, seen); err != nil {
				m.failed = true
				m.failure = err
				return tea.Quit
			}
			for _, s := range next {
				if l, ok := s.(*LoadStep); ok {
					l.ctx = m.ctx
				}
			}
			m.steps = append(m.steps[:idx+1:idx+1], next...)
		}
	}
	m.current++
	if m.current >= len(m.steps) {
		m.result = make(Result, len(m.answers))
		for i, a := range m.answers {
			m.result[m.steps[i].Key()] = a
		}
		m.done = true
		return tea.Quit
	}
	nextStep := m.steps[m.current]
	if m.width > 0 {
		if ms, ok := nextStep.(MessageStep); ok {
			ms.UpdateMsg(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}
	return nextStep.Init(m.st)
}
func (m wizardModel) View() string {
	var b strings.Builder
	for i, a := range m.answers {
		step := m.steps[i]
		if _, ok := step.(*LoadStep); ok && a == "" {
			continue
		}
		q := step.Question()
		ans := step.ResultView(m.st)
		if m.width > 6 {
			q = lipgloss.NewStyle().MaxWidth(m.width - 4).Render(q)
			ans = lipgloss.NewStyle().MaxWidth(m.width - 4).Render(ans)
		}
		b.WriteString(m.st.marker.Render("◆") + "  " + m.st.stepLabel.Render(q) + "\n")
		b.WriteString(m.st.pipe.Render("│") + "  " + ans + "\n" + m.st.pipe.Render("│") + "\n")
	}
	if m.failed {
		errText := fmt.Sprintf("✗ %v", m.failure)
		if m.width > 6 {
			errText = lipgloss.NewStyle().Width(m.width - 2).Render(errText)
		}
		return b.String() + "\n" + m.st.error.Render(errText) + "\n"
	}
	if m.done {
		if m.showDone {
			b.WriteString(m.st.done.Render("✓  All done!") + "\n")
		}
		return b.String()
	}
	if m.current < len(m.steps) {
		s := m.steps[m.current]
		q := s.Question()
		if m.width > 6 {
			q = lipgloss.NewStyle().Width(m.width - 4).Render(q)
		}
		b.WriteString(m.st.marker.Render("◆") + "  " + m.st.activeQ.Render(q) + "\n" + s.View(m.st))
	}
	return b.String()
}
