package stepflow

import "github.com/charmbracelet/lipgloss"

// Theme configures the semantic colors used throughout the wizard.
type Theme struct {
	Primary lipgloss.Color // Brand color: markers, prompt symbols, completed questions, search slash
	Accent  lipgloss.Color // Focus/Selection: active cursor bar, focused text, selected dots, pills
	Muted   lipgloss.Color // Subdued/Secondary: pipes, unselected circles, meta descriptions, helper text
	Text    lipgloss.Color // Readable content: unselected items, answer text, body lines
	Success lipgloss.Color // Completion and success indicators
	Error   lipgloss.Color // Failure and error indicators
}

// NewTheme creates a complete, cohesive theme from primary, accent, and muted colors.
// Text, Success, and Error are initialized with sensible defaults.
func NewTheme(primary, accent, muted lipgloss.Color) Theme {
	return Theme{
		Primary: primary,
		Accent:  accent,
		Muted:   muted,
		Text:    lipgloss.Color("#f8fafc"),
		Success: lipgloss.Color("#10b981"),
		Error:   lipgloss.Color("#ef4444"),
	}
}

// DefaultTheme returns Stepflow's signature cohesive green palette.
func DefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.Color("#22c55e"), // Bright green
		Accent:  lipgloss.Color("#4ade80"), // Light emerald focus
		Muted:   lipgloss.Color("#15803d"), // Subdued forest green
		Text:    lipgloss.Color("#ecfdf5"), // Light mint text
		Success: lipgloss.Color("#4ade80"), // Green checkmark
		Error:   lipgloss.Color("#ef4444"), // Red error
	}
}
