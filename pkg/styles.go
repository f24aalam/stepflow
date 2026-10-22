package stepflow

import "github.com/charmbracelet/lipgloss"

type styles struct{ marker, pipe, stepLabel, stepAnswer, activeQ, sectionLine, sectionLbl, alwaysNote, selDot, unselDot, focusBar, focusLabel, itemPath, searchSlash, searchText, searchCur, scrollPos, scrollTotal, pillBracket, pillItem, pillDot, pillMore, pillNone, helper, done, error lipgloss.Style }

func newStyles(t Theme) styles {
	color := func(c lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }
	return styles{marker: color(t.MarkerColor), pipe: color(t.PipeColor), stepLabel: color(t.StepLabelColor), stepAnswer: color(t.StepAnswerColor), activeQ: color(t.ActiveQuestionColor), sectionLine: color(t.SectionLineColor), sectionLbl: color(t.SectionLabelColor), alwaysNote: color(t.AlwaysNoteColor), selDot: color(t.SelectedDotColor), unselDot: color(t.UnselectedDotColor), focusBar: color(t.FocusBarColor), focusLabel: color(t.FocusLabelColor), itemPath: color(t.ItemPathColor), searchSlash: color(t.SearchSlashColor), searchText: color(t.SearchTextColor), searchCur: lipgloss.NewStyle().Foreground(t.SearchCursorFg).Background(t.SearchCursorBg), scrollPos: color(t.ScrollPosColor), scrollTotal: color(t.ScrollTotalColor), pillBracket: color(t.PillBracketColor), pillItem: color(t.PillItemColor), pillDot: color(t.PillDotColor), pillMore: color(t.PillMoreColor), pillNone: color(t.PillNoneColor), helper: color(t.HelperTextColor), done: color(t.DoneColor), error: color(t.ErrorColor)}
}
