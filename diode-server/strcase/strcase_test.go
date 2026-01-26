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
		{"acronym_start", "IPAddress", "ip_address"},
		{"acronym_end", "DeviceID", "device_id"},
		{"acronym_middle", "GetHTTPResponse", "get_http_response"},
		{"already_snake", "already_snake", "already_snake"},
		{"all_caps", "ABC", "abc"},
		{"camelCase", "deviceType", "device_type"},
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
		{"acronym_start", "IPAddress", "IP_ADDRESS"},
		{"acronym_end", "DeviceID", "DEVICE_ID"},
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

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single_word", "site", "Site"},
		{"two_words", "device_type", "DeviceType"},
		{"multiple_words", "tagged_vlans", "TaggedVlans"},
		{"already_pascal", "Site", "Site"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
