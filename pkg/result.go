package stepflow

import (
	"encoding/csv"
	"strings"
)

// Result stores completed answers. Existing string result access is preserved.
type Result map[string]string

func (r Result) Get(key string) string { return r[key] }
func (r Result) Bool(key string) bool  { return strings.EqualFold(r[key], "yes") }

// Strings returns multi-select choices. Accurate structured parsing is supported
// even when selected labels contain commas or quotes.
func (r Result) Strings(key string) []string {
	v := strings.TrimSpace(r[key])
	if v == "" {
		return nil
	}
	reader := csv.NewReader(strings.NewReader(v))
	reader.TrimLeadingSpace = true
	record, err := reader.Read()
	if err != nil {
		bits := strings.Split(v, ",")
		var out []string
		for _, b := range bits {
			if s := strings.TrimSpace(b); s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	var out []string
	for _, s := range record {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
