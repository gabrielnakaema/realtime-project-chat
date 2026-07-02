package repository

import "testing"

func TestTaskCodeSequenceBase(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "strips trailing digits", prefix: "BACKEND-9", want: "BACKEND-"},
		{name: "strips multiple trailing digits", prefix: "BACKEND-009", want: "BACKEND-"},
		{name: "no trailing digits stays unchanged", prefix: "BACKEND-", want: "BACKEND-"},
		{name: "digits only falls back to prefix", prefix: "123", want: "123"},
		{name: "digits in the middle are kept", prefix: "V2-BACKEND-9", want: "V2-BACKEND-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskCodeSequenceBase(tt.prefix)
			if got != tt.want {
				t.Errorf("taskCodeSequenceBase(%q) = %q, want %q", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestTaskCodeLikePattern(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "plain value gets trailing wildcard", value: "backend-", want: "backend-%"},
		{name: "escapes percent", value: "50%", want: "50\\%%"},
		{name: "escapes underscore", value: "ab_1", want: "ab\\_1%"},
		{name: "escapes backslash before other metacharacters", value: `a\_b`, want: `a\\\_b%`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskCodeLikePattern(tt.value)
			if got != tt.want {
				t.Errorf("taskCodeLikePattern(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
