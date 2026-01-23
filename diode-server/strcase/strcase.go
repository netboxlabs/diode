// Package strcase provides string case conversion utilities.
package strcase

import "strings"

// ToSnakeCase converts PascalCase or camelCase to snake_case.
// e.g., "DeviceType" -> "device_type", "Site" -> "site"
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// ToUpperSnakeCase converts PascalCase or camelCase to UPPER_SNAKE_CASE.
// e.g., "DeviceType" -> "DEVICE_TYPE", "site" -> "SITE"
func ToUpperSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		if r >= 'a' && r <= 'z' {
			r = r - 'a' + 'A'
		}
		result.WriteRune(r)
	}

	return result.String()
}
