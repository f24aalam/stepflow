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
// using CSV decoding for labels containing commas, quotes, whitespace, or newlines.
// Note: unquoted legacy comma-separated values (e.g. "a, b") are split by comma with
// leading whitespace trimmed for backward compatibility. Because unquoted values in
// map[string]string are inherently ambiguous, selections requiring preserved leading
// whitespace or commas should use CSV quoting.
// Go's encoding/csv normalizes CRLF inside quoted values to LF, so exact byte-for-byte
// round-tripping of arbitrary newline sequences is not guaranteed.
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
			out = append(out, strings.TrimSpace(b))
		}
		return out
	}
	return record
}
