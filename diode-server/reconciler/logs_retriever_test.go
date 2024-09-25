package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeSpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "string with all considered special characters",
			input: `,.<>{}[]"':;!@#$%^&*()-+=~`,
			want:  `\,\.\<\>\{\}\[\]\"\'\:\;\!\@\#\$\%\^\&\*\(\)\-\+\=\~`,
		},
		{
			name:  "producer app name",
			input: "example-app",
			want:  "example\\-app",
		},
		{
			name:  "producer app version",
			input: "0.1.0",
			want:  "0\\.1\\.0",
		},
		{
			name:  "request ID",
			input: "123e4567-e89b-12d3-a456-426614174000",
			want:  "123e4567\\-e89b\\-12d3\\-a456\\-426614174000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeSpecialChars(tt.input); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
