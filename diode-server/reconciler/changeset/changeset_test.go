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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.site",
				"entity": {
					"Site": {}
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("000000"),
							},
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:          1,
								Name:        "WAN Router",
								Slug:        "wan-router",
								Color:       strPtr("111222"),
								Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean sed molestie felis."),
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:          1,
								Name:        "WAN Router",
								Slug:        "wan-router",
								Color:       strPtr("111222"),
								Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
							},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicerole",
				"entity": {
					"DeviceRole": {}
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "Cisco",
								Slug: "cisco",
							},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.manufacturer",
				"entity": {
					"Manufacturer": {}
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P4] ingest empty dcim.devicetype - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.devicetype",
				"entity": {
					"DeviceType": {}
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{Field: nil},
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest dcim.device with empty site",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"site": {}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
						"serial": "123456",
						"tags": [
							{
								"name": "tag 1"
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{Field: nil},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "router01",
						"device_type": {
							"model": "ISR4321",
							"manufacturer": {
								"name": "Cisco"
							}
						},
						"role": {
							"name": "WAN Router"
						},
						"site": {
							"name": "Site A"
						},
						"status": "active",
						"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
						"serial": "123456",
						"tags": [
							{
								"name": "tag 1"
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{Field: nil},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {}
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
						"serial": "123456",
						"tags": [
							{
								"name": "tag 1"
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{Field: nil},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{Field: nil},
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
								ID:   1,
								Name: "Cisco IOS 15.6",
								Slug: "cisco-ios-15-6",
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "Site A",
								Slug:   "site-a",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("111111"),
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
					queryParams:    map[string]string{"q": "Site B"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "Site B",
								Slug:   "site-b",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
					queryParams:    map[string]string{"q": "ISR4321", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "WAN Router",
								Slug:  "wan-router",
								Color: strPtr("111111"),
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Cisco IOS 15.6"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "router01", "site__name": "Site B"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
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
					queryParams:    map[string]string{"q": "Site A"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:          1,
								Name:        "Site A",
								Slug:        "site-a",
								Status:      (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
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
					queryParams:    map[string]string{"q": "Cisco"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:          1,
								Name:        "Cisco",
								Slug:        "cisco",
								Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{Field: nil},
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
					queryParams:    map[string]string{"q": "WAN Router"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:          1,
								Name:        "WAN Router",
								Slug:        "wan-router",
								Color:       strPtr("509415"),
								Description: strPtr("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.platform",
				"entity": {
					"Platform": {}
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
			name: "[P13] ingest dcim.interface with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.interface",
				"entity": {
					"Interface": {
						"name": "GigabitEthernet0/0/0"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{Field: nil},
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{Field: nil},
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.interface",
				"entity": {
					"Interface": {
						"name": "GigabitEthernet0/0/0",
						"device": {
							"name": "router01"
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "router01", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
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
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P13] ingest dcim.interface with name, device and new label - existing object found - update with new label",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.interface",
				"entity": {
					"Interface": {
						"name": "GigabitEthernet0/0/0",
						"device": {
							"name": "router01"
						},
						"label": "WAN"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "router01", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
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
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
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
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.interface",
				"entity": {
					"Interface": {}
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
			name: "[P14] ingest ipam.ipaddress with address only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{Field: nil},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "ipam.ipaddress",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest ipam.ipaddress with address and interface - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet0/0/0"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
								ID:   1,
								Name: "GigabitEthernet0/0/0",
								Type: strPtr(netbox.DefaultInterfaceType),
								Device: &netbox.DcimDevice{
									ID:   1,
									Name: "undefined",
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
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{Field: nil},
					},
				},
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "ipam.ipaddress",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
							AssignedObject: &netbox.IPAddressInterface{
								Interface: &netbox.DcimInterface{
									ID: 1,
									Device: &netbox.DcimDevice{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest ipam.ipaddress with address and a new interface - existing IP address and interface not found - create an interface and IP address",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet0/0/0"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{Field: nil},
					},
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{Field: nil},
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
							Name: "GigabitEthernet0/0/0",
							Type: strPtr(netbox.DefaultInterfaceType),
							Device: &netbox.DcimDevice{
								ID: 1,
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "ipam.ipaddress",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
							AssignedObject: &netbox.IPAddressInterface{
								Interface: &netbox.DcimInterface{
									Name: "GigabitEthernet0/0/0",
									Type: strPtr(netbox.DefaultInterfaceType),
									Device: &netbox.DcimDevice{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest ipam.ipaddress with address and a new interface - IP address found assigned to a different interface - create the interface and the IP address",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet1/0/1"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet1/0/1", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{Field: nil},
					},
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet1/0/1", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{
							Field: &netbox.IpamIPAddress{
								ID:      1,
								Address: "192.168.0.1/22",
								Status:  &netbox.DefaultIPAddressStatus,
								AssignedObject: &netbox.IPAddressInterface{
									Interface: &netbox.DcimInterface{
										ID:   1,
										Name: "GigabitEthernet0/0/0",
										Type: strPtr(netbox.DefaultInterfaceType),
										Device: &netbox.DcimDevice{
											ID:   1,
											Name: "undefined",
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
							Name: "GigabitEthernet1/0/1",
							Type: strPtr(netbox.DefaultInterfaceType),
							Device: &netbox.DcimDevice{
								ID: 1,
							},
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "ipam.ipaddress",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
							AssignedObject: &netbox.IPAddressInterface{
								Interface: &netbox.DcimInterface{
									Name: "GigabitEthernet1/0/1",
									Type: strPtr(netbox.DefaultInterfaceType),
									Device: &netbox.DcimDevice{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest ipam.ipaddress with assigned interface - existing IP address found assigned a different device - create IP address with a new assigned object (interface)",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet1/0/1"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet1/0/1", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
								ID:   2,
								Name: "GigabitEthernet1/0/1",
								Type: strPtr(netbox.DefaultInterfaceType),
								Device: &netbox.DcimDevice{
									ID:   1,
									Name: "undefined",
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
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet1/0/1", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{
							Field: &netbox.IpamIPAddress{
								ID:      1,
								Address: "192.168.0.1/22",
								Status:  &netbox.DefaultIPAddressStatus,
								AssignedObject: &netbox.IPAddressInterface{
									Interface: &netbox.DcimInterface{
										ID:   1,
										Name: "GigabitEthernet0/0/0",
										Type: strPtr(netbox.DefaultInterfaceType),
										Device: &netbox.DcimDevice{
											ID:   1,
											Name: "undefined",
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
						ObjectType:    "ipam.ipaddress",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
							AssignedObject: &netbox.IPAddressInterface{
								Interface: &netbox.DcimInterface{
									ID: 2,
									Device: &netbox.DcimDevice{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest ipam.ipaddress with address and interface - existing IP address found with same interface assigned - no update needed",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet0/0/0"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
								ID:   1,
								Name: "GigabitEthernet0/0/0",
								Type: strPtr(netbox.DefaultInterfaceType),
								Device: &netbox.DcimDevice{
									ID:   1,
									Name: "undefined",
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
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{
							Field: &netbox.IpamIPAddress{
								ID:      1,
								Address: "192.168.0.1/22",
								Status:  &netbox.DefaultIPAddressStatus,
								AssignedObject: &netbox.IPAddressInterface{
									Interface: &netbox.DcimInterface{
										ID:   1,
										Name: "GigabitEthernet0/0/0",
										Type: strPtr(netbox.DefaultInterfaceType),
										Device: &netbox.DcimDevice{
											ID:   1,
											Name: "undefined",
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
			name: "[P14] ingest ipam.ipaddress with address only - existing IP address found without interface assigned - no update needed",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{
							Field: &netbox.IpamIPAddress{
								ID:      1,
								Address: "192.168.0.1/22",
								Status:  &netbox.DefaultIPAddressStatus,
							},
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
			name: "[P14] ingest ipam.ipaddress with address and new description - existing IP address found - update IP address with new description",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IpAddress": {
						"address": "192.168.0.1/22",
						"description": "new description",
						"AssignedObject": {
							"Interface": {
								"name": "GigabitEthernet0/0/0"
							}
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
								ID:   1,
								Name: "undefined",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{
							Field: &netbox.DcimInterface{
								ID:   1,
								Name: "GigabitEthernet0/0/0",
								Type: strPtr(netbox.DefaultInterfaceType),
								Device: &netbox.DcimDevice{
									ID:   1,
									Name: "undefined",
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
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamIPAddress]{
							Field: &netbox.IpamIPAddress{
								ID:      1,
								Address: "192.168.0.1/22",
								Status:  &netbox.DefaultIPAddressStatus,
								AssignedObject: &netbox.IPAddressInterface{
									Interface: &netbox.DcimInterface{
										ID:   1,
										Name: "GigabitEthernet0/0/0",
										Type: strPtr(netbox.DefaultInterfaceType),
										Device: &netbox.DcimDevice{
											ID:   1,
											Name: "undefined",
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
						ObjectType:    "ipam.ipaddress",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.IpamIPAddress{
							ID:          1,
							Address:     "192.168.0.1/22",
							Status:      &netbox.DefaultIPAddressStatus,
							Description: strPtr("new description"),
							AssignedObject: &netbox.IPAddressInterface{
								Interface: &netbox.DcimInterface{
									ID: 1,
									Device: &netbox.DcimDevice{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P14] ingest empty ipam.ipaddress - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.ipaddress",
				"entity": {
					"IPAddress": {}
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
			name: "[P15] ingest ipam.prefix with prefix only - existing object not found - create prefix and site (placeholder)",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.prefix",
				"entity": {
					"Prefix": {
						"prefix": "192.168.0.0/32"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{Field: nil},
					},
				},
				{
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamPrefix]{Field: nil},
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
						ObjectType:    "ipam.prefix",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.IpamPrefix{
							Prefix: "192.168.0.0/32",
							Site: &netbox.DcimSite{
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							Status: &netbox.DefaultPrefixStatus,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P15] ingest ipam.prefix with prefix only - existing object and its related objects found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.prefix",
				"entity": {
					"Prefix": {
						"prefix": "192.168.0.0/32",
						"site": {
							"name": "undefined"
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamPrefix]{
							Field: &netbox.IpamPrefix{
								ID:     1,
								Prefix: "192.168.0.0/32",
								Site: &netbox.DcimSite{
									ID:     1,
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								Status: &netbox.DefaultPrefixStatus,
							},
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
			name: "[P15] ingest ipam.prefix with empty site",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.prefix",
				"entity": {
					"Prefix": {
						"prefix": "192.168.0.0/32",
						"site": {}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamPrefix]{
							Field: &netbox.IpamPrefix{
								ID:     1,
								Prefix: "192.168.0.0/32",
								Site: &netbox.DcimSite{
									ID:     1,
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								Status: &netbox.DefaultPrefixStatus,
							},
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
			name: "[P15] ingest ipam.prefix with prefix and a tag - existing object found - create tag and update prefix",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "ipam.prefix",
				"entity": {
					"Prefix": {
						"prefix": "192.168.0.0/32",
						"tags": [
							{
								"name": "tag 100"
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
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "extras.tag",
					objectID:       0,
					queryParams:    map[string]string{"q": "tag 100"},
					objectChangeID: 0,
					object: &netbox.TagDataWrapper{
						Tag: nil,
					},
				},
				{
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.IpamPrefix]{
							Field: &netbox.IpamPrefix{
								ID:     1,
								Prefix: "192.168.0.0/32",
								Site: &netbox.DcimSite{
									ID:     1,
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								Status: &netbox.DefaultPrefixStatus,
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
							Name: "tag 100",
							Slug: "tag-100",
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeUpdate,
						ObjectType:    "ipam.prefix",
						ObjectID:      intPtr(1),
						ObjectVersion: nil,
						Data: &netbox.IpamPrefix{
							ID:     1,
							Prefix: "192.168.0.0/32",
							Site: &netbox.DcimSite{
								ID: 1,
							},
							Status: &netbox.DefaultPrefixStatus,
							Tags: []*netbox.Tag{
								{
									Name: "tag 100",
									Slug: "tag-100",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P16] ingest dcim.device with device type and manufacturer - device type and manufacturer objects found - create device with existing device type and manufacturer",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.device",
				"entity": {
					"Device": {
						"name": "Device A",
						"device_type": {
							"model": "Device Type A",
							"manufacturer": {
								"name": "Manufacturer A"
							}
						},
						"role": {
							"name": "Role ABC"
						},
						"platform": {
							"name": "Platform A",
							"manufacturer": {
								"name": "Manufacturer A"
							}
						},
						"serial": "123456",
						"site": {
							"name": "Site ABC"
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "Site ABC"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     1,
								Name:   "Site ABC",
								Slug:   "site-abc",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{
							Field: &netbox.DcimManufacturer{
								ID:   1,
								Name: "Manufacturer A",
								Slug: "manufacturer-a",
							},
						},
					},
				},
				{
					objectType:     "dcim.platform",
					objectID:       0,
					queryParams:    map[string]string{"q": "Platform A", "manufacturer__name": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimPlatformDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimPlatform]{
							Field: &netbox.DcimPlatform{
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
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "Device Type A", "manufacturer__name": "Manufacturer A"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{
							Field: &netbox.DcimDeviceType{
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
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "Role ABC"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    1,
								Name:  "Role ABC",
								Slug:  "role-abc",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "Device A", "site__name": "Site ABC"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P17] ingest dcim.interface with name, mtu, device with site - device exists for platform Arista - create interface with existing device and platform",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "dcim.interface",
				"entity": {
					"Interface": {
						"name": "Ethernet2",
						"device": {
							"name": "CEOS1",
							"site": {
                                "name": "default_namespace"
                            }
						},
						"mtu": 1500
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "Ethernet2", "device__name": "CEOS1", "device__site__name": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimInterface]{Field: nil},
					},
				},
				{
					objectType:     "dcim.manufacturer",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimManufacturerDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimManufacturer]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicetype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "manufacturer__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceTypeDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceType]{Field: nil},
					},
				},
				{
					objectType:     "dcim.devicerole",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceRoleDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDeviceRole]{
							Field: &netbox.DcimDeviceRole{
								ID:    89,
								Name:  "undefined",
								Slug:  "undefined",
								Color: strPtr("000000"),
							},
						},
					},
				},
				{
					objectType:     "dcim.site",
					objectID:       0,
					queryParams:    map[string]string{"q": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimSiteDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimSite]{
							Field: &netbox.DcimSite{
								ID:     21,
								Name:   "default_namespace",
								Slug:   "default_namespace",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
						},
					},
				},
				{
					objectType:     "dcim.device",
					objectID:       0,
					queryParams:    map[string]string{"q": "CEOS1", "site__name": "default_namespace"},
					objectChangeID: 0,
					object: &netbox.DcimDeviceDataWrapper{
						BaseDataWrapper: netbox.BaseDataWrapper[netbox.DcimDevice]{
							Field: &netbox.DcimDevice{
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

func strPtr(s string) *string { return &s }
func intPtr(d int) *int       { return &d }
