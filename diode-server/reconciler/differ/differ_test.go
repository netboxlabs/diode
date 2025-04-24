package differ_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
)

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestGenDeviationName(t *testing.T) {
	type mockGenerateDiffResponse struct {
		result *netboxdiodeplugin.ChangeSetResult
		err    error
	}

	tests := []struct {
		name          string
		ingestEntity  differ.IngestEntity
		response      *mockGenerateDiffResponse
		wantChangeSet changeset.ChangeSet
		wantErr       bool
	}{
		{
			name: "[P1] ingest dcim.site with name only - existing object not found - create",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
			},
			response: &mockGenerateDiffResponse{
				result: &netboxdiodeplugin.ChangeSetResult{
					ChangeSet: &netboxdiodeplugin.ChangeSet{
						Changes: []netboxdiodeplugin.Change{
							{
								ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "create",
								ObjectType:         "dcim.site",
								ObjectID:           nil,
								RefID:              strPtr("ref-1"),
								Data:               json.RawMessage(`{"name": "Site A"}`),
								ObjectPrimaryValue: "Site A",
							},
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				Changes: []changeset.Change{
					{
						ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:         changeset.ChangeTypeCreate,
						ObjectType:         "dcim.site",
						ObjectID:           nil,
						ObjectVersion:      nil,
						RefID:              strPtr("ref-1"),
						ObjectPrimaryValue: "Site A",
						After:              json.RawMessage(`{"name": "Site A"}`),
					},
				},
				DeviationName: strPtr("Site Site A created"),
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site with name only - existing object found - do nothing",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
			},
			response: &mockGenerateDiffResponse{
				result: &netboxdiodeplugin.ChangeSetResult{
					ChangeSet: &netboxdiodeplugin.ChangeSet{
						ID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						Changes: []netboxdiodeplugin.Change{},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ID:            "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				Changes:       []changeset.Change{},
				DeviationName: strPtr("Site unchanged"),
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site with tags - existing object found - update with new tags",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
							Tags: []*diodepb.Tag{
								{
									Name: "tag 1",
								},
								{
									Name: "tag 2",
								},
							},
						},
					},
				},
			},
			response: &mockGenerateDiffResponse{
				result: &netboxdiodeplugin.ChangeSetResult{
					ChangeSet: &netboxdiodeplugin.ChangeSet{
						ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						Changes: []netboxdiodeplugin.Change{
							{
								ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
								ChangeType:         "create",
								ObjectType:         "extras.tag",
								ObjectID:           nil,
								ObjectVersion:      nil,
								Data:               json.RawMessage(`{"name": "tag 2"}`),
								ObjectPrimaryValue: "tag 2",
							},
							{
								ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "update",
								ObjectType:         "dcim.site",
								ObjectID:           intPtr(1),
								ObjectVersion:      nil,
								Before:             json.RawMessage(`{"name": "Site A", "status": "active", "tags": [1, 2]}`),
								Data:               json.RawMessage(`{"name": "Site A", "status": "active", "tags": [1, 2, 3]}`),
								ObjectPrimaryValue: "Site A",
							},
						},
					},
				},
			},

			wantChangeSet: changeset.ChangeSet{
				ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				Changes: []changeset.Change{
					{
						ID:            "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						After:         json.RawMessage(`{"name": "tag 2"}`),
					},
					{
						ID:            "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.site",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Before:        json.RawMessage(`{"name": "Site A", "status": "active", "tags": [1, 2]}`),
						After:         json.RawMessage(`{"name": "Site A", "status": "active", "tags": [1, 2, 3]}`),
					},
				},
				DeviationName: strPtr("Site Site A modified"),
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site create then modify - is a create",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
			},
			response: &mockGenerateDiffResponse{
				result: &netboxdiodeplugin.ChangeSetResult{
					ChangeSet: &netboxdiodeplugin.ChangeSet{
						Changes: []netboxdiodeplugin.Change{
							{
								ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "create",
								ObjectType:         "dcim.site",
								ObjectID:           nil,
								RefID:              strPtr("ref-1"),
								Data:               json.RawMessage(`{"name": "Site A"}`),
								ObjectPrimaryValue: "Site A",
							},
							{
								ID:                 "5663a77f-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "update",
								ObjectType:         "dcim.site",
								ObjectID:           nil,
								RefID:              strPtr("ref-1"),
								Data:               json.RawMessage(`{"name": "Site A", "status": "active"}`),
								Before:             json.RawMessage(`{"name": "Site A"}`),
								ObjectPrimaryValue: "Site A",
							},
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				Changes: []changeset.Change{
					{
						ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:         changeset.ChangeTypeCreate,
						ObjectType:         "dcim.site",
						ObjectID:           nil,
						ObjectVersion:      nil,
						RefID:              strPtr("ref-1"),
						ObjectPrimaryValue: "Site A",
						After:              json.RawMessage(`{"name": "Site A"}`),
					},
					{
						ID:                 "5663a77f-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:         "update",
						ObjectType:         "dcim.site",
						ObjectID:           nil,
						RefID:              strPtr("ref-1"),
						After:              json.RawMessage(`{"name": "Site A", "status": "active"}`),
						Before:             json.RawMessage(`{"name": "Site A"}`),
						ObjectPrimaryValue: "Site A",
					},
				},
				// this is still a create because the ref id is set
				// and the object was created in the same change set
				DeviationName: strPtr("Site Site A created"),
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site modify subordinate object - is a modification",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
							Region: &diodepb.Region{
								Name: "Region A",
							},
						},
					},
				},
			},
			response: &mockGenerateDiffResponse{
				result: &netboxdiodeplugin.ChangeSetResult{
					ChangeSet: &netboxdiodeplugin.ChangeSet{
						Changes: []netboxdiodeplugin.Change{
							{
								ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "create",
								ObjectType:         "dcim.region",
								ObjectID:           nil,
								RefID:              strPtr("ref-1"),
								Data:               json.RawMessage(`{"name": "Region A"}`),
								ObjectPrimaryValue: "Region A",
							},
							{
								ID:                 "5663a77f-9bad-4981-afe9-77d8a9f2b8b5",
								ChangeType:         "noop",
								ObjectType:         "dcim.site",
								ObjectID:           intPtr(1),
								RefID:              nil,
								Data:               json.RawMessage(`{"name": "Site A", "region": "ref-1"}`),
								Before:             json.RawMessage(`{}`),
								ObjectPrimaryValue: "Site A",
								NewRefs:            []string{"region"},
							},
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				Changes: []changeset.Change{
					{
						ID:                 "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:         "create",
						ObjectType:         "dcim.region",
						ObjectID:           nil,
						RefID:              strPtr("ref-1"),
						After:              json.RawMessage(`{"name": "Region A"}`),
						ObjectPrimaryValue: "Region A",
					},
					{
						ID:                 "5663a77f-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:         "noop",
						ObjectType:         "dcim.site",
						ObjectID:           intPtr(1),
						RefID:              nil,
						After:              json.RawMessage(`{"name": "Site A", "region": "ref-1"}`),
						Before:             json.RawMessage(`{}`),
						ObjectPrimaryValue: "Site A",
						NewRefs:            []string{"region"},
					},
				},
				// this is still a modification because the subordinate object
				// changed, even though the primary object is the same
				DeviationName: strPtr("Site Site A modified"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := mocks.NewNetBoxAPI(t)
			mockClient.EXPECT().GenerateDiff(ctx, mock.Anything).Return(tt.response.result, tt.response.err)

			cs, err := differ.Diff(ctx, tt.ingestEntity, "", mockClient)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, len(tt.wantChangeSet.Changes), len(cs.Changes))
			if tt.wantChangeSet.DeviationName != nil {
				assert.NotNil(t, cs.DeviationName)
				assert.Equal(t, *tt.wantChangeSet.DeviationName, *cs.DeviationName)
			} else {
				assert.Nil(t, cs.DeviationName)
			}
			for _, change := range cs.Changes {
				assert.NotNil(t, change.ID)
				assert.NotNil(t, change.ObjectType)
				assert.NotEmpty(t, change.ObjectPrimaryValue)
				assert.NotNil(t, change.ChangeType)
				assert.NotNil(t, change.After)
			}
		})
	}
}
