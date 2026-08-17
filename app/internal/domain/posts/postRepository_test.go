package posts

import "testing"

func TestBuildTSQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"single term gets a prefix match", "фитнес", "фитнес:*"},
		{"terms are ANDed", "фитнес зала", "фитнес & зала:*"},
		{"case is normalised", "Фитнес", "фитнес:*"},
		{"latin works too", "protein shake", "protein & shake:*"},
		{"digits are kept", "top 10", "top & 10:*"},
		{"empty input", "", ""},
		{"only punctuation", "!!! ???", ""},
		{"surrounding whitespace", "  фитнес  ", "фитнес:*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTSQuery(tt.query); got != tt.want {
				t.Errorf("buildTSQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

// tsquery operators in user input must not reach the parser, which would turn a
// stray "&" or "!" into a query syntax error at request time.
func TestBuildTSQuery_StripsQueryOperators(t *testing.T) {
	tests := []string{
		"foo & bar",
		"foo | bar",
		"foo:*",
		"!foo",
		"(foo)",
		"foo <-> bar",
		"'foo'",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			got := buildTSQuery(query)

			// The only operators left must be the ones we add ourselves.
			for _, forbidden := range []string{"|", "!", "(", ")", "<", ">", "'"} {
				if contains(got, forbidden) {
					t.Errorf("buildTSQuery(%q) = %q, leaked operator %q", query, got, forbidden)
				}
			}
		})
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}

	return false
}
