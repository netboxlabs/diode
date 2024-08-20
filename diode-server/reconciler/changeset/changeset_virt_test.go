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

func TestVirtualizationPrepare(t *testing.T) {
	type mockRetrieveObjectState struct {
		objectType     string
		objectID       int
		queryParams    map[string]string
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
			name: "[P1] ingest virtualization.clustergroup with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustergroup",
				"entity": {
					"ClusterGroup": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.clustergroup",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationClusterGroup{
							Name: "Test",
							Slug: "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest virtualization.clustergroup with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustergroup",
				"entity": {
					"ClusterGroup": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "Test",
							Slug: "test",
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
			name: "[P1] ingest empty virtualization.clustergroup - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustergroup",
				"entity": {
					"ClusterGroup": {}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P2] ingest virtualization.clustertype with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustertype",
				"entity": {
					"ClusterType": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.clustertype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationClusterType{
							Name: "Test",
							Slug: "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest virtualization.clustergroup with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustertype",
				"entity": {
					"ClusterType": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "Test",
							Slug: "test",
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
			name: "[P2] ingest empty virtualization.clustertype - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.clustertype",
				"entity": {
					"ClusterType": {}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ingestEntity changeset.IngestEntity
			err := json.Unmarshal(tt.rawIngestEntity, &ingestEntity)
			require.NoError(t, err)

			mockClient := mocks.NewNetBoxAPI(t)

			for _, m := range tt.retrieveObjectStates {
				mockClient.EXPECT().RetrieveObjectState(context.Background(), netboxdiodeplugin.RetrieveObjectStateQueryParams{
					ObjectType: m.objectType,
					ObjectID:   m.objectID,
					Params:     m.queryParams,
				}).Return(&netboxdiodeplugin.ObjectState{
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

			require.Equal(t, len(tt.wantChangeSet.ChangeSet), len(cs.ChangeSet))

			for i := range tt.wantChangeSet.ChangeSet {
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].ChangeType, cs.ChangeSet[i].ChangeType)
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].ObjectType, cs.ChangeSet[i].ObjectType)
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].Data, cs.ChangeSet[i].Data)
			}
		})
	}
}
