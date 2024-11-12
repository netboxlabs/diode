package changeset

const (
	// ChangeTypeCreate is the change type for a creation
	ChangeTypeCreate = "create"

	// ChangeTypeUpdate is the change type for an update
	ChangeTypeUpdate = "update"
)

// ChangeSet represents a change set
type ChangeSet struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
}

// Change represents a change for the change set
type Change struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"`
	ObjectType    string `json:"object_type"`
	ObjectID      *int   `json:"object_id,omitempty"`
	ObjectVersion *int   `json:"object_version,omitempty"`
	Data          any    `json:"data"`
}
