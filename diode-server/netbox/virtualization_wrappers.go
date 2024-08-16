package netbox

import (
	"errors"

	"github.com/gosimple/slug"
	"github.com/mitchellh/hashstructure/v2"
)

// VirtualizationClusterGroupDataWrapper represents a virtualization cluster group data wrapper
type VirtualizationClusterGroupDataWrapper struct {
	ClusterGroup *VirtualizationClusterGroup

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*VirtualizationClusterGroupDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationClusterGroupDataWrapper) Data() any {
	return vw.ClusterGroup
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationClusterGroupDataWrapper) IsValid() bool {
	if vw.ClusterGroup != nil && !vw.hasParent && vw.ClusterGroup.Name == "" {
		vw.ClusterGroup = nil
	}
	return vw.ClusterGroup != nil
}

// Normalise normalises the data
func (vw *VirtualizationClusterGroupDataWrapper) Normalise() {
	if vw.IsValid() && vw.ClusterGroup.Tags != nil && len(vw.ClusterGroup.Tags) == 0 {
		vw.ClusterGroup.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationClusterGroupDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.ClusterGroup != nil && vw.hasParent && vw.ClusterGroup.Name == "" {
		vw.ClusterGroup = nil
	}

	objects := make([]ComparableData, 0)

	if vw.ClusterGroup == nil && vw.intended {
		return objects, nil
	}

	if vw.ClusterGroup == nil && vw.hasParent {
		vw.ClusterGroup = NewVirtualizationClusterGroup()
		vw.placeholder = true
	}

	if vw.ClusterGroup.Slug == "" {
		vw.ClusterGroup.Slug = slug.Make(vw.ClusterGroup.Name)
	}

	if vw.ClusterGroup.Tags != nil {
		for _, t := range vw.ClusterGroup.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	vw.nestedObjects = objects

	objects = append(objects, vw)

	return objects, nil
}

// DataType returns the data type
func (vw *VirtualizationClusterGroupDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationClusterGroupDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.ClusterGroup.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationClusterGroupDataWrapper) ID() int {
	return vw.ClusterGroup.ID
}

// HasChanged returns true if the data has changed
func (vw *VirtualizationClusterGroupDataWrapper) HasChanged() bool {
	return vw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (vw *VirtualizationClusterGroupDataWrapper) IsPlaceholder() bool {
	return vw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationClusterGroupDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationClusterGroupDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		vw.ClusterGroup.ID = intended.ClusterGroup.ID
		vw.ClusterGroup.Name = intended.ClusterGroup.Name
		vw.ClusterGroup.Slug = intended.ClusterGroup.Slug

		if vw.ClusterGroup.Description == nil {
			vw.ClusterGroup.Description = intended.ClusterGroup.Description
		}

		tagsToMerge := mergeTags(vw.ClusterGroup.Tags, intended.ClusterGroup.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.ClusterGroup.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(vw.ClusterGroup.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.ClusterGroup.Tags = tagsToMerge
		}
	}

	for _, t := range vw.ClusterGroup.Tags {
		if t.ID == 0 {
			vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationClusterGroupDataWrapper) SetDefaults() {}

// VirtualizationClusterTypeDataWrapper represents a virtualization cluster type data wrapper
type VirtualizationClusterTypeDataWrapper struct {
	ClusterType *VirtualizationClusterType

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*VirtualizationClusterTypeDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationClusterTypeDataWrapper) Data() any {
	return vw.ClusterType
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationClusterTypeDataWrapper) IsValid() bool {
	if vw.ClusterType != nil && !vw.hasParent && vw.ClusterType.Name == "" {
		vw.ClusterType = nil
	}
	return vw.ClusterType != nil
}

// Normalise normalises the data
func (vw *VirtualizationClusterTypeDataWrapper) Normalise() {
	if vw.IsValid() && vw.ClusterType.Tags != nil && len(vw.ClusterType.Tags) == 0 {
		vw.ClusterType.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationClusterTypeDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.ClusterType != nil && vw.hasParent && vw.ClusterType.Name == "" {
		vw.ClusterType = nil
	}

	objects := make([]ComparableData, 0)

	if vw.ClusterType == nil && vw.intended {
		return objects, nil
	}

	if vw.ClusterType == nil && vw.hasParent {
		vw.ClusterType = NewVirtualizationClusterType()
		vw.placeholder = true
	}

	if vw.ClusterType.Slug == "" {
		vw.ClusterType.Slug = slug.Make(vw.ClusterType.Name)
	}

	if vw.ClusterType.Tags != nil {
		for _, t := range vw.ClusterType.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	vw.nestedObjects = objects

	objects = append(objects, vw)

	return objects, nil
}

// DataType returns the data type
func (vw *VirtualizationClusterTypeDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationClusterTypeDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.ClusterType.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationClusterTypeDataWrapper) ID() int {
	return vw.ClusterType.ID
}

// HasChanged returns true if the data has changed
func (vw *VirtualizationClusterTypeDataWrapper) HasChanged() bool {
	return vw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (vw *VirtualizationClusterTypeDataWrapper) IsPlaceholder() bool {
	return vw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationClusterTypeDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationClusterTypeDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		vw.ClusterType.ID = intended.ClusterType.ID
		vw.ClusterType.Name = intended.ClusterType.Name
		vw.ClusterType.Slug = intended.ClusterType.Slug

		if vw.ClusterType.Description == nil {
			vw.ClusterType.Description = intended.ClusterType.Description
		}

		tagsToMerge := mergeTags(vw.ClusterType.Tags, intended.ClusterType.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.ClusterType.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(vw.ClusterType.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.ClusterType.Tags = tagsToMerge
		}
	}

	for _, t := range vw.ClusterType.Tags {
		if t.ID == 0 {
			vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationClusterTypeDataWrapper) SetDefaults() {}

// VirtualizationClusterDataWrapper represents a virtualization cluster data wrapper
type VirtualizationClusterDataWrapper struct {
	Cluster *VirtualizationCluster

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*VirtualizationClusterDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationClusterDataWrapper) Data() any {
	return vw.Cluster
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationClusterDataWrapper) IsValid() bool {
	if vw.Cluster != nil && !vw.hasParent && vw.Cluster.Name == "" {
		vw.Cluster = nil
	}
	return vw.Cluster != nil
}

// Normalise normalises the data
func (vw *VirtualizationClusterDataWrapper) Normalise() {
	if vw.IsValid() && vw.Cluster.Tags != nil && len(vw.Cluster.Tags) == 0 {
		vw.Cluster.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationClusterDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.Cluster != nil && vw.hasParent && vw.Cluster.Name == "" {
		vw.Cluster = nil
	}

	objects := make([]ComparableData, 0)

	if vw.Cluster == nil && vw.intended {
		return objects, nil
	}

	if vw.Cluster == nil && vw.hasParent {
		vw.Cluster = NewVirtualizationCluster()
		vw.placeholder = true
	}

	clusterGroup := VirtualizationClusterGroupDataWrapper{ClusterGroup: vw.Cluster.Group, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	cgo, err := clusterGroup.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, cgo...)

	vw.Cluster.Group = clusterGroup.ClusterGroup

	clusterType := VirtualizationClusterTypeDataWrapper{ClusterType: vw.Cluster.Type, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	cto, err := clusterGroup.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, cto...)

	vw.Cluster.Type = clusterType.ClusterType

	site := DcimSiteDataWrapper{Site: vw.Cluster.Site, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	vw.Cluster.Site = site.Site

	if vw.Cluster.Tags != nil {
		for _, t := range vw.Cluster.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	vw.nestedObjects = objects

	objects = append(objects, vw)

	return objects, nil
}

// DataType returns the data type
func (vw *VirtualizationClusterDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationClusterDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.Cluster.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationClusterDataWrapper) ID() int {
	return vw.Cluster.ID
}

// HasChanged returns true if the data has changed
func (vw *VirtualizationClusterDataWrapper) HasChanged() bool {
	return vw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (vw *VirtualizationClusterDataWrapper) IsPlaceholder() bool {
	return vw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationClusterDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimManufacturerDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		vw.Cluster.ID = intended.Manufacturer.ID
		vw.Cluster.Name = intended.Manufacturer.Name

		if vw.Cluster.Description == nil {
			vw.Cluster.Description = intended.Manufacturer.Description
		}

		tagsToMerge := mergeTags(vw.Cluster.Tags, intended.Manufacturer.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.Cluster.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(vw.Cluster.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.Cluster.Tags = tagsToMerge
		}
	}

	for _, t := range vw.Cluster.Tags {
		if t.ID == 0 {
			vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationClusterDataWrapper) SetDefaults() {}

// VirtualizationVirtualMachineDataWrapper represents a virtualization virtual machine data wrapper
type VirtualizationVirtualMachineDataWrapper struct {
	VirtualMachine *VirtualizationVirtualMachine

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*VirtualizationVirtualMachineDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationVirtualMachineDataWrapper) Data() any {
	return vw.VirtualMachine
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationVirtualMachineDataWrapper) IsValid() bool {
	if vw.VirtualMachine != nil && !vw.hasParent && vw.VirtualMachine.Name == "" {
		vw.VirtualMachine = nil
	}
	return vw.VirtualMachine != nil
}

// Normalise normalises the data
func (vw *VirtualizationVirtualMachineDataWrapper) Normalise() {
	if vw.IsValid() && vw.VirtualMachine.Tags != nil && len(vw.VirtualMachine.Tags) == 0 {
		vw.VirtualMachine.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationVirtualMachineDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.VirtualMachine != nil && vw.hasParent && vw.VirtualMachine.Name == "" {
		vw.VirtualMachine = nil
	}

	objects := make([]ComparableData, 0)

	if vw.VirtualMachine == nil && vw.intended {
		return objects, nil
	}

	if vw.VirtualMachine == nil && vw.hasParent {
		vw.VirtualMachine = NewVirtualizationVirtualMachine()
		vw.placeholder = true
	}

	// Ignore primary IP addresses for time being
	vw.VirtualMachine.PrimaryIPv4 = nil
	vw.VirtualMachine.PrimaryIPv6 = nil

	cluster := VirtualizationClusterDataWrapper{Cluster: vw.VirtualMachine.Cluster, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	co, err := cluster.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, co...)

	vw.VirtualMachine.Cluster = cluster.Cluster

	site := DcimSiteDataWrapper{Site: vw.VirtualMachine.Site, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	vw.VirtualMachine.Site = site.Site

	if vw.VirtualMachine.Platform != nil {
		platform := DcimPlatformDataWrapper{Platform: vw.VirtualMachine.Platform, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

		po, err := platform.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, po...)

		vw.VirtualMachine.Platform = platform.Platform
	}

	deviceRole := DcimDeviceRoleDataWrapper{DeviceRole: vw.VirtualMachine.Role, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	dro, err := deviceRole.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dro...)

	vw.VirtualMachine.Role = deviceRole.DeviceRole

	device := DcimDeviceDataWrapper{Device: vw.VirtualMachine.Device, placeholder: vw.placeholder, hasParent: true, intended: vw.intended}

	do, err := device.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, do...)

	vw.VirtualMachine.Device = device.Device

	if vw.VirtualMachine.Tags != nil {
		for _, t := range vw.VirtualMachine.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	vw.nestedObjects = objects

	objects = append(objects, vw)

	return objects, nil
}

// DataType returns the data type
func (vw *VirtualizationVirtualMachineDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationVirtualMachineDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.VirtualMachine.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationVirtualMachineDataWrapper) ID() int {
	return vw.VirtualMachine.ID
}

// HasChanged returns true if the data has changed
func (vw *VirtualizationVirtualMachineDataWrapper) HasChanged() bool {
	return vw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (vw *VirtualizationVirtualMachineDataWrapper) IsPlaceholder() bool {
	return vw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationVirtualMachineDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimManufacturerDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		vw.VirtualMachine.ID = intended.Manufacturer.ID
		vw.VirtualMachine.Name = intended.Manufacturer.Name

		if vw.VirtualMachine.Description == nil {
			vw.VirtualMachine.Description = intended.Manufacturer.Description
		}

		tagsToMerge := mergeTags(vw.VirtualMachine.Tags, intended.Manufacturer.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualMachine.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(vw.VirtualMachine.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualMachine.Tags = tagsToMerge
		}
	}

	for _, t := range vw.VirtualMachine.Tags {
		if t.ID == 0 {
			vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationVirtualMachineDataWrapper) SetDefaults() {}
