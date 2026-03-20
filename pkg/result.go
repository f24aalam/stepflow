package stepflow

// Result holds all answers keyed by their step key.
// Values are always strings; list answers are comma-separated.
type Result map[string]string

// Get returns the answer for the given key, or "" if not present.
func (r Result) Get(key string) string {
	return r[key]
}

// Bool returns true if the value for key is "Yes" (case-insensitive).
func (r Result) Bool(key string) bool {
	v := r[key]
	return v == "Yes" || v == "yes"
}
