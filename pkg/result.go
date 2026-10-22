package stepflow

import "strings"

// Result stores completed answers. Existing string result access is preserved.
type Result map[string]string

func (r Result) Get(key string) string { return r[key] }
func (r Result) Bool(key string) bool  { return strings.EqualFold(r[key], "yes") }

// Strings returns multi-select choices; this helper preserves legacy comma-delimited answers.
// Labels containing commas cannot be round-tripped; prefer using unique simple labels.
func (r Result) Strings(key string) []string {
	if strings.TrimSpace(r[key]) == "" {
		return nil
	}
	bits := strings.Split(r[key], ",")
	for i := range bits {
		bits[i] = strings.TrimSpace(bits[i])
	}
	return bits
}
