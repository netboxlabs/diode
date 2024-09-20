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

func TestIpamPrepare(t *testing.T) {
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
			name: "[P1] ingest ipam.ipaddress with address and interface - existing object not found - create",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet0/0/0",
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
							Status: (*netbox.DcimDeviceStatus)(strPtr(string(netbox.DcimDeviceStatusActive))),
						},
					},
				},
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
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
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: nil,
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
			name: "[P1] ingest ipam.ipaddress with address and a new interface - existing IP address and interface not found - create an interface and IP address",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet0/0/0",
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
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: nil,
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
			name: "[P1] ingest ipam.ipaddress with address and a new interface - IP address found assigned to a different interface - create the interface and the IP address",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet1/0/1",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet1/0/1", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: nil,
					},
				},
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet1/0/1", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: &netbox.IpamIPAddress{
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
			name: "[P1] ingest ipam.ipaddress with assigned interface - existing IP address found assigned a different device - create IP address with a new assigned object (interface)",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet1/0/1",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet1/0/1", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
						Interface: &netbox.DcimInterface{
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
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet1/0/1", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: &netbox.IpamIPAddress{
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
			name: "[P1] ingest ipam.ipaddress with address and interface - existing IP address found with same interface assigned - no update needed",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet0/0/0",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
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
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: &netbox.IpamIPAddress{
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
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P1] ingest ipam.ipaddress with address only - existing IP address found without interface assigned - no update needed",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address: "192.168.0.1/22",
						},
					},
				},
			},
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: &netbox.IpamIPAddress{
							ID:      1,
							Address: "192.168.0.1/22",
							Status:  &netbox.DefaultIPAddressStatus,
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
			name: "[P1] ingest ipam.ipaddress with address and new description - existing IP address found - update IP address with new description",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{
							Address:     "192.168.0.1/22",
							Description: strPtr("new description"),
							AssignedObject: &diodepb.IPAddress_Interface{
								Interface: &diodepb.Interface{
									Name: "GigabitEthernet0/0/0",
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
				{
					objectType:     "dcim.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "GigabitEthernet0/0/0", "device__name": "undefined", "device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.DcimInterfaceDataWrapper{
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
				{
					objectType:     "ipam.ipaddress",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.1/22", "interface__name": "GigabitEthernet0/0/0", "interface__device__name": "undefined", "interface__device__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.IpamIPAddressDataWrapper{
						IPAddress: &netbox.IpamIPAddress{
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
			name: "[P1] ingest empty ipam.ipaddress - error",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.ipaddress",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_IpAddress{
						IpAddress: &diodepb.IPAddress{},
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
			name: "[P2] ingest ipam.prefix with prefix only - existing object not found - create prefix and site (placeholder)",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.prefix",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Prefix{
						Prefix: &diodepb.Prefix{
							Prefix: "192.168.0.0/32",
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
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						Prefix: nil,
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
			name: "[P2] ingest ipam.prefix with prefix only - existing object and its related objects found - do nothing",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.prefix",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Prefix{
						Prefix: &diodepb.Prefix{
							Prefix: "192.168.0.0/32",
							Site: &diodepb.Site{
								Name: "undefined",
							},
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
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						Prefix: &netbox.IpamPrefix{
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
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest ipam.prefix with empty site",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.prefix",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Prefix{
						Prefix: &diodepb.Prefix{
							Prefix: "192.168.0.0/32",
							Site:   &diodepb.Site{},
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
					objectType:     "ipam.prefix",
					objectID:       0,
					queryParams:    map[string]string{"q": "192.168.0.0/32"},
					objectChangeID: 0,
					object: &netbox.IpamPrefixDataWrapper{
						Prefix: &netbox.IpamPrefix{
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
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P2] ingest ipam.prefix with prefix and a tag - existing object found - create tag and update prefix",
			ingestEntity: changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "ipam.prefix",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Prefix{
						Prefix: &diodepb.Prefix{
							Prefix: "192.168.0.0/32",
							Tags: []*diodepb.Tag{
								{
									Name: "tag 100",
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
						Prefix: &netbox.IpamPrefix{
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
