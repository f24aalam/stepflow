package stepflow

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListItem struct{ Label, Meta string }

func Item(label string, meta ...string) ListItem {
	v := ""
	if len(meta) > 0 {
		v = meta[0]
	}
	return ListItem{label, v}
}

type listEntry struct {
	item     ListItem
	selected bool
}
type ListStep struct {
	key, question               string
	items                       []listEntry
	multi                       bool
	visibleRows, cursor, offset int
	search                      string
	next                        func(Result) []Step
	pinnedTitle, pinnedNote     string
	pinnedItems                 []string
	preselected                 []string
	width                       int
}

func List(key, question string) *ListStep {
	return &ListStep{key: key, question: question, visibleRows: 5}
}
func (s *ListStep) Items(items ...ListItem) *ListStep {
	s.items = make([]listEntry, len(items))
	for i, it := range items {
		s.items[i] = listEntry{item: it}
	}
	s.applyPreselected()
	return s
}
func (s *ListStep) MultiSelect(v bool) *ListStep { s.multi = v; return s }
func (s *ListStep) VisibleRows(n int) *ListStep  { s.visibleRows = n; return s }
func (s *ListStep) Pinned(title, note string, items ...string) *ListStep {
	s.pinnedTitle = title
	s.pinnedNote = note
	s.pinnedItems = items
	return s
}
func (s *ListStep) applyPreselected() {
	if len(s.preselected) == 0 {
		return
	}
	set := map[string]bool{}
	for _, v := range s.preselected {
		set[v] = true
	}
	for i := range s.items {
		s.items[i].selected = set[s.items[i].item.Label]
	}
}
func (s *ListStep) PreSelect(labels ...string) *ListStep {
	s.preselected = append([]string(nil), labels...)
	s.applyPreselected()
	return s
}
func (s *ListStep) WithNext(f func(Result) []Step) *ListStep { s.next = f; return s }
func (s *ListStep) Key() string                              { return s.key }
func (s *ListStep) Question() string                         { return s.question }
func (s *ListStep) reset() {
	s.search = ""
	s.offset = 0
	s.cursor = 0
	for i := range s.items {
		s.items[i].selected = false
	}
	s.applyPreselected()
	s.ensureCursorValid()
}
func (s *ListStep) Init(styles) tea.Cmd {
	s.reset()
	return nil
}
func (s *ListStep) UpdateMsg(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		return s.Update(m)
	case tea.WindowSizeMsg:
		s.width = m.Width
		return false, nil
	}
	return false, nil
}
func (s *ListStep) HasError() bool { return false }
func (s *ListStep) NextSteps(r Result) []Step {
	if s.next != nil {
		return s.next(r)
	}
	return nil
}
func (s *ListStep) filtered() []int {
	out := []int{}
	for i, e := range s.items {
		if strings.Contains(strings.ToLower(e.item.Label), strings.ToLower(s.search)) {
			out = append(out, i)
		}
	}
	return out
}
func (s *ListStep) cursorPos(ids []int) int {
	for i, n := range ids {
		if n == s.cursor {
			return i
		}
	}
	return 0
}
func (s *ListStep) ensureCursorValid() {
	ids := s.filtered()
	if len(ids) == 0 {
		s.cursor = -1
		s.offset = 0
		return
	}
	found := false
	for _, n := range ids {
		if n == s.cursor {
			found = true
			break
		}
	}
	if !found {
		s.cursor = ids[0]
		s.offset = 0
	}
	pos := s.cursorPos(ids)
	if s.offset > pos {
		s.offset = pos
	}
	if s.visibleRows > 0 && pos >= s.offset+s.visibleRows {
		s.offset = pos - s.visibleRows + 1
	}
}
func (s *ListStep) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	ids := s.filtered()
	switch msg.String() {
	case "up":
		if len(ids) > 0 {
			pos := s.cursorPos(ids)
			if pos > 0 {
				s.cursor = ids[pos-1]
			}
			s.ensureCursorValid()
		}
	case "down":
		if len(ids) > 0 {
			pos := s.cursorPos(ids)
			if pos < len(ids)-1 {
				s.cursor = ids[pos+1]
			}
			s.ensureCursorValid()
		}
	case " ", "space":
		if s.multi && s.cursor >= 0 && s.cursor < len(s.items) && len(ids) > 0 {
			s.items[s.cursor].selected = !s.items[s.cursor].selected
		}
	case "enter":
		if len(ids) == 0 {
			return false, nil
		}
		if !s.multi {
			for i := range s.items {
				s.items[i].selected = i == s.cursor
			}
		}
		return true, nil
	case "backspace":
		if s.search != "" {
			_, width := utf8.DecodeLastRuneInString(s.search)
			s.search = s.search[:len(s.search)-width]
			s.ensureCursorValid()
		}
	case "esc":
		if s.search != "" {
			s.search = ""
			s.ensureCursorValid()
		}
	default:
		if msg.Type == tea.KeyRunes {
			s.search += string(msg.Runes)
			s.ensureCursorValid()
		}
	}
	return false, nil
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
	labels := s.selectedLabels()
	if len(labels) == 0 {
		return ""
	}
	if !s.multi {
		return labels[0]
	}
	needsCSV := false
	for _, l := range labels {
		if l == "" || strings.TrimSpace(l) != l || strings.ContainsAny(l, ",\"\r\n") {
			needsCSV = true
			break
		}
	}
	if !needsCSV {
		return strings.Join(labels, ", ")
	}
	var b strings.Builder
	for i, l := range labels {
		if i > 0 {
			b.WriteByte(',')
		}
		if l == "" || strings.TrimSpace(l) != l || strings.ContainsAny(l, ",\"\r\n") {
			b.WriteByte('"')
			b.WriteString(strings.ReplaceAll(l, `"`, `""`))
			b.WriteByte('"')
		} else {
			b.WriteString(l)
		}
	}
	return b.String()
}
func (s *ListStep) View(st styles) string {
	var b strings.Builder
	writeLine := func(line string) {
		if s.width > 6 {
			line = lipgloss.NewStyle().MaxWidth(s.width - 2).Render(line)
		}
		b.WriteString(line + "\n")
	}

	if len(s.pinnedItems) > 0 {
		writeLine("── " + s.pinnedTitle + " (" + s.pinnedNote + ") ──")
		for _, v := range s.pinnedItems {
			writeLine("  •  " + v)
		}
	}
	writeLine("  " + st.searchSlash.Render("/") + " " + st.searchText.Render(s.search) + st.searchCur.Render("█"))
	ids := s.filtered()
	if len(ids) == 0 {
		writeLine("  " + st.helper.Render("No matching items"))
	}
	rows := s.visibleRows
	if rows < 1 {
		rows = 5
	}
	start := s.offset
	if start < 0 {
		start = 0
	}
	if start > len(ids) {
		start = len(ids)
	}
	end := start + rows
	if end > len(ids) {
		end = len(ids)
	}
	for _, idx := range ids[start:end] {
		e := s.items[idx]
		cursor := " "
		if idx == s.cursor {
			cursor = st.focusBar.Render("▌")
		}
		dot := st.unselDot.Render("○")
		if e.selected {
			dot = st.selDot.Render("●")
		}
		label := st.stepAnswer.Render(e.item.Label)
		if idx == s.cursor {
			label = st.focusLabel.Render(e.item.Label)
		}
		line := fmt.Sprintf(" %s  %s  %s", cursor, dot, label)
		if e.item.Meta != "" {
			line += "  " + st.itemPath.Render(e.item.Meta)
		}
		writeLine(line)
	}
	if len(ids) > rows {
		writeLine(fmt.Sprintf("  %d / %d", s.cursorPos(ids)+1, len(ids)))
	}
	writeLine("  " + st.helper.Render("Selected: ") + st.pillItem.Render(strings.Join(s.selectedLabels(), " · ")))
	if s.multi {
		writeLine("  " + st.helper.Render("↑↓ navigate · space toggle · enter confirm · esc clear search"))
	} else {
		writeLine("  " + st.helper.Render("↑↓ navigate · enter select · esc clear search"))
	}
	return b.String()
}
func (s *ListStep) ResultView(st styles) string { return st.stepAnswer.Render(s.Answer()) }
