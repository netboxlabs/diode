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
			name: "[P1] ingest dcim.site with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
					"Site": {
						"name": "Site A",
						"tags": [
							{
								"name": "tag 1"
							},
							{
								"name": "tag 2"
							}
						]
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
					query:          "tag 1",
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
					query:          "tag 2",
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
			name: "[P2] ingest dcim.devicerole with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
					"DeviceRole": {
						"name": "WAN Router",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
					"DeviceRole": {
						"name": "WAN Router",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"color": "111222"
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
			name: "[P3] ingest dcim.manufacturer with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.manufacturer",
				"entity": {
					"Manufacturer": {
						"name": "Cisco"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.manufacturer",
				"entity": {
					"Manufacturer": {
						"name": "Cisco"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
			name: "[P4] ingest dcim.devicetype with model only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
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
					query:          "ISR4321",
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
			name: "[P5] ingest dcim.devicetype with manufacturer - existing object not found - create manufacturer and devicetype",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321",
						"manufacturer": {
							"name": "Cisco"
						},
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"part_number": "xyz123"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321",
						"manufacturer": {
							"name": "Cisco",
							"tags": [
								{
									"name": "tag 1"
								},
								{
									"name": "tag 10"
								},
								{
									"name": "tag 11"
								}
							]
						},
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"part_number": "xyz123",
						"tags": [
							{
								"name": "tag 3"
							}
						]
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
					query:          "tag 1",
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
					query:          "tag 10",
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
					query:          "tag 11",
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					query:          "tag 3",
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"part_number": "xyz123",
						"tags": [
							{
								"name": "tag 3"
							}
						]
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
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
					query:          "tag 3",
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {
						"model": "ISR4321",
						"manufacturer": {
							"name": "Cisco"
						},
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"part_number": "xyz123",
						"tags": [
							{
								"name": "tag 3"
							}
						]
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
					query:          "tag 3",
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						Site: nil,
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					query:          "router01",
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
						ObjectType:    "dcim.platform",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimPlatform{
							Name: "undefined",
							Slug: "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
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
							Platform: &netbox.DcimPlatform{
								Name: "undefined",
								Slug: "undefined",
								Manufacturer: &netbox.DcimManufacturer{
									Name: "undefined",
									Slug: "undefined",
								},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "undefined",
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
					query:          "undefined",
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
					query:          "undefined",
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
					query:          "undefined",
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
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "undefined",
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
					objectType:     "dcim.device",
					objectID:       0,
					query:          "router01",
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
			name: "[P7] ingest dcim.device - existing object not found - create device and all related objects",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					query:          "router01",
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
						ObjectType:    "dcim.platform",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimPlatform{
							Name: "undefined",
							Slug: "undefined",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
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
								Name: "undefined",
								Slug: "undefined",
								Manufacturer: &netbox.DcimManufacturer{
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
			wantErr: false,
		},
		{
			name: "[P7] ingest dcim.device - existing object found - create missing related objects and update device",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
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
					query:          "undefined",
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
					query:          "ISR4321",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					query:          "router01",
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
								ID:   1,
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
									ID:   1,
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
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object not found - create device and all related objects",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"platform": {
							"name": "Cisco IOS 15.6"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						Manufacturer: nil,
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					query:          "ISR4321",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "Cisco IOS 15.6",
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						Platform: nil,
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					query:          "router01",
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
						ObjectType:    "dcim.platform",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimPlatform{
							Name: "Cisco IOS 15.6",
							Slug: "cisco-ios-15-6",
							Manufacturer: &netbox.DcimManufacturer{
								Name: "undefined",
								Slug: "undefined",
							},
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
								Manufacturer: &netbox.DcimManufacturer{
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
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object found - create missing related objects and update device",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"platform": {
							"name": "Cisco IOS 15.6"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
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
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
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
					query:          "ISR4321",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "Cisco IOS 15.6",
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
					query:          "router01",
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
								ID:   1,
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
									ID:   1,
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
			wantErr: false,
		},
		{
			name: "[P8] ingest dcim.device - existing object found - create some missing related objects, use other existing one and update device",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"platform": {
							"name": "Cisco IOS 15.6"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456-2"
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
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "undefined",
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
					query:          "ISR4321",
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
					query:          "WAN Router",
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
					query:          "Cisco IOS 15.6",
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
					query:          "router01",
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
								ID:     1,
								Name:   "Site A",
								Slug:   "site-a",
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
							Serial:      strPtr("123456-2"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P8.1] ingest dcim.device with partial data - existing object found - create missing related objects and update device",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"platform": {
							"name": "Cisco IOS 15.6"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "undefined",
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
					query:          "undefined",
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
					query:          "ISR4321",
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						DeviceType: nil,
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					query:          "WAN Router",
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						DeviceRole: nil,
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					query:          "Cisco IOS 15.6",
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
					query:          "router01",
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
								ID:   1,
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
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "dcim.device",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							ID:   1,
							Name: "router01",
							Site: &netbox.DcimSite{
								ID:     2,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							DeviceType: &netbox.DcimDeviceType{
								Model: "ISR4321",
								Slug:  "isr4321",
								Manufacturer: &netbox.DcimManufacturer{
									ID:   1,
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
			wantErr: false,
		},
		{
			name: "[P8.2] ingest dcim.device - existing object found - no changes to apply",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321"
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site B"
						},
						"platform": {
							"name": "Cisco IOS 15.6"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					query:          "Site B",
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
					query:          "undefined",
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
					query:          "ISR4321",
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
					query:          "WAN Router",
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
					query:          "Cisco IOS 15.6",
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
					query:          "router01",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
					"Site": {
						"name": "Site A",
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
					"Site": {
						"name": "Site A",
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.manufacturer",
				"entity": {
					"Manufacturer": {
						"name": "Cisco",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.manufacturer",
				"entity": {
					"Manufacturer": {
						"name": "Cisco",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					query:          "Cisco",
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
					"DeviceRole": {
						"name": "WAN Router",
						"color": "509415",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
					"DeviceRole": {
						"name": "WAN Router",
						"color": "ffffff",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."
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
