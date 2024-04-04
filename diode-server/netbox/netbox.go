package netbox

import (
	"slices"
)

// NestedTypes returns the nested types for a given object type
func NestedTypes(objectType string) []string {
	typesMap := map[string][]string{
		"dcim.device":     {"dcim.site", "dcim.devicetype", "dcim.devicerole", "dcim.platform"},
		"dcim.devicetype": {"dcim.manufacturer"},
	}

	types, ok := typesMap[objectType]
	if !ok {
		return nil
	}

	var results []string

	for _, t := range types {
		for _, n := range NestedTypes(t) {
			if !slices.Contains(results, n) {
				results = append(results, n)
			}
		}
		if !slices.Contains(results, t) {
			results = append(results, t)
		}
	}

	return results
}
