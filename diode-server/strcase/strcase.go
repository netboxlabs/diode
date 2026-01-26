// Package strcase provides string case conversion utilities.
package strcase

import "strings"

// ToSnakeCase converts PascalCase or camelCase to snake_case.
// Handles acronyms correctly: "IPAddress" -> "ip_address", "DeviceID" -> "device_id"
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	var result strings.Builder

	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			// Add underscore if:
			// 1. Previous char is lowercase (normal word boundary): "deviceType" -> "device_Type"
			// 2. Previous char is uppercase AND next char is lowercase (end of acronym): "IPAddress" -> "IP_Address"
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune('_')
			} else if prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
				result.WriteRune('_')
			}
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// ToUpperSnakeCase converts PascalCase or camelCase to UPPER_SNAKE_CASE.
// Handles acronyms correctly: "IPAddress" -> "IP_ADDRESS", "DeviceID" -> "DEVICE_ID"
func ToUpperSnakeCase(s string) string {
	return strings.ToUpper(ToSnakeCase(s))
}

// ToPascalCase converts snake_case to PascalCase.
// e.g., "tagged_vlans" -> "TaggedVlans", "site" -> "Site"
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}
