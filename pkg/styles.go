package stepflow

import "github.com/charmbracelet/lipgloss"

type styles struct {
	marker      lipgloss.Style
	pipe        lipgloss.Style
	stepLabel   lipgloss.Style
	stepAnswer  lipgloss.Style
	activeQ     lipgloss.Style
	selDot      lipgloss.Style
	unselDot    lipgloss.Style
	focusBar    lipgloss.Style
	focusLabel  lipgloss.Style
	itemPath    lipgloss.Style
	searchSlash lipgloss.Style
	searchText  lipgloss.Style
	searchCur   lipgloss.Style
	pillItem    lipgloss.Style
	helper      lipgloss.Style
	done        lipgloss.Style
	error       lipgloss.Style
}

func newStyles(t Theme) styles {
	color := func(c lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

	return styles{
		marker:      color(t.Primary),
		pipe:        color(t.Muted),
		stepLabel:   color(t.Primary),
		stepAnswer:  color(t.Text),
		activeQ:     color(t.Text).Bold(true),
		selDot:      color(t.Accent),
		unselDot:    color(t.Muted),
		focusBar:    color(t.Accent),
		focusLabel:  color(t.Accent),
		itemPath:    color(t.Muted),
		searchSlash: color(t.Primary),
		searchText:  color(t.Text),
		searchCur:   lipgloss.NewStyle().Foreground(t.Muted).Background(t.Accent),
		pillItem:    color(t.Accent),
		helper:      color(t.Muted),
		done:        color(t.Success),
		error:       color(t.Error),
	}
}
