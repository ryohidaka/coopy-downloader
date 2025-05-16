package kikaku

import "testing"

func TestCalculateKikakuCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-01-21", "2025011"},
		{"2025-02-18", "2025021"},
		{"2025-03-18", "2025031"},
		{"2025-04-15", "2025041"},
		{"2025-05-20", "2025051"},
	}

	for _, tt := range tests {
		got, err := CalculateKikakuCode(tt.input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("got %s, want %s", got, tt.expected)
		}
	}
}
