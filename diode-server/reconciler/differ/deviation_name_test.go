package differ

import (
	"testing"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

func TestGenDeviationName(t *testing.T) {
	refA := "ref-a"
	refB := "ref-b"

	tests := []struct {
		name       string
		changes    []changeset.Change
		objectType string
		want       string
	}{
		{
			name: "create of primary object yields discovered",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A discovered",
		},
		{
			name: "update of primary object yields modified",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.site", ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A modified",
		},
		{
			name: "noop-only primary still classified as modified",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeNoop, ObjectType: "dcim.site", ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A modified",
		},
		{
			name: "create-then-update on same refID returns the create",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A discovered",
		},
		{
			name: "no change of target object type yields unchanged",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.device", ObjectPrimaryValue: "Device A"},
			},
			objectType: "dcim.site",
			want:       "Site unchanged",
		},
		{
			name:       "empty change set yields unchanged",
			changes:    nil,
			objectType: "dcim.site",
			want:       "Site unchanged",
		},
		{
			name: "unknown object type falls back to raw type string",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "made.up", RefID: &refA, ObjectPrimaryValue: "Thing"},
			},
			objectType: "made.up",
			want:       "made.up Thing discovered",
		},
		{
			name: "unrecognized change type is flagged in name",
			changes: []changeset.Change{
				{ChangeType: "delete", ObjectType: "dcim.site", ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A (unrecognized change type delete)",
		},
		{
			name: "primary value missing yields name without value",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA},
			},
			objectType: "dcim.site",
			want:       "Site discovered",
		},
		{
			name: "mixed types: target-type change is selected over others",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.device", RefID: &refB, ObjectPrimaryValue: "Device A"},
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.site", ObjectPrimaryValue: "Site A"},
			},
			objectType: "dcim.site",
			want:       "Site Site A modified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := genDeviationName(tt.changes, tt.objectType)
			if got == nil {
				t.Fatalf("genDeviationName() = nil, want %q", tt.want)
			}
			if *got != tt.want {
				t.Errorf("genDeviationName() = %q, want %q", *got, tt.want)
			}
		})
	}
}

func TestFindPrimaryChange(t *testing.T) {
	refA := "ref-a"
	refB := "ref-b"

	tests := []struct {
		name        string
		changes     []changeset.Change
		objectType  string
		wantNil     bool
		wantRefID   *string
		wantType    string
		wantPrimary string
	}{
		{
			name: "single create returns it",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
			},
			objectType:  "dcim.site",
			wantRefID:   &refA,
			wantType:    "dcim.site",
			wantPrimary: "Site A",
		},
		{
			name: "update on object_id (nil refID) returns the update",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.site", ObjectPrimaryValue: "Site A"},
			},
			objectType:  "dcim.site",
			wantRefID:   nil,
			wantType:    "dcim.site",
			wantPrimary: "Site A",
		},
		{
			name: "create-then-update on same refID returns the create",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
			},
			objectType:  "dcim.site",
			wantRefID:   &refA,
			wantType:    "dcim.site",
			wantPrimary: "Site A",
		},
		{
			name: "different refIDs returns last of target type",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refB, ObjectPrimaryValue: "Site B"},
			},
			objectType:  "dcim.site",
			wantRefID:   &refB,
			wantType:    "dcim.site",
			wantPrimary: "Site B",
		},
		{
			name: "no change of target type returns nil",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.device", ObjectPrimaryValue: "Device A"},
			},
			objectType: "dcim.site",
			wantNil:    true,
		},
		{
			name:       "empty changes returns nil",
			changes:    nil,
			objectType: "dcim.site",
			wantNil:    true,
		},
		{
			name: "mixed types: returns the target-type change",
			changes: []changeset.Change{
				{ChangeType: changeset.ChangeTypeUpdate, ObjectType: "dcim.device", ObjectPrimaryValue: "Device A"},
				{ChangeType: changeset.ChangeTypeCreate, ObjectType: "dcim.site", RefID: &refA, ObjectPrimaryValue: "Site A"},
			},
			objectType:  "dcim.site",
			wantRefID:   &refA,
			wantType:    "dcim.site",
			wantPrimary: "Site A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindPrimaryChange(tt.changes, tt.objectType)
			if tt.wantNil {
				if got != nil {
					t.Errorf("FindPrimaryChange() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("FindPrimaryChange() = nil, want change for %q", tt.objectType)
			}
			if got.ObjectType != tt.wantType {
				t.Errorf("ObjectType = %q, want %q", got.ObjectType, tt.wantType)
			}
			if got.ObjectPrimaryValue != tt.wantPrimary {
				t.Errorf("ObjectPrimaryValue = %q, want %q", got.ObjectPrimaryValue, tt.wantPrimary)
			}
			if !refEqual(got.RefID, tt.wantRefID) {
				t.Errorf("RefID = %v, want %v", got.RefID, tt.wantRefID)
			}
		})
	}
}

func TestDeviationNameForDiffFailure(t *testing.T) {
	tests := []struct {
		name   string
		entity IngestEntity
		want   string
	}{
		{
			name: "known object type with primary value",
			entity: IngestEntity{
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{Site: &diodepb.Site{Name: "Site A"}},
				},
			},
			want: "Site Site A reported",
		},
		{
			name: "unknown object type",
			entity: IngestEntity{
				ObjectType: "made.up",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{Site: &diodepb.Site{Name: "Site A"}},
				},
			},
			want: "Unresolved made.up reported",
		},
		{
			name: "nil entity payload yields unresolved",
			entity: IngestEntity{
				ObjectType: "dcim.site",
				Entity:     nil,
			},
			want: "Unresolved dcim.site reported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deviationNameForDiffFailure(tt.entity)
			if got != tt.want {
				t.Errorf("deviationNameForDiffFailure() = %q, want %q", got, tt.want)
			}
		})
	}
}

func refEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
