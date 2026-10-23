package stepflow

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
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
}

func newWizardModel(steps []Step, st styles, showDone bool) wizardModel {
	return wizardModel{steps: steps, st: st, showDone: showDone}
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
			return m, m.advance(step)
		}
		return m, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		done, cmd := step.Update(key)
		if done {
			return m, m.advance(step)
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
	return m.steps[m.current].Init(m.st)
}
func (m wizardModel) View() string {
	var b strings.Builder
	for i, a := range m.answers {
		step := m.steps[i]
		if _, ok := step.(*LoadStep); ok && a == "" {
			continue
		}
		b.WriteString(m.st.marker.Render("◆") + "  " + m.st.stepLabel.Render(step.Question()) + "\n")
		b.WriteString(m.st.pipe.Render("│") + "  " + step.ResultView(m.st) + "\n" + m.st.pipe.Render("│") + "\n")
	}
	if m.failed {
		return b.String() + "\n" + m.st.error.Render(fmt.Sprintf("✗ %v", m.failure)) + "\n"
	}
	if m.done {
		if m.showDone {
			b.WriteString(m.st.done.Render("✓  All done!") + "\n")
		}
		return b.String()
	}
	if m.current < len(m.steps) {
		s := m.steps[m.current]
		b.WriteString(m.st.marker.Render("◆") + "  " + m.st.activeQ.Render(s.Question()) + "\n" + s.View(m.st))
	}
	return b.String()
}
