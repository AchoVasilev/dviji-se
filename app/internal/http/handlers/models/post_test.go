package models

import "testing"

func TestAuthorInitials(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"both names", "Иван", "Петров", "ИП"},
		{"latin names", "John", "Smith", "JS"},
		{"only first name", "Иван", "", "И"},
		{"only last name", "", "Петров", "П"},
		{"no name at all", "", "", "?"},
		{"multibyte is not split", "Ъ", "Я", "ЪЯ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AuthorInitials(tt.firstName, tt.lastName); got != tt.want {
				t.Errorf("AuthorInitials(%q, %q) = %q, want %q", tt.firstName, tt.lastName, got, tt.want)
			}
		})
	}
}
