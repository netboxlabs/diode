package changeset_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

func TestPrepare(t *testing.T) {
	type mockRetrieveObjectState struct {
		objectType     string
		objectID       int
		query          string
		objectChangeID int
		object         netbox.ComparableData
	}
	tests := []struct {
		name                 string
		rawIngestEntity      []byte
		retrieveObjectStates []mockRetrieveObjectState
		wantChangeSet        changeset.ChangeSet
		wantErr              bool
	}{
		{
			name: "[P1] ingest dcim.site - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site": {
						"name": "Site A"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "Site A",
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    "create",
						ObjectType:    "dcim.site",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: netbox.DcimSiteStatusActive,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"data": {
					"Site": {
						"name": "Site A"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "Site A",
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: netbox.DcimSiteStatusActive,
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest dcim.devicerole - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"data": {
					"DeviceRole": {
						"name": "WAN Router"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    "create",
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: "000000",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest dcim.devicerole - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"data": {
					"DeviceRole": {
						"name": "WAN Router"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: "000000",
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ingestEntity changeset.IngestEntity
			err := json.Unmarshal(tt.rawIngestEntity, &ingestEntity)
			require.NoError(t, err)

			mockClient := mocks.NewNetBoxAPI(t)

			for _, m := range tt.retrieveObjectStates {
				mockClient.EXPECT().RetrieveObjectState(context.Background(), m.objectType, m.objectID, m.query).Return(&netboxdiodeplugin.ObjectState{
					ObjectID:       m.objectID,
					ObjectType:     m.objectType,
					ObjectChangeID: m.objectChangeID,
					Object:         m.object,
				}, nil)
			}

			cs, err := changeset.Prepare(ingestEntity, mockClient)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.wantChangeSet.ChangeSet), len(cs.ChangeSet))
			for i := range tt.wantChangeSet.ChangeSet {
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].ChangeType, cs.ChangeSet[i].ChangeType)
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].ObjectType, cs.ChangeSet[i].ObjectType)
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].Data, cs.ChangeSet[i].Data)
			}
		})
	}
}
