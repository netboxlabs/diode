package differ_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
)

func TestGenDeviationName(t *testing.T) {
	type mockRetrieveObjectState struct {
		objectType     string
		objectID       int
		queryParams    map[string]string
		objectChangeID int
		object         netbox.ComparableData
	}
	tests := []struct {
		name                 string
		ingestEntity         differ.IngestEntity
		retrieveObjectStates []mockRetrieveObjectState
		wantChangeSet        changeset.ChangeSet
		wantErr              bool
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
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
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.site",
						ObjectID:      nil,
						ObjectVersion: nil,
						After: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				DeviationName: strPtr("Site Site A discovered"),
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID:   "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:     []changeset.Change{},
				DeviationName: nil,
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 1"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: &netbox.Tag{
							ID:   1,
							Name: "tag 1",
							Slug: "tag-1",
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 2"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						After: &netbox.Tag{
							Name: "tag 2",
							Slug: "tag-2",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.site",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Before: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
						After: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
								{
									Name: "tag 2",
									Slug: "tag-2",
								},
							},
						},
					},
				},
				DeviationName: strPtr("Site Site A modified"),
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest empty dcim.site - error",
			ingestEntity: differ.IngestEntity{
				RequestID:  "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ObjectType: "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID:   "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:     []changeset.Change{},
				DeviationName: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := mocks.NewNetBoxAPI(t)

			for _, m := range tt.retrieveObjectStates {
				mockClient.EXPECT().RetrieveObjectState(ctx, netboxdiodeplugin.RetrieveObjectStateQueryParams{
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

			cs, err := differ.Diff(ctx, tt.ingestEntity, "", mockClient)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, len(tt.wantChangeSet.ChangeSet), len(cs.ChangeSet))
			assert.Equal(t, tt.wantChangeSet.DeviationName, cs.DeviationName)
			for _, change := range cs.ChangeSet {
				assert.NotNil(t, change.ChangeID)
				assert.NotNil(t, change.ObjectType)
				assert.NotNil(t, change.ChangeType)
				assert.NotNil(t, change.After)
			}
		})
	}
}

func TestChangesWithBeforeAfter(t *testing.T) {
	type mockRetrieveObjectState struct {
		objectType     string
		objectID       int
		queryParams    map[string]string
		objectChangeID int
		object         netbox.ComparableData
	}
	tests := []struct {
		name                 string
		ingestEntity         differ.IngestEntity
		retrieveObjectStates []mockRetrieveObjectState
		wantChangeSet        changeset.ChangeSet
		wantErr              bool
	}{
		{
			name: "existing object not found - change.before is nil",
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
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
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.site",
						ObjectID:      nil,
						ObjectVersion: nil,
						After: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				DeviationName: strPtr("Site Site A discovered"),
			},
			wantErr: false,
		},
		{
			name: "existing object found - no update required - no changes",
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID:   "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:     []changeset.Change{},
				DeviationName: nil,
			},
			wantErr: false,
		},
		{
			name: "existing object found - update required - changes with before and after",
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
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 1"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: &netbox.Tag{
							ID:   1,
							Name: "tag 1",
							Slug: "tag-1",
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 2"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						After: &netbox.Tag{
							Name: "tag 2",
							Slug: "tag-2",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.site",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Before: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
						After: &netbox.DcimSite{
							ID:     1,
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   3,
									Name: "tag 3",
									Slug: "tag-3",
								},
								{
									Name: "tag 2",
									Slug: "tag-2",
								},
							},
						},
					},
				},
				DeviationName: strPtr("Site Site A modified"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockClient := mocks.NewNetBoxAPI(t)

			for _, m := range tt.retrieveObjectStates {
				mockClient.EXPECT().RetrieveObjectState(ctx, netboxdiodeplugin.RetrieveObjectStateQueryParams{
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

			cs, err := differ.Diff(ctx, tt.ingestEntity, "", mockClient)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, len(tt.wantChangeSet.ChangeSet), len(cs.ChangeSet))
			for i, change := range cs.ChangeSet {
				assert.NotNil(t, change.After)
				assert.Equal(t, tt.wantChangeSet.ChangeSet[i].After, change.After)
				if change.ObjectID != nil {
					assert.NotNil(t, change.Before)
					assert.Equal(t, tt.wantChangeSet.ChangeSet[i].Before, change.Before)
				}
			}
		})
	}
}
