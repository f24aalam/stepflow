package stepflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ListItem represents a single selectable entry in a ListStep.
type ListItem struct {
	Label string
	Meta  string // optional subtitle / path shown dimmed
}

// Item creates a ListItem with an optional meta string.
func Item(label string, meta ...string) ListItem {
	m := ""
	if len(meta) > 0 {
		m = meta[0]
	}

	return ListItem{Label: label, Meta: m}
}

type listEntry struct {
	item     ListItem
	selected bool
}

// ListStep is a searchable, scrollable multi- or single-select list.
type ListStep struct {
	key         string
	question    string
	items       []listEntry
	multi       bool
	visibleRows int
	cursor      int
	offset      int
	search      string

	// pinned section shown above the interactive list (read-only)
	pinnedTitle string
	pinnedNote  string
	pinnedItems []string
}

// List creates a new list selection step.
func List(key, question string) *ListStep {
	return &ListStep{
		key:         key,
		question:    question,
		visibleRows: 5,
	}
}

// Items sets the selectable entries.
func (s *ListStep) Items(items ...ListItem) *ListStep {
	s.items = make([]listEntry, len(items))
	for i, it := range items {
		s.items[i] = listEntry{item: it}
	}

	return s
}

// MultiSelect enables multi-selection (space to toggle). Default: false.
func (s *ListStep) MultiSelect(m bool) *ListStep {
	s.multi = m
	return s
}

// VisibleRows controls how many rows are shown at once (default 5).
func (s *ListStep) VisibleRows(n int) *ListStep {
	s.visibleRows = n
	return s
}

// Pinned adds a read-only section above the interactive list.
func (s *ListStep) Pinned(title, note string, items ...string) *ListStep {
	s.pinnedTitle = title
	s.pinnedNote = note
	s.pinnedItems = items

	return s
}

// PreSelect marks items as selected by label.
func (s *ListStep) PreSelect(labels ...string) *ListStep {
	set := make(map[string]bool, len(labels))
	for _, l := range labels {
		set[l] = true
	}

	for i, e := range s.items {
		if set[e.item.Label] {
			s.items[i].selected = true
		}
	}

	return s
}

func (s *ListStep) Key() string           { return s.key }
func (s *ListStep) Question() string      { return s.question }
func (s *ListStep) Init(_ styles) tea.Cmd { return nil }

func (s *ListStep) filtered() []int {
	var out []int
	q := strings.ToLower(s.search)
	for i, e := range s.items {
		if q == "" || strings.Contains(strings.ToLower(e.item.Label), q) {
			out = append(out, i)
		}
	}

	return out
}

func (s *ListStep) cursorPos(idxs []int) int {
	for i, idx := range idxs {
		if idx == s.cursor {
			return i
		}
	}

	return 0
}

func (s *ListStep) ensureCursorValid() {
	idxs := s.filtered()
	if len(idxs) == 0 {
		s.cursor = 0
		s.offset = 0
		return
	}

	for _, idx := range idxs {
		if idx == s.cursor {
			return
		}
	}

	s.cursor = idxs[0]
	s.offset = 0
}

func (s *ListStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	idxs := s.filtered()

	switch msg.String() {
	case "up", "k":
		pos := s.cursorPos(idxs)
		if pos > 0 {
			s.cursor = idxs[pos-1]
			if pos-1 < s.offset {
				s.offset = pos - 1
			}
		}

	case "down", "j":
		pos := s.cursorPos(idxs)
		if pos < len(idxs)-1 {
			s.cursor = idxs[pos+1]
			if pos+1 >= s.offset+s.visibleRows {
				s.offset = pos + 2 - s.visibleRows
			}
		}

	case " ":
		if s.multi {
			s.items[s.cursor].selected = !s.items[s.cursor].selected
		}

	case "enter":
		if !s.multi {
			for i := range s.items {
				s.items[i].selected = i == s.cursor
			}
		}
		return true, nil

	case "backspace":
		if len(s.search) > 0 {
			s.search = s.search[:len(s.search)-1]
			s.ensureCursorValid()
		}

	default:
		if len(msg.String()) == 1 {
			s.search += msg.String()
			s.ensureCursorValid()
		}
	}

	return false, nil
}

func (s *ListStep) View(st styles) string {
	var b strings.Builder

	// Pinned section
	if len(s.pinnedItems) > 0 {
		b.WriteString(renderDivider(st, s.pinnedTitle, s.pinnedNote, 58) + "\n")
		for _, name := range s.pinnedItems {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				st.selDot.Render("•"),
				st.stepAnswer.Render(name),
			))
		}
		b.WriteString("\n")
	}

	if len(s.pinnedItems) > 0 || s.pinnedTitle != "" {
		b.WriteString(renderDivider(st, "Additional", "", 58) + "\n\n")
	}

	// Search bar
	searchLine := st.searchSlash.Render("/") + " "
	if s.search == "" {
		searchLine += st.helper.Render("type to filter") + st.searchCur.Render("█")
	} else {
		searchLine += st.searchText.Render(s.search) + st.searchCur.Render("█")
	}
	b.WriteString("  " + searchLine + "\n\n")

	// Scrolling list
	idxs := s.filtered()
	total := len(idxs)
	end := s.offset + s.visibleRows
	if end > total {
		end = total
	}

	for _, idx := range idxs[s.offset:end] {
		e := s.items[idx]
		focused := idx == s.cursor

		gutter := " "
		if focused {
			gutter = st.focusBar.Render("▌")
		}

		dot := st.unselDot.Render("○")
		if e.selected {
			dot = st.selDot.Render("●")
		}

		label := st.stepAnswer.Render(e.item.Label)
		if focused {
			label = st.focusLabel.Render(e.item.Label)
		}

		meta := ""
		if e.item.Meta != "" {
			meta = "  " + st.itemPath.Render(e.item.Meta)
		}

		b.WriteString(fmt.Sprintf(" %s  %s  %-20s%s\n", gutter, dot, label, meta))
	}

	// Scroll counter
	if total > s.visibleRows {
		pos := s.cursorPos(idxs) + 1
		b.WriteString("  " +
			st.scrollPos.Render(fmt.Sprintf("%d", pos)) +
			st.scrollTotal.Render(fmt.Sprintf(" / %d", total)) +
			"\n")
	} else {
		b.WriteString("\n")
	}

	// Selected pill
	b.WriteString("\n")
	selected := s.selectedLabels()
	if len(selected) == 0 {
		b.WriteString("  " +
			st.pillBracket.Render("[") + " " +
			st.pillNone.Render("none selected") + " " +
			st.pillBracket.Render("]") + "\n")
	} else {
		const maxPill = 3
		shown := selected
		extra := 0

		if len(selected) > maxPill {
			shown = selected[:maxPill]
			extra = len(selected) - maxPill
		}

		var parts []string
		for _, n := range shown {
			parts = append(parts, st.pillItem.Render(n))
		}

		pill := st.pillBracket.Render("[") + " " +
			strings.Join(parts, st.pillDot.Render(" · ")) + " " +
			st.pillBracket.Render("]")
		if extra > 0 {
			pill += " " + st.pillMore.Render(fmt.Sprintf("+%d more", extra))
		}

		b.WriteString("  " + pill + "\n")
	}

	hint := "↑↓ navigate   space toggle   enter confirm"
	if !s.multi {
		hint = "↑↓ navigate   enter select"
	}

	b.WriteString("\n  " + st.helper.Render(hint) + "\n")

	return b.String()
}

func (s *ListStep) selectedLabels() []string {
	var out []string
	for _, e := range s.items {
		if e.selected {
			out = append(out, e.item.Label)
		}
	}

	return out
}

func (s *ListStep) Answer() string {
	return strings.Join(s.selectedLabels(), ", ")
}

// ── shared divider helper ─────────────────────────────────────────────────────

func renderDivider(st styles, label, note string, width int) string {
	rawLen := 3 + len(label) + 1
	var mid string

	if note != "" {
		rawLen += 3 + len(note)
		mid = st.sectionLbl.Render(label+" (") +
			st.alwaysNote.Render(note) +
			st.sectionLbl.Render(") ")
	} else {
		mid = st.sectionLbl.Render(label + " ")
	}

	dashes := width - rawLen
	if dashes < 2 {
		dashes = 2
	}

	return st.sectionLine.Render("── ") + mid + st.sectionLine.Render(strings.Repeat("─", dashes))
}
