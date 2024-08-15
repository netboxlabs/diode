package netbox

import "github.com/gosimple/slug"

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
	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationVirtualMachineDataWrapper) SetDefaults() {}
