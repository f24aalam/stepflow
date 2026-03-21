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

// DefaultTheme returns a green-only palette (no gray/red/blue accents).
func DefaultTheme() Theme {
	return Theme{
		MarkerColor:         lipgloss.Color("#4ade80"),
		PipeColor:           lipgloss.Color("#15803d"),
		StepLabelColor:      lipgloss.Color("#22c55e"),
		StepAnswerColor:     lipgloss.Color("#bbf7d0"),
		ActiveQuestionColor: lipgloss.Color("#ecfdf5"),
		SectionLineColor:    lipgloss.Color("#14532d"),
		SectionLabelColor:   lipgloss.Color("#166534"),
		AlwaysNoteColor:     lipgloss.Color("#4ade80"),
		SelectedDotColor:    lipgloss.Color("#4ade80"),
		UnselectedDotColor:  lipgloss.Color("#14532d"),
		FocusBarColor:       lipgloss.Color("#4ade80"),
		FocusLabelColor:     lipgloss.Color("#4ade80"),
		ItemPathColor:       lipgloss.Color("#15803d"),
		SearchSlashColor:    lipgloss.Color("#4ade80"),
		SearchTextColor:     lipgloss.Color("#ecfdf5"),
		SearchCursorFg:      lipgloss.Color("#052e16"),
		SearchCursorBg:      lipgloss.Color("#4ade80"),
		ScrollPosColor:      lipgloss.Color("#22c55e"),
		ScrollTotalColor:    lipgloss.Color("#166534"),
		PillBracketColor:    lipgloss.Color("#15803d"),
		PillItemColor:       lipgloss.Color("#4ade80"),
		PillDotColor:        lipgloss.Color("#22c55e"),
		PillMoreColor:       lipgloss.Color("#86efac"),
		PillNoneColor:       lipgloss.Color("#14532d"),
		HelperTextColor:     lipgloss.Color("#22c55e"),
		DoneColor:           lipgloss.Color("#4ade80"),
	}
}
