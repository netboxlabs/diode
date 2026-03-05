package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/graph"
)

func TestUpdateRefsInData_MatchingNameCopiesUpdate(t *testing.T) {
	data := map[string]any{
		"device": map[string]any{
			"manufacturer": map[string]any{
				"name": "ciscoSystems",
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "mfr-1",
			NodeType: "Manufacturer",
			EntityMap: map[string]any{
				"name": "ciscoSystems",
				"slug": "ciscosystems",
			},
		},
	}

	graph.ExportUpdateRefsInData(data, updates)

	device := data["device"].(map[string]any)
	mfr := device["manufacturer"].(map[string]any)
	assert.Equal(t, "ciscoSystems", mfr["name"])
	assert.Equal(t, "ciscosystems", mfr["slug"])
}

func TestUpdateRefsInData_NoMatchLeavesDataUnchanged(t *testing.T) {
	data := map[string]any{
		"device": map[string]any{
			"site": map[string]any{
				"name": "DC1",
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-2",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name": "DC2",
				"slug": "dc2",
			},
		},
	}

	graph.ExportUpdateRefsInData(data, updates)

	device := data["device"].(map[string]any)
	site := device["site"].(map[string]any)
	assert.Equal(t, "DC1", site["name"])
	_, hasSlug := site["slug"]
	assert.False(t, hasSlug)
}

func TestUpdateRefsInData_RecursesIntoUnmatchedNestedMaps(t *testing.T) {
	data := map[string]any{
		"device": map[string]any{
			"site": map[string]any{
				"name": "DC2",
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-1",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name":   "DC2",
				"slug":   "dc2",
				"status": "active",
			},
		},
	}

	graph.ExportUpdateRefsInData(data, updates)

	// The "device" map doesn't have a "name" field matching any update,
	// so the function should recurse into it and find the "site" sub-map.
	device := data["device"].(map[string]any)
	site := device["site"].(map[string]any)
	assert.Equal(t, "DC2", site["name"])
	assert.Equal(t, "dc2", site["slug"])
	assert.Equal(t, "active", site["status"])
}

func TestUpdateRefsInData_NoInfiniteRecursionWithSelfReferencing(t *testing.T) {
	// This reproduces the exact crash scenario: a Platform update contains a
	// nested "manufacturer" with the same name as another update. Without the
	// fix, maps.Copy adds the nested structure and recursion re-matches it
	// infinitely.
	data := map[string]any{
		"interface": map[string]any{
			"device": map[string]any{
				"platform": map[string]any{
					"name": "ciscoSystems",
				},
				"device_type": map[string]any{
					"manufacturer": map[string]any{
						"name": "ciscoSystems",
					},
				},
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "mfr-1",
			NodeType: "Manufacturer",
			EntityMap: map[string]any{
				"name": "ciscoSystems",
				"slug": "ciscosystems",
			},
		},
		{
			Key:      "platform-1",
			NodeType: "Platform",
			EntityMap: map[string]any{
				"name": "ciscoSystems",
				"slug": "ciscosystems",
				"manufacturer": map[string]any{
					"name": "ciscoSystems",
				},
			},
		},
	}

	// This would stack overflow before the fix.
	graph.ExportUpdateRefsInData(data, updates)

	device := data["interface"].(map[string]any)["device"].(map[string]any)
	platform := device["platform"].(map[string]any)
	assert.Equal(t, "ciscoSystems", platform["name"])
	assert.Equal(t, "ciscosystems", platform["slug"])

	dt := device["device_type"].(map[string]any)
	mfr := dt["manufacturer"].(map[string]any)
	assert.Equal(t, "ciscoSystems", mfr["name"])
	assert.Equal(t, "ciscosystems", mfr["slug"])
}

func TestUpdateRefsInData_HandlesArrayItems(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{
				"site": map[string]any{
					"name": "DC1",
				},
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-1",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name":   "DC1",
				"slug":   "dc1",
				"status": "active",
			},
		},
	}

	graph.ExportUpdateRefsInData(data, updates)

	items := data["items"].([]any)
	item := items[0].(map[string]any)
	site := item["site"].(map[string]any)
	assert.Equal(t, "DC1", site["name"])
	assert.Equal(t, "dc1", site["slug"])
}

func TestCheckForUpdatedRefs_FindsMatch(t *testing.T) {
	data := map[string]any{
		"device": map[string]any{
			"site": map[string]any{
				"name": "DC2",
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-1",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name": "DC2",
			},
		},
	}

	result := graph.ExportCheckForUpdatedRefs(data, "device-1", updates)
	assert.True(t, result)
}

func TestCheckForUpdatedRefs_ExcludesSelf(t *testing.T) {
	data := map[string]any{
		"name": "DC2",
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-1",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name": "DC2",
			},
		},
	}

	// When excludeNodeKey matches the update key, it should not report a match.
	result := graph.ExportCheckForUpdatedRefs(data, "site-1", updates)
	assert.False(t, result)
}

func TestCheckForUpdatedRefs_NoMatch(t *testing.T) {
	data := map[string]any{
		"device": map[string]any{
			"site": map[string]any{
				"name": "DC1",
			},
		},
	}

	updates := []graph.ParsedUpdateForTest{
		{
			Key:      "site-2",
			NodeType: "Site",
			EntityMap: map[string]any{
				"name": "DC2",
			},
		},
	}

	result := graph.ExportCheckForUpdatedRefs(data, "device-1", updates)
	assert.False(t, result)
}