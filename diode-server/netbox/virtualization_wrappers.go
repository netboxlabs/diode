package netbox

import (
	"errors"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/mitchellh/hashstructure/v2"
)

// VirtualizationClusterGroupDataWrapper represents a virtualization cluster group data wrapper
type VirtualizationClusterGroupDataWrapper struct {
	BaseDataWrapper
	ClusterGroup *VirtualizationClusterGroup
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
	return VirtualizationClusterGroupObjectType
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
	BaseDataWrapper
	ClusterType *VirtualizationClusterType
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
	return VirtualizationClusterTypeObjectType
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
	BaseDataWrapper
	Cluster *VirtualizationCluster
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

	clusterGroup := VirtualizationClusterGroupDataWrapper{ClusterGroup: vw.Cluster.Group, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	cgo, err := clusterGroup.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, cgo...)

	vw.Cluster.Group = clusterGroup.ClusterGroup

	clusterType := VirtualizationClusterTypeDataWrapper{ClusterType: vw.Cluster.Type, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	cto, err := clusterType.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, cto...)

	vw.Cluster.Type = clusterType.ClusterType

	site := DcimSiteDataWrapper{Site: vw.Cluster.Site, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

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
	return VirtualizationClusterObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationClusterDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": vw.Cluster.Name,
	}
	if vw.Cluster.Site != nil {
		params["site__name"] = vw.Cluster.Site.Name
	}
	return params
}

// ID returns the ID of the data
func (vw *VirtualizationClusterDataWrapper) ID() int {
	return vw.Cluster.ID
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationClusterDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationClusterDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range vw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.Cluster.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.Cluster.Site))

	actualType := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.Cluster.Type))
	intendedType := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.Cluster.Type))

	actualGroup := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.Cluster.Group))
	intendedGroup := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.Cluster.Group))

	reconciliationRequired := true

	if intended != nil {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		vw.Cluster.ID = intended.Cluster.ID
		vw.Cluster.Name = intended.Cluster.Name

		if vw.Cluster.Description == nil {
			vw.Cluster.Description = intended.Cluster.Description
		}

		if actualSite.IsPlaceholder() && intended.Cluster.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Cluster.Site))
		}

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}

			intendedSiteID := intendedSite.ID()
			if intended.Cluster.Site != nil {
				intendedSiteID = intended.Cluster.Site.ID
			}

			intended.Cluster.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		vw.Cluster.Site = site

		vw.objectsToReconcile = append(vw.objectsToReconcile, siteObjectsToReconcile...)

		if actualType.IsPlaceholder() && intended.Cluster.Type != nil {
			intendedType = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Cluster.Type))
		}

		typeObjectsToReconcile, typeErr := actualType.Patch(intendedType, intendedNestedObjects)
		if typeErr != nil {
			return nil, typeErr
		}

		vType, err := copyData(actualType.Data().(*VirtualizationClusterType))
		if err != nil {
			return nil, err
		}
		vType.Tags = nil

		if !actualType.HasChanged() {
			vType = &VirtualizationClusterType{
				ID: actualType.ID(),
			}

			intendedTypeID := intendedType.ID()
			if intended.Cluster.Type != nil {
				intendedTypeID = intended.Cluster.Type.ID
			}

			intended.Cluster.Type = &VirtualizationClusterType{
				ID: intendedTypeID,
			}
		}

		vw.Cluster.Type = vType

		vw.objectsToReconcile = append(vw.objectsToReconcile, typeObjectsToReconcile...)

		if actualGroup.IsPlaceholder() && intended.Cluster.Group != nil {
			intendedGroup = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Cluster.Group))
		}

		groupObjectsToReconcile, groupErr := actualGroup.Patch(intendedGroup, intendedNestedObjects)
		if groupErr != nil {
			return nil, groupErr
		}

		group, err := copyData(actualGroup.Data().(*VirtualizationClusterGroup))
		if err != nil {
			return nil, err
		}
		group.Tags = nil

		if !actualGroup.HasChanged() {
			group = &VirtualizationClusterGroup{
				ID: actualGroup.ID(),
			}

			intendedGroupID := intendedGroup.ID()
			if intended.Cluster.Group != nil {
				intendedGroupID = intended.Cluster.Group.ID
			}

			intended.Cluster.Group = &VirtualizationClusterGroup{
				ID: intendedGroupID,
			}
		}

		vw.Cluster.Group = group

		vw.objectsToReconcile = append(vw.objectsToReconcile, groupObjectsToReconcile...)

		tagsToMerge := mergeTags(vw.Cluster.Tags, intended.Cluster.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.Cluster.Tags = tagsToMerge
		}

		for _, t := range vw.Cluster.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}
		}
		vw.Cluster.Site = site

		vw.objectsToReconcile = append(vw.objectsToReconcile, siteObjectsToReconcile...)

		typeObjectsToReconcile, typeErr := actualType.Patch(intendedType, intendedNestedObjects)
		if typeErr != nil {
			return nil, typeErr
		}

		vType, err := copyData(actualType.Data().(*VirtualizationClusterType))
		if err != nil {
			return nil, err
		}
		vType.Tags = nil

		if !actualType.HasChanged() {
			vType = &VirtualizationClusterType{
				ID: actualType.ID(),
			}
		}
		vw.Cluster.Type = vType

		vw.objectsToReconcile = append(vw.objectsToReconcile, typeObjectsToReconcile...)

		groupObjectsToReconcile, groupErr := actualGroup.Patch(intendedGroup, intendedNestedObjects)
		if groupErr != nil {
			return nil, groupErr
		}

		group, err := copyData(actualGroup.Data().(*VirtualizationClusterGroup))
		if err != nil {
			return nil, err
		}
		group.Tags = nil

		if !actualGroup.HasChanged() {
			group = &VirtualizationClusterGroup{
				ID: actualGroup.ID(),
			}
		}
		vw.Cluster.Group = group

		vw.objectsToReconcile = append(vw.objectsToReconcile, groupObjectsToReconcile...)

		tagsToMerge := mergeTags(vw.Cluster.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.Cluster.Tags = tagsToMerge
		}

		for _, t := range vw.Cluster.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
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
	BaseDataWrapper
	VirtualMachine *VirtualizationVirtualMachine
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

	cluster := VirtualizationClusterDataWrapper{Cluster: vw.VirtualMachine.Cluster, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	co, err := cluster.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, co...)

	vw.VirtualMachine.Cluster = cluster.Cluster

	site := DcimSiteDataWrapper{Site: vw.VirtualMachine.Site, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	vw.VirtualMachine.Site = site.Site

	if vw.VirtualMachine.Platform != nil {
		platform := DcimPlatformDataWrapper{Platform: vw.VirtualMachine.Platform, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

		po, err := platform.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, po...)

		vw.VirtualMachine.Platform = platform.Platform
	}

	deviceRole := DcimDeviceRoleDataWrapper{DeviceRole: vw.VirtualMachine.Role, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	dro, err := deviceRole.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dro...)

	vw.VirtualMachine.Role = deviceRole.DeviceRole

	if vw.VirtualMachine.Device != nil {
		device := DcimDeviceDataWrapper{Device: vw.VirtualMachine.Device, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

		do, err := device.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, do...)

		vw.VirtualMachine.Device = device.Device
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
	return VirtualizationVirtualMachineObjectType
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

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationVirtualMachineDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationVirtualMachineDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range vw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualMachine.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualMachine.Site))

	actualCluster := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualMachine.Cluster))
	intendedCluster := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualMachine.Cluster))

	actualRole := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualMachine.Role))
	intendedRole := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualMachine.Role))

	actualDevice := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualMachine.Device))
	intendedDevice := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualMachine.Device))

	actualPlatform := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualMachine.Platform))
	intendedPlatform := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualMachine.Platform))

	reconciliationRequired := true

	if intended != nil {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		vw.VirtualMachine.ID = intended.VirtualMachine.ID
		vw.VirtualMachine.Name = intended.VirtualMachine.Name

		if vw.VirtualMachine.Status == nil {
			vw.VirtualMachine.Status = intended.VirtualMachine.Status
		}

		if actualSite.IsPlaceholder() && intended.VirtualMachine.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualMachine.Site))
		}

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}

			intendedSiteID := intendedSite.ID()
			if intended.VirtualMachine.Site != nil {
				intendedSiteID = intended.VirtualMachine.Site.ID
			}

			intended.VirtualMachine.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		vw.VirtualMachine.Site = site

		vw.objectsToReconcile = append(vw.objectsToReconcile, siteObjectsToReconcile...)

		if actualCluster.IsPlaceholder() && intended.VirtualMachine.Cluster != nil {
			intendedCluster = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualMachine.Cluster))
		}

		clusterObjectsToReconcile, clusterErr := actualCluster.Patch(intendedCluster, intendedNestedObjects)
		if clusterErr != nil {
			return nil, clusterErr
		}

		cluster, err := copyData(actualCluster.Data().(*VirtualizationCluster))
		if err != nil {
			return nil, err
		}
		cluster.Tags = nil

		if !actualCluster.HasChanged() {
			cluster = &VirtualizationCluster{
				ID: actualCluster.ID(),
			}

			intendedClusterID := intendedCluster.ID()
			if intended.VirtualMachine.Cluster != nil {
				intendedClusterID = intended.VirtualMachine.Cluster.ID
			}

			intended.VirtualMachine.Cluster = &VirtualizationCluster{
				ID: intendedClusterID,
			}
		}

		vw.VirtualMachine.Cluster = cluster

		vw.objectsToReconcile = append(vw.objectsToReconcile, clusterObjectsToReconcile...)

		if actualRole.IsPlaceholder() && intended.VirtualMachine.Role != nil {
			intendedRole = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualMachine.Role))
		}

		roleObjectsToReconcile, roleErr := actualRole.Patch(intendedRole, intendedNestedObjects)
		if roleErr != nil {
			return nil, roleErr
		}

		role, err := copyData(actualRole.Data().(*DcimDeviceRole))
		if err != nil {
			return nil, err
		}
		role.Tags = nil

		if !actualRole.HasChanged() {
			role = &DcimDeviceRole{
				ID: actualRole.ID(),
			}

			intendedRoleID := intendedRole.ID()
			if intended.VirtualMachine.Role != nil {
				intendedRoleID = intended.VirtualMachine.Role.ID
			}

			intended.VirtualMachine.Role = &DcimDeviceRole{
				ID: intendedRoleID,
			}
		}

		vw.VirtualMachine.Role = role

		vw.objectsToReconcile = append(vw.objectsToReconcile, roleObjectsToReconcile...)

		if actualDevice != nil {
			if actualDevice.IsPlaceholder() && intended.VirtualMachine.Device != nil {
				intendedDevice = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualMachine.Device))
			}

			deviceObjectsToReconcile, deviceErr := actualDevice.Patch(intendedDevice, intendedNestedObjects)
			if deviceErr != nil {
				return nil, deviceErr
			}

			device, err := copyData(actualDevice.Data().(*DcimDevice))
			if err != nil {
				return nil, err
			}
			device.Tags = nil

			if !actualDevice.HasChanged() {
				device = &DcimDevice{
					ID: actualDevice.ID(),
				}

				intendedDeviceID := intendedDevice.ID()
				if intended.VirtualMachine.Device != nil {
					intendedDeviceID = intended.VirtualMachine.Device.ID
				}

				intended.VirtualMachine.Device = &DcimDevice{
					ID: intendedDeviceID,
				}
			}

			vw.VirtualMachine.Device = device

			vw.objectsToReconcile = append(vw.objectsToReconcile, deviceObjectsToReconcile...)
		}

		if actualPlatform != nil {
			if actualPlatform.IsPlaceholder() && intended.VirtualMachine.Platform != nil {
				intendedPlatform = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualMachine.Platform))
			}

			platformObjectsToReconcile, platformErr := actualPlatform.Patch(intendedPlatform, intendedNestedObjects)
			if platformErr != nil {
				return nil, platformErr
			}

			platform, err := copyData(actualPlatform.Data().(*DcimPlatform))
			if err != nil {
				return nil, err
			}
			platform.Tags = nil

			if !actualPlatform.HasChanged() {
				platform = &DcimPlatform{
					ID: actualPlatform.ID(),
				}

				intendedPlatformID := intendedPlatform.ID()
				if intended.VirtualMachine.Platform != nil {
					intendedPlatformID = intended.VirtualMachine.Platform.ID
				}

				intended.VirtualMachine.Platform = &DcimPlatform{
					ID: intendedPlatformID,
				}
			}

			vw.VirtualMachine.Platform = platform

			vw.objectsToReconcile = append(vw.objectsToReconcile, platformObjectsToReconcile...)
		} else {
			if intended.VirtualMachine.Platform != nil {
				platformID := intended.VirtualMachine.Platform.ID
				vw.VirtualMachine.Platform = &DcimPlatform{
					ID: platformID,
				}
				intended.VirtualMachine.Platform = &DcimPlatform{
					ID: platformID,
				}
			}
		}

		if vw.VirtualMachine.Vcpus == nil {
			vw.VirtualMachine.Vcpus = intended.VirtualMachine.Vcpus
		}
		if vw.VirtualMachine.Memory == nil {
			vw.VirtualMachine.Memory = intended.VirtualMachine.Memory
		}
		if vw.VirtualMachine.Disk == nil {
			vw.VirtualMachine.Disk = intended.VirtualMachine.Disk
		}
		if vw.VirtualMachine.Comments == nil {
			vw.VirtualMachine.Comments = intended.VirtualMachine.Comments
		}

		if vw.VirtualMachine.Description == nil {
			vw.VirtualMachine.Description = intended.VirtualMachine.Description
		}

		tagsToMerge := mergeTags(vw.VirtualMachine.Tags, intended.VirtualMachine.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualMachine.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualMachine.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}
		}
		vw.VirtualMachine.Site = site

		vw.objectsToReconcile = append(vw.objectsToReconcile, siteObjectsToReconcile...)

		clusterObjectsToReconcile, clusterErr := actualCluster.Patch(intendedCluster, intendedNestedObjects)
		if clusterErr != nil {
			return nil, clusterErr
		}

		cluster, err := copyData(actualCluster.Data().(*VirtualizationCluster))
		if err != nil {
			return nil, err
		}
		cluster.Tags = nil

		if !actualCluster.HasChanged() {
			cluster = &VirtualizationCluster{
				ID: actualCluster.ID(),
			}
		}
		vw.VirtualMachine.Cluster = cluster

		vw.objectsToReconcile = append(vw.objectsToReconcile, clusterObjectsToReconcile...)

		roleObjectsToReconcile, roleErr := actualRole.Patch(intendedRole, intendedNestedObjects)
		if roleErr != nil {
			return nil, roleErr
		}

		role, err := copyData(actualRole.Data().(*DcimDeviceRole))
		if err != nil {
			return nil, err
		}
		role.Tags = nil

		if !actualRole.HasChanged() {
			role = &DcimDeviceRole{
				ID: actualRole.ID(),
			}
		}
		vw.VirtualMachine.Role = role

		vw.objectsToReconcile = append(vw.objectsToReconcile, roleObjectsToReconcile...)

		if actualPlatform != nil {
			platformObjectsToReconcile, platformErr := actualPlatform.Patch(intendedPlatform, intendedNestedObjects)
			if platformErr != nil {
				return nil, platformErr
			}

			platform, err := copyData(actualPlatform.Data().(*DcimPlatform))
			if err != nil {
				return nil, err
			}
			platform.Tags = nil

			if !actualPlatform.HasChanged() {
				platform = &DcimPlatform{
					ID: actualPlatform.ID(),
				}
			}
			vw.VirtualMachine.Platform = platform

			vw.objectsToReconcile = append(vw.objectsToReconcile, platformObjectsToReconcile...)
		}

		if actualDevice != nil {
			deviceObjectsToReconcile, deviceErr := actualDevice.Patch(intendedDevice, intendedNestedObjects)
			if deviceErr != nil {
				return nil, deviceErr
			}

			device, err := copyData(actualDevice.Data().(*DcimDevice))
			if err != nil {
				return nil, err
			}
			device.Tags = nil

			if !actualDevice.HasChanged() {
				device = &DcimDevice{
					ID: actualDevice.ID(),
				}
			}
			vw.VirtualMachine.Device = device

			vw.objectsToReconcile = append(vw.objectsToReconcile, deviceObjectsToReconcile...)
		}

		tagsToMerge := mergeTags(vw.VirtualMachine.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualMachine.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualMachine.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
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

// VirtualizationInterfaceDataWrapper represents a virtualization interface data wrapper
type VirtualizationInterfaceDataWrapper struct {
	BaseDataWrapper
	VirtualInterface *VirtualizationInterface
}

func (*VirtualizationInterfaceDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationInterfaceDataWrapper) Data() any {
	return vw.VirtualInterface
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationInterfaceDataWrapper) IsValid() bool {
	if vw.VirtualInterface != nil && !vw.hasParent && vw.VirtualInterface.Name == "" {
		vw.VirtualInterface = nil
	}
	return vw.VirtualInterface != nil
}

// Normalise normalises the data
func (vw *VirtualizationInterfaceDataWrapper) Normalise() {
	if vw.IsValid() && vw.VirtualInterface.Tags != nil && len(vw.VirtualInterface.Tags) == 0 {
		vw.VirtualInterface.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationInterfaceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.VirtualInterface != nil && vw.hasParent && vw.VirtualInterface.Name == "" {
		vw.VirtualInterface = nil
	}

	objects := make([]ComparableData, 0)

	if vw.VirtualInterface == nil && vw.intended {
		return objects, nil
	}

	if vw.VirtualInterface == nil && vw.hasParent {
		vw.VirtualInterface = NewVirtualizationInterface()
		vw.placeholder = true
	}

	virtualMachine := VirtualizationVirtualMachineDataWrapper{VirtualMachine: vw.VirtualInterface.VirtualMachine, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	vmo, err := virtualMachine.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, vmo...)

	vw.VirtualInterface.VirtualMachine = virtualMachine.VirtualMachine

	if vw.VirtualInterface.Tags != nil {
		for _, t := range vw.VirtualInterface.Tags {
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
func (vw *VirtualizationInterfaceDataWrapper) DataType() string {
	return VirtualizationInterfaceObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationInterfaceDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.VirtualInterface.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationInterfaceDataWrapper) ID() int {
	return vw.VirtualInterface.ID
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationInterfaceDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationInterfaceDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range vw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualVirtualMachine := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualInterface.VirtualMachine))
	intendedVirtualMachine := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualInterface.VirtualMachine))

	reconciliationRequired := true

	if intended != nil {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		vw.VirtualInterface.ID = intended.VirtualInterface.ID
		vw.VirtualInterface.Name = intended.VirtualInterface.Name

		if actualVirtualMachine.IsPlaceholder() && intended.VirtualInterface.VirtualMachine != nil {
			intendedVirtualMachine = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualInterface.VirtualMachine))
		}

		virtualMachineObjectsToReconcile, virtualMachineErr := actualVirtualMachine.Patch(intendedVirtualMachine, intendedNestedObjects)
		if virtualMachineErr != nil {
			return nil, virtualMachineErr
		}

		virtualMachine, err := copyData(actualVirtualMachine.Data().(*VirtualizationVirtualMachine))
		if err != nil {
			return nil, err
		}
		virtualMachine.Tags = nil

		if !actualVirtualMachine.HasChanged() {
			virtualMachine = &VirtualizationVirtualMachine{
				ID: actualVirtualMachine.ID(),
			}

			intendedVirtualMachineID := intendedVirtualMachine.ID()
			if intended.VirtualInterface.VirtualMachine != nil {
				intendedVirtualMachineID = intended.VirtualInterface.VirtualMachine.ID
			}

			intended.VirtualInterface.VirtualMachine = &VirtualizationVirtualMachine{
				ID: intendedVirtualMachineID,
			}
		}

		vw.VirtualInterface.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		if vw.VirtualInterface.Enabled == nil {
			vw.VirtualInterface.Enabled = intended.VirtualInterface.Enabled
		}

		if vw.VirtualInterface.MTU == nil {
			vw.VirtualInterface.MTU = intended.VirtualInterface.MTU
		}

		if vw.VirtualInterface.MACAddress == nil {
			vw.VirtualInterface.MACAddress = intended.VirtualInterface.MACAddress
		}

		if vw.VirtualInterface.Description == nil {
			vw.VirtualInterface.Description = intended.VirtualInterface.Description
		}

		tagsToMerge := mergeTags(vw.VirtualInterface.Tags, intended.VirtualInterface.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualInterface.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualInterface.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		virtualMachineObjectsToReconcile, virtualMachineErr := actualVirtualMachine.Patch(intendedVirtualMachine, intendedNestedObjects)
		if virtualMachineErr != nil {
			return nil, virtualMachineErr
		}

		virtualMachine, err := copyData(actualVirtualMachine.Data().(*VirtualizationVirtualMachine))
		if err != nil {
			return nil, err
		}
		virtualMachine.Tags = nil

		if !actualVirtualMachine.HasChanged() {
			virtualMachine = &VirtualizationVirtualMachine{
				ID: actualVirtualMachine.ID(),
			}
		}
		vw.VirtualInterface.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		tagsToMerge := mergeTags(vw.VirtualInterface.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualInterface.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualInterface.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationInterfaceDataWrapper) SetDefaults() {}

// VirtualizationVirtualDiskDataWrapper represents a virtualization interface data wrapper
type VirtualizationVirtualDiskDataWrapper struct {
	BaseDataWrapper
	VirtualDisk *VirtualizationVirtualDisk
}

func (*VirtualizationVirtualDiskDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (vw *VirtualizationVirtualDiskDataWrapper) Data() any {
	return vw.VirtualDisk
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationVirtualDiskDataWrapper) IsValid() bool {
	if vw.VirtualDisk != nil && !vw.hasParent && vw.VirtualDisk.Name == "" {
		vw.VirtualDisk = nil
	}
	return vw.VirtualDisk != nil
}

// Normalise normalises the data
func (vw *VirtualizationVirtualDiskDataWrapper) Normalise() {
	if vw.IsValid() && vw.VirtualDisk.Tags != nil && len(vw.VirtualDisk.Tags) == 0 {
		vw.VirtualDisk.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationVirtualDiskDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.VirtualDisk != nil && vw.hasParent && vw.VirtualDisk.Name == "" {
		vw.VirtualDisk = nil
	}

	objects := make([]ComparableData, 0)

	if vw.VirtualDisk == nil && vw.intended {
		return objects, nil
	}

	if vw.VirtualDisk == nil && vw.hasParent {
		vw.VirtualDisk = NewVirtualizationVirtualDisk()
		vw.placeholder = true
	}

	virtualMachine := VirtualizationVirtualMachineDataWrapper{VirtualMachine: vw.VirtualDisk.VirtualMachine, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	vmo, err := virtualMachine.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, vmo...)

	vw.VirtualDisk.VirtualMachine = virtualMachine.VirtualMachine

	if vw.VirtualDisk.Tags != nil {
		for _, t := range vw.VirtualDisk.Tags {
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
func (vw *VirtualizationVirtualDiskDataWrapper) DataType() string {
	return VirtualizationVirtualDiskObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationVirtualDiskDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": vw.VirtualDisk.Name,
	}
}

// ID returns the ID of the data
func (vw *VirtualizationVirtualDiskDataWrapper) ID() int {
	return vw.VirtualDisk.ID
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationVirtualDiskDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationVirtualDiskDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range vw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualVirtualMachine := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VirtualDisk.VirtualMachine))
	intendedVirtualMachine := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VirtualDisk.VirtualMachine))

	reconciliationRequired := true

	if intended != nil {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		vw.VirtualDisk.ID = intended.VirtualDisk.ID
		vw.VirtualDisk.Name = intended.VirtualDisk.Name

		if actualVirtualMachine.IsPlaceholder() && intended.VirtualDisk.VirtualMachine != nil {
			intendedVirtualMachine = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VirtualDisk.VirtualMachine))
		}

		virtualMachineObjectsToReconcile, virtualMachineErr := actualVirtualMachine.Patch(intendedVirtualMachine, intendedNestedObjects)
		if virtualMachineErr != nil {
			return nil, virtualMachineErr
		}

		virtualMachine, err := copyData(actualVirtualMachine.Data().(*VirtualizationVirtualMachine))
		if err != nil {
			return nil, err
		}
		virtualMachine.Tags = nil

		if !actualVirtualMachine.HasChanged() {
			virtualMachine = &VirtualizationVirtualMachine{
				ID: actualVirtualMachine.ID(),
			}

			intendedVirtualMachineID := intendedVirtualMachine.ID()
			if intended.VirtualDisk.VirtualMachine != nil {
				intendedVirtualMachineID = intended.VirtualDisk.VirtualMachine.ID
			}

			intended.VirtualDisk.VirtualMachine = &VirtualizationVirtualMachine{
				ID: intendedVirtualMachineID,
			}
		}

		vw.VirtualDisk.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		if vw.VirtualDisk.Description == nil {
			vw.VirtualDisk.Description = intended.VirtualDisk.Description
		}

		tagsToMerge := mergeTags(vw.VirtualDisk.Tags, intended.VirtualDisk.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualDisk.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualDisk.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		virtualMachineObjectsToReconcile, virtualMachineErr := actualVirtualMachine.Patch(intendedVirtualMachine, intendedNestedObjects)
		if virtualMachineErr != nil {
			return nil, virtualMachineErr
		}

		virtualMachine, err := copyData(actualVirtualMachine.Data().(*VirtualizationVirtualMachine))
		if err != nil {
			return nil, err
		}
		virtualMachine.Tags = nil

		if !actualVirtualMachine.HasChanged() {
			virtualMachine = &VirtualizationVirtualMachine{
				ID: actualVirtualMachine.ID(),
			}
		}
		vw.VirtualDisk.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		tagsToMerge := mergeTags(vw.VirtualDisk.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VirtualDisk.Tags = tagsToMerge
		}

		for _, t := range vw.VirtualDisk.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		vw.hasChanged = true
		vw.objectsToReconcile = append(vw.objectsToReconcile, vw)
	}

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationVirtualDiskDataWrapper) SetDefaults() {}
