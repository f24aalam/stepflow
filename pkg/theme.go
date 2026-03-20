package stepflow

import "github.com/charmbracelet/lipgloss"

// Theme controls every color in the wizard.
type Theme struct {
	// Markers & connectors
	MarkerColor lipgloss.Color
	PipeColor   lipgloss.Color

	// Completed step
	StepLabelColor  lipgloss.Color
	StepAnswerColor lipgloss.Color

	// Active question
	ActiveQuestionColor lipgloss.Color

	// Section dividers
	SectionLineColor  lipgloss.Color
	SectionLabelColor lipgloss.Color
	AlwaysNoteColor   lipgloss.Color

	// List items
	SelectedDotColor   lipgloss.Color
	UnselectedDotColor lipgloss.Color
	FocusBarColor      lipgloss.Color
	FocusLabelColor    lipgloss.Color
	ItemPathColor      lipgloss.Color

	// Search bar
	SearchSlashColor lipgloss.Color
	SearchTextColor  lipgloss.Color
	SearchCursorFg   lipgloss.Color
	SearchCursorBg   lipgloss.Color

	// Scroll counter
	ScrollPosColor   lipgloss.Color
	ScrollTotalColor lipgloss.Color

	// Selected pill
	PillBracketColor lipgloss.Color
	PillItemColor    lipgloss.Color
	PillDotColor     lipgloss.Color
	PillMoreColor    lipgloss.Color
	PillNoneColor    lipgloss.Color

	// Misc
	HelperTextColor lipgloss.Color
	DoneColor       lipgloss.Color
}

// DefaultTheme returns the green-on-dark agentsync theme.
func DefaultTheme() Theme {
	return Theme{
		MarkerColor:         lipgloss.Color("#4ade80"),
		PipeColor:           lipgloss.Color("#6b6b6b"),
		StepLabelColor:      lipgloss.Color("#555555"),
		StepAnswerColor:     lipgloss.Color("#e5e5e5"),
		ActiveQuestionColor: lipgloss.Color("#e5e5e5"),
		SectionLineColor:    lipgloss.Color("#2a2a2a"),
		SectionLabelColor:   lipgloss.Color("#3a3a3a"),
		AlwaysNoteColor:     lipgloss.Color("#2a6648"),
		SelectedDotColor:    lipgloss.Color("#4ade80"),
		UnselectedDotColor:  lipgloss.Color("#333333"),
		FocusBarColor:       lipgloss.Color("#4ade80"),
		FocusLabelColor:     lipgloss.Color("#4ade80"),
		ItemPathColor:       lipgloss.Color("#3a3a3a"),
		SearchSlashColor:    lipgloss.Color("#4ade80"),
		SearchTextColor:     lipgloss.Color("#e5e5e5"),
		SearchCursorFg:      lipgloss.Color("#000000"),
		SearchCursorBg:      lipgloss.Color("#4ade80"),
		ScrollPosColor:      lipgloss.Color("#3a3a3a"),
		ScrollTotalColor:    lipgloss.Color("#2a2a2a"),
		PillBracketColor:    lipgloss.Color("#2a6648"),
		PillItemColor:       lipgloss.Color("#4ade80"),
		PillDotColor:        lipgloss.Color("#2a6648"),
		PillMoreColor:       lipgloss.Color("#4ade80"),
		PillNoneColor:       lipgloss.Color("#2a2a2a"),
		HelperTextColor:     lipgloss.Color("#4a4a4a"),
		DoneColor:           lipgloss.Color("#4ade80"),
	}
}
