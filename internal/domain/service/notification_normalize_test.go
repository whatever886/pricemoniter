package service

import "testing"

func TestNormalizeBarkKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain key",
			input: "ABC123",
			want:  "ABC123",
		},
		{
			name:  "full url without trailing slash",
			input: "https://api.day.app/ABC123",
			want:  "ABC123",
		},
		{
			name:  "full url with trailing slash",
			input: "https://api.day.app/ABC123/",
			want:  "ABC123",
		},
		{
			name:  "full url with query and fragment",
			input: "https://api.day.app/ABC123/?isArchive=1#foo",
			want:  "ABC123",
		},
		{
			name:  "empty input",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeBarkKey(tt.input); got != tt.want {
				t.Fatalf("normalizeBarkKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

