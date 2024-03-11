package changeset_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

func TestPrepareChange(t *testing.T) {
	tests := []struct {
		name            string
		rawIngestEntity []byte
		rawObjectState  []byte
		wantChange      changeset.Change
		wantErr         bool
	}{
		{
			name: "Create dcim.site",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site": {
						"name": "test",	
						"slug": "test"
					}
				},
				"state": 0
			}`),
			rawObjectState: nil,
			wantChange: changeset.Change{
				ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeType:    "create",
				ObjectType:    "dcim.site",
				ObjectID:      nil,
				ObjectVersion: nil,
				Data: &netbox.DcimSite{
					Name: "test",
					Slug: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "Update dcim.site",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site": {
						"name": "test 2",	
						"slug": "test"
					}
				},
				"state": 0
			}`),
			rawObjectState: []byte(`{
			  "object": {
				"Site": {
				  "id": 1,
				  "name": "test",
				  "slug": "test",
				  "url": "http://localhost:8000/api/dcim/sites/1/"
				}
			  },
			  "object_change_id": 1,
			  "object_id": 1,
			  "object_type": "dcim.site"
			}`),
			wantChange: changeset.Change{
				ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeType:    "update",
				ObjectType:    "dcim.site",
				ObjectID:      ptrInt(1),
				ObjectVersion: ptrInt(1),
				Data: &netbox.DcimSite{
					Name: "test 2",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid ingest entity",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site1": {
						"name": "test 2",	
						"slug": "test"
					}
				},
				"state": 0
			}`),
			rawObjectState: []byte(`{
			  "object": {
				"Site": {
				  "id": 1,
				  "name": "test",
				  "slug": "test",
				  "url": "http://localhost:8000/api/dcim/sites/1/"
				}
			  },
			  "object_change_id": 1,
			  "object_id": 1,
			  "object_type": "dcim.site"
			}`),
			wantChange: changeset.Change{},
			wantErr:    true,
		},
		{
			name: "Invalid object state",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site": {
						"name": "test 2",	
						"slug": "test"
					}
				},
				"state": 0
			}`),
			rawObjectState: []byte(`{
			  "object": {
				"Site1": {
				  "id": 1,
				  "name": "test",
				  "slug": "test",
				  "url": "http://localhost:8000/api/dcim/sites/1/"
				}
			  },
			  "object_change_id": 1,
			  "object_id": 1,
			  "object_type": "dcim.site"
			}`),
			wantChange: changeset.Change{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ingestEntity changeset.IngestEntity
			err := json.Unmarshal(tt.rawIngestEntity, &ingestEntity)
			require.NoError(t, err)

			var objectState *changeset.ObjectState
			if tt.rawObjectState != nil {
				err = json.Unmarshal(tt.rawObjectState, &objectState)
				require.NoError(t, err)
			}

			change, err := changeset.PrepareChange(ingestEntity, objectState)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantChange.ChangeType, change.ChangeType)
			assert.Equal(t, tt.wantChange.ObjectType, change.ObjectType)
			assert.Equal(t, tt.wantChange.ObjectID, change.ObjectID)
			assert.Equal(t, tt.wantChange.ObjectVersion, change.ObjectVersion)
			assert.Equal(t, tt.wantChange.Data, change.Data)
		})
	}
}

func ptrInt(i int) *int {
	return &i
}
