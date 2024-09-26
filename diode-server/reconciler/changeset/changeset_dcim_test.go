package changeset_test

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
)

func TestDcimPrepare(t *testing.T) {
	type mockRetrieveObjectState struct {
		objectType     string
		objectID       int
		queryParams    map[string]string
		objectChangeID int
		object         netbox.ComparableData
	}
	tests := []struct {
		name                 string
		ingestEntity         changeset.IngestEntity
		retrieveObjectStates []mockRetrieveObjectState
		wantChangeSet        changeset.ChangeSet
		wantErr              bool
	}{
		{
			name: "[P1] ingest dcim.site with name only - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site with name only - existing object found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
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
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest dcim.site with tags - existing object found - update with new tags",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
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
						Data: &netbox.Tag{
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
						Data: &netbox.DcimSite{
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
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest empty dcim.site - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P2] ingest dcim.devicerole with name only - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name: "WAN Router",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
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
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest dcim.devicerole with name only - existing object found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name: "WAN Router",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
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
			name: "[P2] ingest dcim.devicerole with name and new description - existing object found - update",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name:        "WAN Router",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:          1,
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("111222"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							ID:          1,
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("111222"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest dcim.devicerole with same color - existing object found - nothing to update",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name:        "WAN Router",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Color:       "111222",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:          1,
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("111222"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
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
			name: "[P2] ingest empty dcim.devicerole - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P3] ingest dcim.manufacturer with name only - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.manufacturer",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Manufacturer{
						Manufacturer: &diodepb.Manufacturer{
							Name: "Cisco",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "Cisco",
							Slug: "cisco",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P3] ingest dcim.manufacturer with name only - existing object found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.manufacturer",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Manufacturer{
						Manufacturer: &diodepb.Manufacturer{
							Name: "Cisco",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "Cisco",
							Slug: "cisco",
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
			name: "[P3] ingest empty dcim.manufacturer - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.manufacturer",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Manufacturer{
						Manufacturer: &diodepb.Manufacturer{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P4] ingest dcim.devicetype with model only - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model: "ISR4321",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "undefined",
							Slug: "undefined",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P4] ingest dcim.devicetype with model only - existing object found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model: "ISR4321",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
							Tags: []*netbox.Tag{},
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
			name: "[P4] ingest empty dcim.devicetype - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P5] ingest dcim.devicetype with manufacturer - existing object not found - create manufacturer and devicetype",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model: "ISR4321",
							Manufacturer: &diodepb.Manufacturer{
								Name: "Cisco",
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "Cisco",
							Slug: "cisco",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "Cisco",
								Slug: "cisco",
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P5] ingest dcim.devicetype with new manufacturer - existing object found - create manufacturer and update devicetype",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model: "ISR4321",
							Manufacturer: &diodepb.Manufacturer{
								Name: "Cisco",
								Tags: []*diodepb.Tag{
									{
										Name: "tag 1",
									},
									{
										Name: "tag 10",
									},
									{
										Name: "tag 11",
									},
								},
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 3",
								},
							},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   2,
							Name: "Cisco",
							Slug: "cisco",
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   5,
									Name: "tag 5",
									Slug: "tag-5",
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
					queryParams:    map[string]string{"q": "tag 10"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: &netbox.Tag{
							ID:   10,
							Name: "tag 10",
							Slug: "tag-10",
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 11"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 3"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Tags: []*netbox.Tag{
									{
										ID:   4,
										Name: "tag 4",
										Slug: "tag-4",
									},
								},
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
							},
						},
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
						Data: &netbox.Tag{
							Name: "tag 3",
							Slug: "tag-3",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.Tag{
							Name: "tag 11",
							Slug: "tag-11",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      intPtr(2),
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							ID:   2,
							Name: "Cisco",
							Slug: "cisco",
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
								{
									ID:   5,
									Name: "tag 5",
									Slug: "tag-5",
								},
								{
									ID:   10,
									Name: "tag 10",
									Slug: "tag-10",
								},
								{
									Name: "tag 11",
									Slug: "tag-11",
								},
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   2,
								Name: "Cisco",
								Slug: "cisco",
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
								{
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P5.2] ingest dcim.devicetype with new manufacturer - existing object found - create manufacturer and update devicetype",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model:       "ISR4321",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 3",
								},
							},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
							Tags: []*netbox.Tag{
								{
									ID:   4,
									Name: "tag 4",
									Slug: "tag-4",
								},
							},
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 3"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Tags: []*netbox.Tag{
									{
										ID:   4,
										Name: "tag 4",
										Slug: "tag-4",
									},
								},
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
							},
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.Tag{
							Name: "tag 3",
							Slug: "tag-3",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID: 1,
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
								{
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P5.3] ingest dcim.devicetype with new manufacturer - existing object found - update devicetype with new existing manufacturer",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicetype",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceType{
						DeviceType: &diodepb.DeviceType{
							Model: "ISR4321",
							Manufacturer: &diodepb.Manufacturer{
								Name: "Cisco",
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 3",
								},
							},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "Cisco",
							Slug: "cisco",
							Tags: []*netbox.Tag{
								{
									ID:   4,
									Name: "tag 4",
									Slug: "tag-4",
								},
							},
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 3"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Tags: []*netbox.Tag{
									{
										ID:   4,
										Name: "tag 4",
										Slug: "tag-4",
									},
								},
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
							},
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "extras.tag",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.Tag{
							Name: "tag 3",
							Slug: "tag-3",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID: 1,
							},
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							PartNumber:  strPtr("xyz123"),
							Tags: []*netbox.Tag{
								{
									ID:   2,
									Name: "tag 2",
									Slug: "tag-2",
								},
								{
									Name: "tag 3",
									Slug: "tag-3",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest dcim.device with name only - existing object not found - create device and all related objects (using placeholders)",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: nil,
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
						Data: &netbox.DcimSite{
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "undefined",
							Slug: "undefined",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "undefined",
								Slug:  "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest dcim.device with name only - existing object and its related objects found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "undefined",
								Slug:  "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
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
			name: "[P6] ingest dcim.device with empty site",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							Site: &diodepb.Site{},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "undefined",
								Slug:  "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
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
			name: "[P7] ingest dcim.device - existing object not found - create device and all related objects",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 1",
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
						Site: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: nil,
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "undefined",
							Slug: "undefined",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P7] ingest dcim.device with device type having manufacturer defined - existing object not found - create device and all related objects",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
								Manufacturer: &diodepb.Manufacturer{
									Name: "Cisco",
								},
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 1",
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
						Site: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: nil,
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "Cisco",
							Slug: "cisco",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "Cisco",
								Slug: "cisco",
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									Name: "Cisco",
									Slug: "cisco",
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest empty dcim.device - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P7] ingest dcim.device - existing object found - create missing related objects and update device",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*diodepb.Tag{
								{
									Name: "tag 1",
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
						Site: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
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
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "ISR4322",
								Slug:  "isr4322",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID: 1,
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.device",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									ID: 1,
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
							Tags: []*netbox.Tag{
								{
									ID:   1,
									Name: "tag 1",
									Slug: "tag-1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object not found - create device and all related objects",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Platform: &diodepb.Platform{
								Name: "Cisco IOS 15.6",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: nil,
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.platform",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimPlatform{
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name: "undefined",
							Slug: "undefined",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								Name: "Cisco IOS 15.6",
								Slug: "cisco-ios-15-6",
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object found - create missing related objects and update device",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Platform: &diodepb.Platform{
								Name: "Cisco IOS 15.6",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "ISR4322",
								Slug:  "isr4322",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
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
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID: 1,
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.device",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									ID: 1,
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object found - create some missing related objects, use other existing one and update device",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site A",
							},
							Platform: &diodepb.Platform{
								Name: "Cisco IOS 15.6",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456-2"),
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("111111"),
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "ISR4322",
								Slug:  "isr4322",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("111111"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "Cisco IOS 15.6",
								Slug: "cisco-ios-15-6",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.device",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID: 1,
							},
							DeviceType: &netbox.DcimDeviceType{
								ID: 1,
							},
							Role: &netbox.DcimDeviceRole{
								ID: 1,
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456-2"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8.1] ingest dcim.device with partial data - existing object found - create missing related objects and update device",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Platform: &diodepb.Platform{
								Name: "Cisco IOS 15.6",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     2,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "ISR4322",
								Slug:  "isr4322",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "Cisco IOS 15.6",
								Slug: "cisco-ios-15-6",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicetype",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceType{
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID: 1,
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("000000"),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.device",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID: 2,
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									ID: 1,
								},
							},
							Role: &netbox.DcimDeviceRole{
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8.2] ingest dcim.device - existing object found - no changes to apply",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "router01",
							DeviceType: &diodepb.DeviceType{
								Model: "ISR4321",
							},
							Role: &diodepb.Role{
								Name: "WAN Router",
							},
							Site: &diodepb.Site{
								Name: "Site B",
							},
							Platform: &diodepb.Platform{
								Name: "Cisco IOS 15.6",
							},
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site B"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site B",
							Slug:   "site-b",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "ISR4321",
							Slug:  "isr4321",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "WAN Router",
							Slug:  "wan-router",
							Color: strPtr("111111"),
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site B"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("111111"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "Cisco IOS 15.6",
								Slug: "cisco-ios-15-6",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status:      (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							Serial:      strPtr("123456"),
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
			name: "[P9] ingest dcim.site with name, status and description - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name:        "Site A",
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
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
						Data: &netbox.DcimSite{
							Name:        "Site A",
							Slug:        "site-a",
							Status:      (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P9] ingest dcim.site with name, status and new description - existing object found - update",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name:        "Site A",
							Status:      "active",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
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
							ID:          1,
							Name:        "Site A",
							Slug:        "site-a",
							Status:      (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.site",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimSite{
							ID:          1,
							Name:        "Site A",
							Slug:        "site-a",
							Status:      (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P10] ingest dcim.manufacturer with name and description - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.manufacturer",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Manufacturer{
						Manufacturer: &diodepb.Manufacturer{
							Name:        "Cisco",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							Name:        "Cisco",
							Slug:        "cisco",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P10] ingest dcim.manufacturer with name and new description - existing object found - update",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.manufacturer",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Manufacturer{
						Manufacturer: &diodepb.Manufacturer{
							Name:        "Cisco",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:          1,
							Name:        "Cisco",
							Slug:        "cisco",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.manufacturer",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimManufacturer{
							ID:          1,
							Name:        "Cisco",
							Slug:        "cisco",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P11] ingest dcim.devicerole with name and additional attributes - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name:        "WAN Router",
							Color:       "509415",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
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
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("509415"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P11] ingest dcim.devicerole with name and new additional attributes - existing object found - update",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.devicerole",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_DeviceRole{
						DeviceRole: &diodepb.Role{
							Name:        "WAN Router",
							Color:       "ffffff",
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:          1,
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("509415"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.devicerole",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDeviceRole{
							ID:          1,
							Name:        "WAN Router",
							Slug:        "wan-router",
							Color:       strPtr("ffffff"),
							Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P12] ingest empty dcim.platform - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.platform",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Platform{
						Platform: &diodepb.Platform{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P13] ingest dcim.interface with name only - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.interface",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{
						Interface: &diodepb.Interface{
							Name: "GigabitEthernet0/0/0",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: nil,
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: nil,
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "undefined",
							Site: &netbox.DcimSite{
								ID: 1,
							},
							DeviceType: &netbox.DcimDeviceType{
								ID: 1,
							},
							Role: &netbox.DcimDeviceRole{
								ID: 1,
							},
							Platform: &netbox.DcimPlatform{
								ID: 1,
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.interface",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimInterface{
							Name: "GigabitEthernet0/0/0",
							Device: &netbox.DcimDevice{
								Name: "undefined",
								Site: &netbox.DcimSite{
									ID: 1,
								},
								DeviceType: &netbox.DcimDeviceType{
									ID: 1,
								},
								Role: &netbox.DcimDeviceRole{
									ID: 1,
								},
								Platform: &netbox.DcimPlatform{
									ID: 1,
								},
								Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							},
							Type: strPtr(netbox.DefaultInterfaceType),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P13] ingest dcim.interface with name and device - existing object found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.interface",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{
						Interface: &diodepb.Interface{
							Name: "GigabitEthernet0/0/0",
							Device: &diodepb.Device{
								Name: "router01",
							},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "router01", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: &netbox.DcimInterface{
							ID:   1,
							Name: "GigabitEthernet0/0/0",
							Device: &netbox.DcimDevice{
								ID:   1,
								Name: "router01",
								Site: &netbox.DcimSite{
									ID:     1,
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								DeviceType: &netbox.DcimDeviceType{
									ID:    1,
									Model: "undefined",
									Slug:  "undefined",
									Manufacturer: &netbox.DcimManufacturer{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
								},
								Role: &netbox.DcimDeviceRole{
									ID:    1,
									Name:  "undefined",
									Slug:  "undefined",
									Color: strPtr("000000"),
								},
								Platform: &netbox.DcimPlatform{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
									Manufacturer: &netbox.DcimManufacturer{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
								},
								Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							},
							Type: strPtr(netbox.DefaultInterfaceType),
						},
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "undefined",
								Slug:  "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
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
			name: "[P13] ingest dcim.interface with name, device and new label - existing object found - update with new label",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.interface",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{
						Interface: &diodepb.Interface{
							Name: "GigabitEthernet0/0/0",
							Device: &diodepb.Device{
								Name: "router01",
							},
							Label: strPtr("WAN"),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "router01", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: &netbox.DcimInterface{
							ID:   1,
							Name: "GigabitEthernet0/0/0",
							Device: &netbox.DcimDevice{
								ID:   1,
								Name: "router01",
								Site: &netbox.DcimSite{
									ID:     1,
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								DeviceType: &netbox.DcimDeviceType{
									ID:    1,
									Model: "undefined",
									Slug:  "undefined",
									Manufacturer: &netbox.DcimManufacturer{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
								},
								Role: &netbox.DcimDeviceRole{
									ID:    1,
									Name:  "undefined",
									Slug:  "undefined",
									Color: strPtr("000000"),
								},
								Platform: &netbox.DcimPlatform{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
									Manufacturer: &netbox.DcimManufacturer{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
								},
								Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
							},
							Type: strPtr(netbox.DefaultInterfaceType),
						},
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "undefined",
							Slug:   "undefined",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "undefined",
							Slug:  "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "undefined",
								Slug:  "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.interface",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimInterface{
							ID:   1,
							Name: "GigabitEthernet0/0/0",
							Device: &netbox.DcimDevice{
								ID: 1,
							},
							Type:  strPtr(netbox.DefaultInterfaceType),
							Label: strPtr("WAN"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P13] ingest empty dcim.interface - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.interface",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{
						Interface: &diodepb.Interface{},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: true,
		},
		{
			name: "[P14] ingest dcim.device with device type and manufacturer - device type and manufacturer objects found - create device with existing device type and manufacturer",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: "Device A",
							DeviceType: &diodepb.DeviceType{
								Model: "Device Type A",
								Manufacturer: &diodepb.Manufacturer{
									Name: "Manufacturer A",
								},
							},
							Role: &diodepb.Role{
								Name: "Role ABC",
							},
							Platform: &diodepb.Platform{
								Name: "Platform A",
								Manufacturer: &diodepb.Manufacturer{
									Name: "Manufacturer A",
								},
							},
							Serial: strPtr("123456"),
							Site:   &diodepb.Site{Name: "Site ABC"},
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site ABC"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     1,
							Name:   "Site ABC",
							Slug:   "site-abc",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: &netbox.DcimManufacturer{
							ID:   1,
							Name: "Manufacturer A",
							Slug: "manufacturer-a",
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Platform A", "manufacturer__name": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: &netbox.DcimPlatform{
							ID:   1,
							Name: "Platform A",
							Slug: "platform-a",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "Manufacturer A",
								Slug: "manufacturer-a",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "Device Type A", "manufacturer__name": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: &netbox.DcimDeviceType{
							ID:    1,
							Model: "Device Type A",
							Slug:  "device-type-a",
							Manufacturer: &netbox.DcimManufacturer{
								ID:   1,
								Name: "Manufacturer A",
								Slug: "manufacturer-a",
							},
						},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "Role ABC"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    1,
							Name:  "Role ABC",
							Slug:  "role-abc",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "Device A", "site__name": "Site ABC"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   1,
							Name: "Device A",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "Site ABC",
								Slug:   "site-abc",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    1,
								Model: "Device Type A",
								Slug:  "device-type-a",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "Manufacturer A",
									Slug: "manufacturer-a",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "Role ABC",
								Slug:  "role-abc",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   1,
								Name: "Platform A",
								Slug: "platform-a",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
									Name: "Manufacturer A",
									Slug: "manufacturer-a",
								},
							},
							Serial: strPtr("123456"),
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
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
			name: "[P15] ingest dcim.interface with name, mtu, device with site - device exists for platform Arista - create interface with existing device and platform",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.interface",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{
						Interface: &diodepb.Interface{
							Name: "Ethernet2",
							Device: &diodepb.Device{
								Name: "CEOS1",
								Site: &diodepb.Site{
									Name: "default_namespace",
								},
							},
							Mtu: int32Ptr(1500),
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "Ethernet2", "device__name": "CEOS1", "device__site__name": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: &netbox.DcimDeviceRole{
							ID:    89,
							Name:  "undefined",
							Slug:  "undefined",
							Color: strPtr("000000"),
						},
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: &netbox.DcimSite{
							ID:     21,
							Name:   "default_namespace",
							Slug:   "default_namespace",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "CEOS1", "site__name": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						Device: &netbox.DcimDevice{
							ID:   111,
							Name: "CEOS1",
							Site: &netbox.DcimSite{
								ID:     21,
								Name:   "default_namespace",
								Slug:   "default_namespace",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								ID:    10,
								Model: "cEOSLab",
								Slug:  "ceoslab",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   15,
									Name: "Arista",
									Slug: "arista",
								},
							},
							Role: &netbox.DcimDeviceRole{
								ID:    89,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
							Platform: &netbox.DcimPlatform{
								ID:   68,
								Name: "eos:4.29.0.2F-29226602.42902F (engineering build)",
								Slug: "eos-4-29-0-2f-29226602-42902f-engineering-build",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   15,
									Name: "Arista",
									Slug: "arista",
								},
							},
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
						},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.interface",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimInterface{
							Name: "Ethernet2",
							Device: &netbox.DcimDevice{
								ID: 111,
							},
							MTU:  intPtr(1500),
							Type: strPtr(netbox.DefaultInterfaceType),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			cs, err := changeset.Prepare(tt.ingestEntity, mockClient)
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

func strPtr(s string) *string { return &s }
func intPtr(d int) *int       { return &d }
func int32Ptr(d int32) *int32 { return &d }
