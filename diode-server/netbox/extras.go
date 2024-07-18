package netbox

const (
	// ExtrasTagObjectType represents the tag object type
	ExtrasTagObjectType = "extras.tag"
)

// Tag represents a tag
type Tag struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Color string `json:"color,omitempty"`
}
