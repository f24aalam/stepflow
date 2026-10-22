package stepflow

import "github.com/charmbracelet/lipgloss"

// Theme configures the colors used throughout the wizard.
type Theme struct {
	MarkerColor, PipeColor, StepLabelColor, StepAnswerColor                             lipgloss.Color
	ActiveQuestionColor, SectionLineColor, SectionLabelColor, AlwaysNoteColor           lipgloss.Color
	SelectedDotColor, UnselectedDotColor, FocusBarColor, FocusLabelColor, ItemPathColor lipgloss.Color
	SearchSlashColor, SearchTextColor, SearchCursorFg, SearchCursorBg                   lipgloss.Color
	ScrollPosColor, ScrollTotalColor, PillBracketColor, PillItemColor                   lipgloss.Color
	PillDotColor, PillMoreColor, PillNoneColor, HelperTextColor, DoneColor, ErrorColor  lipgloss.Color
}

func DefaultTheme() Theme {
	return Theme{
		MarkerColor: "#4ade80", PipeColor: "#15803d", StepLabelColor: "#22c55e", StepAnswerColor: "#bbf7d0", ActiveQuestionColor: "#ecfdf5",
		SectionLineColor: "#14532d", SectionLabelColor: "#166534", AlwaysNoteColor: "#4ade80", SelectedDotColor: "#4ade80", UnselectedDotColor: "#14532d",
		FocusBarColor: "#4ade80", FocusLabelColor: "#4ade80", ItemPathColor: "#15803d", SearchSlashColor: "#4ade80", SearchTextColor: "#ecfdf5",
		SearchCursorFg: "#052e16", SearchCursorBg: "#4ade80", ScrollPosColor: "#22c55e", ScrollTotalColor: "#166534", PillBracketColor: "#15803d",
		PillItemColor: "#4ade80", PillDotColor: "#22c55e", PillMoreColor: "#86efac", PillNoneColor: "#14532d", HelperTextColor: "#22c55e", DoneColor: "#4ade80", ErrorColor: "#ef4444",
	}
}
