package netbox

import (
	"errors"
	"fmt"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

const (
	// VirtualizationClusterObjectType represents the Virtualization Cluster object type
	VirtualizationClusterObjectType = "virtualization.cluster"

	// VirtualizationClusterGroupObjectType represents the Virtualization Cluster Group object type
	VirtualizationClusterGroupObjectType = "virtualization.clustergroup"

	// VirtualizationClusterTypeObjectType represents the Virtualization Cluster Type object type
	VirtualizationClusterTypeObjectType = "virtualization.clustertype"

	// VirtualizationVirtualMachineObjectType represents the Virtualization Virtual Machine object type
	VirtualizationVirtualMachineObjectType = "virtualization.virtualmachine"

	// VirtualizationVMInterfaceObjectType represents the Virtualization Interface object type
	VirtualizationVMInterfaceObjectType = "virtualization.vminterface"

	// VirtualizationVirtualDiskObjectType represents the Virtualization Virtual Disk object type
	VirtualizationVirtualDiskObjectType = "virtualization.virtualdisk"
)

var (
	// ErrInvalidVirtualizationStatus is returned when the virtualization status is invalid
	ErrInvalidVirtualizationStatus = errors.New("invalid virtualization status")

	// DefaultVirtualizationStatus is the default status for Virtualization objects
	DefaultVirtualizationStatus = "active"
)

var virtualizationStatusMap = map[string]struct{}{
	"offline":         {},
	"active":          {},
	"planned":         {},
	"staged":          {},
	"failed":          {},
	"decommissioning": {},
}

func validateVirtualizationStatus(s string) bool {
	_, ok := virtualizationStatusMap[s]
	return ok
}

// VirtualizationClusterGroup represents a Virtualization Cluster Group
type VirtualizationClusterGroup struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        []*Tag  `json:"tags,omitempty"`
}

// VirtualizationClusterType represents a Virtualization Cluster Type
type VirtualizationClusterType struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        []*Tag  `json:"tags,omitempty"`
}

// VirtualizationCluster represents a Virtualization Cluster
type VirtualizationCluster struct {
	ID          int                         `json:"id,omitempty"`
	Name        string                      `json:"name,omitempty"`
	Type        *VirtualizationClusterType  `json:"type,omitempty" mapstructure:"type"`
	Group       *VirtualizationClusterGroup `json:"group,omitempty" mapstructure:"group"`
	Site        *DcimSite                   `json:"site,omitempty"`
	Status      *string                     `json:"status,omitempty"`
	Description *string                     `json:"description,omitempty"`
	Tags        []*Tag                      `json:"tags,omitempty"`
}

// Validate checks if the Virtualization Cluster is valid
func (cluster *VirtualizationCluster) Validate() error {
	if cluster.Status != nil && !validateVirtualizationStatus(*cluster.Status) {
		return ErrInvalidVirtualizationStatus
	}
	return nil
}

// VirtualizationVirtualMachine represents a Virtualization Virtual Machine
type VirtualizationVirtualMachine struct {
	ID          int                    `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Status      *string                `json:"status,omitempty"`
	Site        *DcimSite              `json:"site,omitempty"`
	Cluster     *VirtualizationCluster `json:"cluster,omitempty"`
	Role        *DcimDeviceRole        `json:"role,omitempty" mapstructure:"role"`
	Device      *DcimDevice            `json:"device,omitempty"`
	Platform    *DcimPlatform          `json:"platform,omitempty"`
	PrimaryIPv4 *IpamIPAddress         `json:"primary_ip4,omitempty" mapstructure:"primary_ip4"`
	PrimaryIPv6 *IpamIPAddress         `json:"primary_ip6,omitempty" mapstructure:"primary_ip6"`
	Vcpus       *int                   `json:"vcpus,omitempty"`
	Memory      *int                   `json:"memory,omitempty"`
	Disk        *int                   `json:"disk,omitempty"`
	Description *string                `json:"description,omitempty"`
	Comments    *string                `json:"comments,omitempty"`
	Tags        []*Tag                 `json:"tags,omitempty"`
}

// VirtualizationVMInterface represents a Virtualization Interface
type VirtualizationVMInterface struct {
	ID             int                           `json:"id,omitempty"`
	VirtualMachine *VirtualizationVirtualMachine `json:"virtual_machine,omitempty"  mapstructure:"virtual_machine"`
	Name           string                        `json:"name,omitempty"`
	Enabled        *bool                         `json:"enabled,omitempty"`
	MTU            *int                          `json:"mtu,omitempty"`
	MACAddress     *string                       `json:"mac_address,omitempty" mapstructure:"mac_address,omitempty"`
	Description    *string                       `json:"description,omitempty"`
	Tags           []*Tag                        `json:"tags,omitempty"`
}

// VirtualizationVirtualDisk represents a Virtualization Virtual Disk
type VirtualizationVirtualDisk struct {
	ID             int                           `json:"id,omitempty"`
	VirtualMachine *VirtualizationVirtualMachine `json:"virtual_machine,omitempty"  mapstructure:"virtual_machine"`
	Name           string                        `json:"name,omitempty"`
	Size           int                           `json:"size,omitempty"`
	Description    *string                       `json:"description,omitempty"`
	Tags           []*Tag                        `json:"tags,omitempty"`
}

// Validate checks if the Virtualization Virtual Machine is valid
func (vm *VirtualizationVirtualMachine) Validate() error {
	if vm.Status != nil && !validateVirtualizationStatus(*vm.Status) {
		return ErrInvalidVirtualizationStatus
	}
	return nil
}

// NewVirtualizationClusterGroup creates a new virtualization cluster group placeholder
func NewVirtualizationClusterGroup() *VirtualizationClusterGroup {
	return &VirtualizationClusterGroup{
		Name: "undefined",
		Slug: "undefined",
	}
}

// NewVirtualizationClusterType creates a new virtualization cluster type placeholder
func NewVirtualizationClusterType() *VirtualizationClusterType {
	return &VirtualizationClusterType{
		Name: "undefined",
		Slug: "undefined",
	}
}

// NewVirtualizationCluster creates a new virtualization cluster placeholder
func NewVirtualizationCluster() *VirtualizationCluster {
	status := DefaultVirtualizationStatus
	return &VirtualizationCluster{
		Name:   "undefined",
		Type:   NewVirtualizationClusterType(),
		Group:  NewVirtualizationClusterGroup(),
		Site:   NewDcimSite(),
		Status: &status,
	}
}

// NewVirtualizationVirtualMachine creates a new virtualization virtual machine placeholder
func NewVirtualizationVirtualMachine() *VirtualizationVirtualMachine {
	status := DefaultVirtualizationStatus
	return &VirtualizationVirtualMachine{
		Name:   "undefined",
		Status: &status,
		Site:   NewDcimSite(),
		Role:   NewDcimDeviceRole(),
	}
}

// NewVirtualizationVMInterface creates a new virtualization VM interface placeholder
func NewVirtualizationVMInterface() *VirtualizationVMInterface {
	return &VirtualizationVMInterface{
		Name:           "undefined",
		VirtualMachine: NewVirtualizationVirtualMachine(),
	}
}

// NewVirtualizationVirtualDisk creates a new virtualization virtual disk placeholder
func NewVirtualizationVirtualDisk() *VirtualizationVirtualDisk {
	return &VirtualizationVirtualDisk{
		Name:           "undefined",
		VirtualMachine: NewVirtualizationVirtualMachine(),
		Size:           0,
	}
}

// FromProtoClusterEntity converts a diode cluster entity to a cluster
func FromProtoClusterEntity(entity *diodepb.Entity) (*VirtualizationCluster, error) {
	if entity == nil || entity.GetCluster() == nil {
		return nil, fmt.Errorf("entity is nil or not a cluster")
	}

	return FromProtoCluster(entity.GetCluster()), nil
}

// FromProtoCluster converts a diode cluster to a virtualization cluster
func FromProtoCluster(clusterPb *diodepb.Cluster) *VirtualizationCluster {
	if clusterPb == nil {
		return nil
	}

	var status *string
	if clusterPb.Status != "" {
		status = &clusterPb.Status
	}

	return &VirtualizationCluster{
		Name:        clusterPb.Name,
		Type:        FromProtoClusterType(clusterPb.Type),
		Group:       FromProtoClusterGroup(clusterPb.Group),
		Site:        FromProtoSite(clusterPb.Site),
		Status:      status,
		Description: clusterPb.Description,
		Tags:        FromProtoTags(clusterPb.Tags),
	}
}

// FromProtoClusterTypeEntity converts a diode cluster type entity to a cluster type
func FromProtoClusterTypeEntity(entity *diodepb.Entity) (*VirtualizationClusterType, error) {
	if entity == nil || entity.GetClusterType() == nil {
		return nil, fmt.Errorf("entity is nil or not a cluster type")
	}

	return FromProtoClusterType(entity.GetClusterType()), nil
}

// FromProtoClusterType converts a diode cluster type to a cluster type
func FromProtoClusterType(clusterTypePb *diodepb.ClusterType) *VirtualizationClusterType {
	if clusterTypePb == nil {
		return nil
	}

	return &VirtualizationClusterType{
		Name:        clusterTypePb.Name,
		Slug:        clusterTypePb.Slug,
		Description: clusterTypePb.Description,
		Tags:        FromProtoTags(clusterTypePb.Tags),
	}
}

// FromProtoClusterGroupEntity converts a diode cluster group entity to a cluster group
func FromProtoClusterGroupEntity(entity *diodepb.Entity) (*VirtualizationClusterGroup, error) {
	if entity == nil || entity.GetClusterGroup() == nil {
		return nil, fmt.Errorf("entity is nil or not a cluster group")
	}

	return FromProtoClusterGroup(entity.GetClusterGroup()), nil
}

// FromProtoClusterGroup converts a diode cluster group to a cluster group
func FromProtoClusterGroup(clusterGroupPb *diodepb.ClusterGroup) *VirtualizationClusterGroup {
	if clusterGroupPb == nil {
		return nil
	}

	return &VirtualizationClusterGroup{
		Name:        clusterGroupPb.Name,
		Slug:        clusterGroupPb.Slug,
		Description: clusterGroupPb.Description,
		Tags:        FromProtoTags(clusterGroupPb.Tags),
	}
}

// FromProtoVirtualMachineEntity converts a diode virtual machine entity to a virtual machine
func FromProtoVirtualMachineEntity(entity *diodepb.Entity) (*VirtualizationVirtualMachine, error) {
	if entity == nil || entity.GetVirtualMachine() == nil {
		return nil, fmt.Errorf("entity is nil or not a virtual machine")
	}

	return FromProtoVirtualMachine(entity.GetVirtualMachine()), nil
}

// FromProtoVirtualMachine converts a diode virtual machine to a virtual machine
func FromProtoVirtualMachine(virtualMachinePb *diodepb.VirtualMachine) *VirtualizationVirtualMachine {
	if virtualMachinePb == nil {
		return nil
	}

	var status *string
	if virtualMachinePb.Status != "" {
		status = &virtualMachinePb.Status
	}

	return &VirtualizationVirtualMachine{
		Name:        virtualMachinePb.Name,
		Status:      status,
		Site:        FromProtoSite(virtualMachinePb.Site),
		Cluster:     FromProtoCluster(virtualMachinePb.Cluster),
		Role:        FromProtoRole(virtualMachinePb.Role),
		Device:      FromProtoDevice(virtualMachinePb.Device),
		Platform:    FromProtoPlatform(virtualMachinePb.Platform),
		PrimaryIPv4: FromProtoIPAddress(virtualMachinePb.PrimaryIp4),
		PrimaryIPv6: FromProtoIPAddress(virtualMachinePb.PrimaryIp6),
		Vcpus:       int32PtrToIntPtr(virtualMachinePb.Vcpus),
		Memory:      int32PtrToIntPtr(virtualMachinePb.Memory),
		Disk:        int32PtrToIntPtr(virtualMachinePb.Disk),
		Description: virtualMachinePb.Description,
		Comments:    virtualMachinePb.Comments,
		Tags:        FromProtoTags(virtualMachinePb.Tags),
	}
}

// FromProtoVMInterfaceEntity converts a diode virtual machine interface entity to a virtual machine interface
func FromProtoVMInterfaceEntity(entity *diodepb.Entity) (*VirtualizationVMInterface, error) {
	if entity == nil || entity.GetVminterface() == nil {
		return nil, fmt.Errorf("entity is nil or not a virtual machine interface")
	}

	return FromProtoVMInterface(entity.GetVminterface()), nil
}

// FromProtoVMInterface converts a diode virtual machine interface to a virtual machine interface
func FromProtoVMInterface(vmInterfacePb *diodepb.VMInterface) *VirtualizationVMInterface {
	if vmInterfacePb == nil {
		return nil
	}

	return &VirtualizationVMInterface{
		VirtualMachine: FromProtoVirtualMachine(vmInterfacePb.VirtualMachine),
		Name:           vmInterfacePb.Name,
		Enabled:        vmInterfacePb.Enabled,
		MTU:            int32PtrToIntPtr(vmInterfacePb.Mtu),
		MACAddress:     vmInterfacePb.MacAddress,
		Description:    vmInterfacePb.Description,
		Tags:           FromProtoTags(vmInterfacePb.Tags),
	}
}

// FromProtoVirtualDiskEntity converts a diode virtual disk entity to a virtual disk
func FromProtoVirtualDiskEntity(entity *diodepb.Entity) (*VirtualizationVirtualDisk, error) {
	if entity == nil || entity.GetVirtualDisk() == nil {
		return nil, fmt.Errorf("entity is nil or not a virtual disk")
	}

	return FromProtoVirtualDisk(entity.GetVirtualDisk()), nil
}

// FromProtoVirtualDisk converts a diode virtual disk to a virtual disk
func FromProtoVirtualDisk(virtualDiskPb *diodepb.VirtualDisk) *VirtualizationVirtualDisk {
	if virtualDiskPb == nil {
		return nil
	}

	return &VirtualizationVirtualDisk{
		VirtualMachine: FromProtoVirtualMachine(virtualDiskPb.VirtualMachine),
		Name:           virtualDiskPb.Name,
		Size:           int(virtualDiskPb.Size),
		Description:    virtualDiskPb.Description,
		Tags:           FromProtoTags(virtualDiskPb.Tags),
	}
}
