package strcase

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single_word", "Site", "site"},
		{"two_words", "DeviceType", "device_type"},
		{"consecutive_caps", "IPAddress", "i_p_address"},
		{"already_snake", "already_snake", "already_snake"},
		{"all_caps", "ABC", "a_b_c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToUpperSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single_word_caps", "Site", "SITE"},
		{"two_words", "DeviceType", "DEVICE_TYPE"},
		{"lowercase", "site", "SITE"},
		{"already_snake", "already_snake", "ALREADY_SNAKE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToUpperSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("ToUpperSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
