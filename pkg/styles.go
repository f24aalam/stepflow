package stepflow

import "github.com/charmbracelet/lipgloss"

// styles holds computed lipgloss.Style values derived from a Theme.
// Rebuilt each time the wizard runs so callers can swap themes freely.
type styles struct {
	marker      lipgloss.Style
	pipe        lipgloss.Style
	stepLabel   lipgloss.Style
	stepAnswer  lipgloss.Style
	activeQ     lipgloss.Style
	sectionLine lipgloss.Style
	sectionLbl  lipgloss.Style
	alwaysNote  lipgloss.Style
	selDot      lipgloss.Style
	unselDot    lipgloss.Style
	focusBar    lipgloss.Style
	focusLabel  lipgloss.Style
	itemPath    lipgloss.Style
	searchSlash lipgloss.Style
	searchText  lipgloss.Style
	searchCur   lipgloss.Style
	scrollPos   lipgloss.Style
	scrollTotal lipgloss.Style
	pillBracket lipgloss.Style
	pillItem    lipgloss.Style
	pillDot     lipgloss.Style
	pillMore    lipgloss.Style
	pillNone    lipgloss.Style
	helper      lipgloss.Style
	done        lipgloss.Style
}

func newStyles(t Theme) styles {
	return styles{
		marker:      lipgloss.NewStyle().Foreground(t.MarkerColor),
		pipe:        lipgloss.NewStyle().Foreground(t.PipeColor),
		stepLabel:   lipgloss.NewStyle().Foreground(t.StepLabelColor),
		stepAnswer:  lipgloss.NewStyle().Foreground(t.StepAnswerColor),
		activeQ:     lipgloss.NewStyle().Foreground(t.ActiveQuestionColor).Bold(true),
		sectionLine: lipgloss.NewStyle().Foreground(t.SectionLineColor),
		sectionLbl:  lipgloss.NewStyle().Foreground(t.SectionLabelColor),
		alwaysNote:  lipgloss.NewStyle().Foreground(t.AlwaysNoteColor),
		selDot:      lipgloss.NewStyle().Foreground(t.SelectedDotColor),
		unselDot:    lipgloss.NewStyle().Foreground(t.UnselectedDotColor),
		focusBar:    lipgloss.NewStyle().Foreground(t.FocusBarColor),
		focusLabel:  lipgloss.NewStyle().Foreground(t.FocusLabelColor),
		itemPath:    lipgloss.NewStyle().Foreground(t.ItemPathColor),
		searchSlash: lipgloss.NewStyle().Foreground(t.SearchSlashColor).Bold(true),
		searchText:  lipgloss.NewStyle().Foreground(t.SearchTextColor),
		searchCur:   lipgloss.NewStyle().Foreground(t.SearchCursorFg).Background(t.SearchCursorBg),
		scrollPos:   lipgloss.NewStyle().Foreground(t.ScrollPosColor),
		scrollTotal: lipgloss.NewStyle().Foreground(t.ScrollTotalColor),
		pillBracket: lipgloss.NewStyle().Foreground(t.PillBracketColor),
		pillItem:    lipgloss.NewStyle().Foreground(t.PillItemColor),
		pillDot:     lipgloss.NewStyle().Foreground(t.PillDotColor),
		pillMore:    lipgloss.NewStyle().Foreground(t.PillMoreColor).Faint(true),
		pillNone:    lipgloss.NewStyle().Foreground(t.PillNoneColor).Italic(true),
		helper:      lipgloss.NewStyle().Foreground(t.HelperTextColor),
		done:        lipgloss.NewStyle().Foreground(t.DoneColor).Bold(true),
	}
}
