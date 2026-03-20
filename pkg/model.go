package stepflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// doneMsg signals that all steps are complete.
type doneMsg struct{}

// wizardModel is the internal bubbletea model.
type wizardModel struct {
	steps   []Step
	current int
	answers []string // one per completed step
	st      styles
	done    bool
	failed  bool
	result  Result
}

func newWizardModel(steps []Step, st styles) wizardModel {
	return wizardModel{
		steps: steps,
		st:    st,
	}
}

func (m wizardModel) Init() tea.Cmd {
	if len(m.steps) == 0 {
		return func() tea.Msg { return doneMsg{} }
	}
	return m.steps[0].Init(m.st)
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── global quit ───────────────────────────────────────────────────────────
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
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

	// ── MessageStep: route all tea.Msg (spinner ticks, goroutine results) ────
	if ms, ok := step.(MessageStep); ok {
		done, cmd := ms.UpdateMsg(msg)

		if ms.HasError() {
			m.failed = true
			return m, nil
		}

		if done {
			return m, m.advance(step)
		}

		return m, cmd
	}

	// ── Key-only steps ────────────────────────────────────────────────────────
	if key, ok := msg.(tea.KeyMsg); ok {
		done, cmd := step.Update(key)
		if done {
			return m, m.advance(step)
		}

		return m, cmd
	}

	// ── Forward other messages to TextStep (blink, etc.) ─────────────────────
	if ts, ok := step.(*TextStep); ok {
		var cmd tea.Cmd
		ts.input, cmd = ts.input.Update(msg)

		return m, tea.Batch(cmd, ts.input.Focus())
	}

	return m, nil
}

// advance records the answer and moves to the next step (or finishes).
func (m *wizardModel) advance(step Step) tea.Cmd {
	m.answers = append(m.answers, step.Answer())
	m.current++
	if m.current >= len(m.steps) {
		m.result = make(Result, len(m.steps))
		for i, s := range m.steps {
			m.result[s.Key()] = m.answers[i]
		}

		m.done = true
		return tea.Quit
	}

	return m.steps[m.current].Init(m.st)
}

func (m wizardModel) View() string {
	var b strings.Builder

	// Completed steps — collapsed into history
	for i, answer := range m.answers {
		step := m.steps[i]
		// Skip loading steps with empty answers from history
		if answer == "" && isLoadStep(step) {
			continue
		}

		b.WriteString(fmt.Sprintf("%s  %s\n",
			m.st.marker.Render("◆"),
			m.st.stepLabel.Render(padLabel(step.Question())),
		))

		b.WriteString(fmt.Sprintf("%s  %s\n",
			m.st.pipe.Render("│"),
			m.st.stepAnswer.Render(answer),
		))

		b.WriteString(m.st.pipe.Render("│") + "\n")
	}

	if m.done {
		b.WriteString("\n")
		b.WriteString(m.st.done.Render("✓  All done!"))
		b.WriteString("\n\n")
		b.WriteString(m.st.helper.Render("press any key to exit") + "\n")

		return b.String()
	}

	if m.current < len(m.steps) {
		step := m.steps[m.current]

		if ls, ok := step.(*LoadStep); ok {
			// Loading step: render question + spinner inline (no ◆ prefix repeated)
			b.WriteString(fmt.Sprintf("%s  %s\n",
				m.st.marker.Render("◆"),
				m.st.activeQ.Render(step.Question()),
			))
			b.WriteString(ls.View(m.st))
		} else {
			b.WriteString(fmt.Sprintf("%s  %s\n",
				m.st.marker.Render("◆"),
				m.st.activeQ.Render(step.Question()),
			))
			b.WriteString(step.View(m.st))
		}
	}

	return b.String()
}

func isLoadStep(s Step) bool {
	_, ok := s.(*LoadStep)
	return ok
}

// padLabel right-pads short questions so the history column aligns nicely.
func padLabel(q string) string {
	const width = 20
	if len(q) < width {
		return q + strings.Repeat(" ", width-len(q))
	}

	return q
}
