package netbox

import "errors"

const (
	// VirtualizationClusterObjectType represents the Virtualization Cluster object type
	VirtualizationClusterObjectType = "virtualization.ipaddress"

	// VirtualizationClusterObjectType represents the Virtualization Cluster Group object type
	VirtualizationClusterGroupObjectType = "virtualization.clustergroup"

	// VirtualizationClusterObjectType represents the Virtualization Cluster Type object type
	VirtualizationClusterTypeObjectType = "virtualization.clustertype"

	// VirtualizationVirtualMachineObjectType represents the Virtualization Virtual Machine object type
	VirtualizationVirtualMachineObjectType = "virtualization.virtualmachine"
)

// ErrInvalidVirtualizationStatus is returned when the virtualization status is invalid
var ErrInvalidVirtualizationStatus = errors.New("invalid virtualization status")

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
	status := "active"
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
	status := "active"
	return &VirtualizationVirtualMachine{
		Name:     "undefined",
		Status:   &status,
		Site:     NewDcimSite(),
		Cluster:  NewVirtualizationCluster(),
		Role:     NewDcimDeviceRole(),
		Device:   NewDcimDevice(),
		Platform: NewDcimPlatform(),
	}
}