package fate

import "testing"

func TestQuoteFix_QuoteFix(t *testing.T) {
	tests := []struct {
		s        string
		expected string
	}{
		{"this is \"a test", "this is \"a test\""},
		{"\"this\" is \"a test", "\"this\" is \"a test\""},
		{"this) is \"a test", "(this) is \"a test\""},
		{"this)) is ((a test)", "((this)) is ((a test))"},
		{"this]) is ((a test)", "([this]) is ((a test))"},
		{"this” is “a test”", "“this” is “a test”"},
		{"(this is a test\"", "\"(this is a test)\""},
		{"this is a test :)", "this is a test :)"},
		{":) :( :-) :-( ;)", ":) :( :-) :-( ;)"},
	}

	for _, tt := range tests {
		result := QuoteFix(tt.s)
		if result != tt.expected {
			t.Fatalf("Quote(%s) -> %s, want %s", tt.s, result, tt.expected)
		}
	}
}
