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
			name: "[P2] ingest virtualization.clustertype with name only - existing object found - do nothing",
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
						ClusterType: &netbox.VirtualizationClusterType{
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
		{
			name: "[P3] ingest virtualization.cluster with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.cluster",
				"entity": {
					"Cluster": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: nil,
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.cluster",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationCluster{
							Name: "Test",
							Group: &netbox.VirtualizationClusterGroup{
								ID: 1,
							},
							Type: &netbox.VirtualizationClusterType{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								ID: 1,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P3] ingest virtualization.cluster with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.cluster",
				"entity": {
					"Cluster": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: &netbox.VirtualizationCluster{
							ID:   1,
							Name: "Test",
						},
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P3] ingest empty virtualization.cluster - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.cluster",
				"entity": {
					"Cluster": {}
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
			name: "[P4] ingest virtualization.virtualmachine with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualmachine",
				"entity": {
					"VirtualMachine": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: nil,
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: nil,
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.cluster",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationCluster{
							Name: "undefined",
							Group: &netbox.VirtualizationClusterGroup{
								ID: 1,
							},
							Type: &netbox.VirtualizationClusterType{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								ID: 1,
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.virtualmachine",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationVirtualMachine{
							Name: "Test",
							Cluster: &netbox.VirtualizationCluster{
								Name: "undefined",
								Group: &netbox.VirtualizationClusterGroup{
									ID: 1,
								},
								Type: &netbox.VirtualizationClusterType{
									ID: 1,
								},
								Site: &netbox.DcimSite{
									ID: 1,
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
							Role: &netbox.DcimDeviceRole{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								ID: 1,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P4] ingest virtualization.virtualmachine with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualmachine",
				"entity": {
					"VirtualMachine": {
						"name": "Test",
						"cluster": {
							"name": "cluster1"
						}
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: &netbox.VirtualizationVirtualMachine{
							ID:   1,
							Name: "Test",
							Site: &netbox.DcimSite{
								ID:     1,
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							Cluster: &netbox.VirtualizationCluster{
								ID:   1,
								Name: "cluster1",
								Group: &netbox.VirtualizationClusterGroup{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
								Type: &netbox.VirtualizationClusterType{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "cluster1", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: &netbox.VirtualizationCluster{
							ID:   1,
							Name: "cluster1",
							Group: &netbox.VirtualizationClusterGroup{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
							Type: &netbox.VirtualizationClusterType{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P4] ingest empty virtualization.virtualmachine - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualmachine",
				"entity": {
					"VirtualMachine": {}
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
			name: "[P5] ingest virtualization.interface with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.interface",
				"entity": {
					"VirtualInterface": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "virtual_machine__name": "undefined", "virtual_machine__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationInterfaceDataWrapper{
						VirtualInterface: nil,
					},
				},
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: nil,
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: nil,
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.cluster",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationCluster{
							Name: "undefined",
							Group: &netbox.VirtualizationClusterGroup{
								ID: 1,
							},
							Type: &netbox.VirtualizationClusterType{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								ID: 1,
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.virtualmachine",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationVirtualMachine{
							Name: "undefined",
							Cluster: &netbox.VirtualizationCluster{
								Name: "undefined",
								Group: &netbox.VirtualizationClusterGroup{
									ID: 1,
								},
								Type: &netbox.VirtualizationClusterType{
									ID: 1,
								},
								Site: &netbox.DcimSite{
									ID: 1,
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
							Role: &netbox.DcimDeviceRole{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								ID: 1,
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.interface",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationInterface{
							Name: "Test",
							VirtualMachine: &netbox.VirtualizationVirtualMachine{
								Name: "undefined",
								Cluster: &netbox.VirtualizationCluster{
									Name: "undefined",
									Group: &netbox.VirtualizationClusterGroup{
										ID: 1,
									},
									Type: &netbox.VirtualizationClusterType{
										ID: 1,
									},
									Site: &netbox.DcimSite{
										ID: 1,
									},
									Status: strPtr(netbox.DefaultVirtualizationStatus),
								},
								Role: &netbox.DcimDeviceRole{
									ID: 1,
								},
								Site: &netbox.DcimSite{
									ID: 1,
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P5] ingest virtualization.interface with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.interface",
				"entity": {
					"VirtualInterface": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.interface",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "virtual_machine__name": "undefined", "virtual_machine__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationInterfaceDataWrapper{
						VirtualInterface: &netbox.VirtualizationInterface{
							ID:   1,
							Name: "Test",
							VirtualMachine: &netbox.VirtualizationVirtualMachine{
								ID:   1,
								Name: "undefined",
								Cluster: &netbox.VirtualizationCluster{
									ID:   1,
									Name: "undefined",
									Group: &netbox.VirtualizationClusterGroup{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
									Type: &netbox.VirtualizationClusterType{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
									Status: strPtr(netbox.DefaultVirtualizationStatus),
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
						},
					},
				},
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: &netbox.VirtualizationVirtualMachine{
							ID:   1,
							Name: "undefined",
							Cluster: &netbox.VirtualizationCluster{
								ID:   1,
								Name: "undefined",
								Group: &netbox.VirtualizationClusterGroup{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
								Type: &netbox.VirtualizationClusterType{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
						},
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: &netbox.VirtualizationCluster{
							ID:   1,
							Name: "undefined",
							Group: &netbox.VirtualizationClusterGroup{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
							Type: &netbox.VirtualizationClusterType{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
						},
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P5] ingest empty virtualization.interface - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.interface",
				"entity": {
					"VirtualInterface": {}
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
			name: "[P6] ingest virtualization.virtualdisk with name only - existing object not found - create",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualdisk",
				"entity": {
					"VirtualDisk": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.virtualdisk",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "virtual_machine__name": "undefined", "virtual_machine__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualDiskDataWrapper{
						VirtualDisk: nil,
					},
				},
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: nil,
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: nil,
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
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
						ObjectType:    "virtualization.cluster",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationCluster{
							Name: "undefined",
							Group: &netbox.VirtualizationClusterGroup{
								ID: 1,
							},
							Type: &netbox.VirtualizationClusterType{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.virtualmachine",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationVirtualMachine{
							Name: "undefined",
							Cluster: &netbox.VirtualizationCluster{
								Name: "undefined",
								Group: &netbox.VirtualizationClusterGroup{
									ID: 1,
								},
								Type: &netbox.VirtualizationClusterType{
									ID: 1,
								},
								Site: &netbox.DcimSite{
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
							Role: &netbox.DcimDeviceRole{
								ID: 1,
							},
							Site: &netbox.DcimSite{
								Name:   "undefined",
								Slug:   "undefined",
								Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "virtualization.virtualdisk",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.VirtualizationVirtualDisk{
							Name: "Test",
							VirtualMachine: &netbox.VirtualizationVirtualMachine{
								Name: "undefined",
								Cluster: &netbox.VirtualizationCluster{
									Name: "undefined",
									Group: &netbox.VirtualizationClusterGroup{
										ID: 1,
									},
									Type: &netbox.VirtualizationClusterType{
										ID: 1,
									},
									Site: &netbox.DcimSite{
										Name:   "undefined",
										Slug:   "undefined",
										Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
									},
									Status: strPtr(netbox.DefaultVirtualizationStatus),
								},
								Role: &netbox.DcimDeviceRole{
									ID: 1,
								},
								Site: &netbox.DcimSite{
									Name:   "undefined",
									Slug:   "undefined",
									Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest virtualization.virtualdisk with name only - existing object found - do nothing",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualdisk",
				"entity": {
					"VirtualDisk": {
						"name": "Test"
					}
				},
				"state": 0
			}`),
			retrieveObjectStates: []mockRetrieveObjectState{
				{
					objectType:     "virtualization.virtualdisk",
					objectID:       0,
					queryParams:    map[string]string{"q": "Test", "virtual_machine__name": "undefined", "virtual_machine__site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualDiskDataWrapper{
						VirtualDisk: &netbox.VirtualizationVirtualDisk{
							ID:   1,
							Name: "Test",
							VirtualMachine: &netbox.VirtualizationVirtualMachine{
								ID:   1,
								Name: "undefined",
								Cluster: &netbox.VirtualizationCluster{
									ID:   1,
									Name: "undefined",
									Group: &netbox.VirtualizationClusterGroup{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
									Type: &netbox.VirtualizationClusterType{
										ID:   1,
										Name: "undefined",
										Slug: "undefined",
									},
									Status: strPtr(netbox.DefaultVirtualizationStatus),
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
						},
					},
				},
				{
					objectType:     "virtualization.virtualmachine",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationVirtualMachineDataWrapper{
						VirtualMachine: &netbox.VirtualizationVirtualMachine{
							ID:   1,
							Name: "undefined",
							Cluster: &netbox.VirtualizationCluster{
								ID:   1,
								Name: "undefined",
								Group: &netbox.VirtualizationClusterGroup{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
								Type: &netbox.VirtualizationClusterType{
									ID:   1,
									Name: "undefined",
									Slug: "undefined",
								},
								Status: strPtr(netbox.DefaultVirtualizationStatus),
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
				},
				{
					objectType:     "virtualization.cluster",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined", "site__name": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterDataWrapper{
						Cluster: &netbox.VirtualizationCluster{
							ID:   1,
							Name: "undefined",
							Group: &netbox.VirtualizationClusterGroup{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
							Type: &netbox.VirtualizationClusterType{
								ID:   1,
								Name: "undefined",
								Slug: "undefined",
							},
							Status: strPtr(netbox.DefaultVirtualizationStatus),
						},
					},
				},
				{
					objectType:     "virtualization.clustergroup",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterGroupDataWrapper{
						ClusterGroup: &netbox.VirtualizationClusterGroup{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
						},
					},
				},
				{
					objectType:     "virtualization.clustertype",
					objectID:       0,
					queryParams:    map[string]string{"q": "undefined"},
					objectChangeID: 0,
					object: &netbox.VirtualizationClusterTypeDataWrapper{
						ClusterType: &netbox.VirtualizationClusterType{
							ID:   1,
							Name: "undefined",
							Slug: "undefined",
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
			},
			wantChangeSet: changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet:   []changeset.Change{},
			},
			wantErr: false,
		},
		{
			name: "[P6] ingest empty virtualization.virtualdisk - error",
			rawIngestEntity: []byte(`{
				"request_id": "cfa0f129-125c-440d-9e41-e87583cd7d89",
				"data_type": "virtualization.virtualdisk",
				"entity": {
					"VirtualDisk": {}
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
