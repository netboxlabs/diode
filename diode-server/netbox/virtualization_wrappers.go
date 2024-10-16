package netbox

import (
	"errors"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// VirtualizationClusterGroupDataWrapper represents a virtualization cluster group data wrapper
type VirtualizationClusterGroupDataWrapper struct {
	BaseDataWrapper
	ClusterGroup *VirtualizationClusterGroup
}

func (*VirtualizationClusterGroupDataWrapper) comparableData() {}

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationClusterGroupDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	clusterGroup, err := FromProtoClusterGroupEntity(entity)
	if err != nil {
		return err
	}
	vw.ClusterGroup = clusterGroup
	return nil
}

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
		vw.SetDefaults()

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

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationClusterTypeDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	clusterType, err := FromProtoClusterTypeEntity(entity)
	if err != nil {
		return err
	}
	vw.ClusterType = clusterType
	return nil
}

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
		vw.SetDefaults()

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

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationClusterDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	cluster, err := FromProtoClusterEntity(entity)
	if err != nil {
		return err
	}
	vw.Cluster = cluster
	return nil
}

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
		vw.SetDefaults()

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

	dedupObjectsToReconcile, err := dedupObjectsToReconcile(vw.objectsToReconcile)
	if err != nil {
		return nil, err
	}
	vw.objectsToReconcile = dedupObjectsToReconcile

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationClusterDataWrapper) SetDefaults() {
	if vw.Cluster.Status == nil || *vw.Cluster.Status == "" {
		status := "active"
		vw.Cluster.Status = &status
	}
}

// VirtualizationVirtualMachineDataWrapper represents a virtualization virtual machine data wrapper
type VirtualizationVirtualMachineDataWrapper struct {
	BaseDataWrapper
	VirtualMachine *VirtualizationVirtualMachine
}

func (*VirtualizationVirtualMachineDataWrapper) comparableData() {}

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationVirtualMachineDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	virtualMachine, err := FromProtoVirtualMachineEntity(entity)
	if err != nil {
		return err
	}
	vw.VirtualMachine = virtualMachine
	return nil
}

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

	// Ignore primary IP addresses and device for time being
	vw.VirtualMachine.PrimaryIPv4 = nil
	vw.VirtualMachine.PrimaryIPv6 = nil
	vw.VirtualMachine.Device = nil

	if vw.VirtualMachine.Cluster != nil {
		cluster := VirtualizationClusterDataWrapper{Cluster: vw.VirtualMachine.Cluster, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

		co, err := cluster.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, co...)

		vw.VirtualMachine.Cluster = cluster.Cluster
	}

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
	params := map[string]string{
		"q": vw.VirtualMachine.Name,
	}
	if vw.VirtualMachine.Site != nil {
		params["site__name"] = vw.VirtualMachine.Site.Name
	} else if vw.VirtualMachine.Cluster != nil {
		params["cluster__name"] = vw.VirtualMachine.Cluster.Name

		if vw.VirtualMachine.Cluster.Site != nil {
			params["cluster__site__name"] = vw.VirtualMachine.Cluster.Site.Name
		}
	}
	return params
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

		if vw.VirtualMachine.Status == nil || *vw.VirtualMachine.Status == "" {
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

		if actualCluster != nil {
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
		} else if intended.VirtualMachine.Cluster != nil {
			clusterID := intended.VirtualMachine.Cluster.ID
			vw.VirtualMachine.Cluster = &VirtualizationCluster{
				ID: clusterID,
			}
			intended.VirtualMachine.Cluster = &VirtualizationCluster{
				ID: clusterID,
			}
		}

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
		} else if intended.VirtualMachine.Platform != nil {
			platformID := intended.VirtualMachine.Platform.ID
			vw.VirtualMachine.Platform = &DcimPlatform{
				ID: platformID,
			}
			intended.VirtualMachine.Platform = &DcimPlatform{
				ID: platformID,
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
		vw.SetDefaults()

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

		if actualCluster != nil {
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

	dedupObjectsToReconcile, err := dedupObjectsToReconcile(vw.objectsToReconcile)
	if err != nil {
		return nil, err
	}
	vw.objectsToReconcile = dedupObjectsToReconcile

	return vw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (vw *VirtualizationVirtualMachineDataWrapper) SetDefaults() {
	if vw.VirtualMachine.Status == nil || *vw.VirtualMachine.Status == "" {
		status := "active"
		vw.VirtualMachine.Status = &status
	}
}

// VirtualizationVMInterfaceDataWrapper represents a virtualization VM interface data wrapper
type VirtualizationVMInterfaceDataWrapper struct {
	BaseDataWrapper
	VMInterface *VirtualizationVMInterface
}

func (*VirtualizationVMInterfaceDataWrapper) comparableData() {}

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationVMInterfaceDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	vmInterface, err := FromProtoVMInterfaceEntity(entity)
	if err != nil {
		return err
	}
	vw.VMInterface = vmInterface
	return nil
}

// Data returns the DeviceRole
func (vw *VirtualizationVMInterfaceDataWrapper) Data() any {
	return vw.VMInterface
}

// IsValid returns true if the DeviceRole is not nil
func (vw *VirtualizationVMInterfaceDataWrapper) IsValid() bool {
	if vw.VMInterface != nil && !vw.hasParent && vw.VMInterface.Name == "" {
		vw.VMInterface = nil
	}
	return vw.VMInterface != nil
}

// Normalise normalises the data
func (vw *VirtualizationVMInterfaceDataWrapper) Normalise() {
	if vw.IsValid() && vw.VMInterface.Tags != nil && len(vw.VMInterface.Tags) == 0 {
		vw.VMInterface.Tags = nil
	}
	vw.intended = true
}

// NestedObjects returns all nested objects
func (vw *VirtualizationVMInterfaceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(vw.nestedObjects) > 0 {
		return vw.nestedObjects, nil
	}

	if vw.VMInterface != nil && vw.hasParent && vw.VMInterface.Name == "" {
		vw.VMInterface = nil
	}

	objects := make([]ComparableData, 0)

	if vw.VMInterface == nil && vw.intended {
		return objects, nil
	}

	if vw.VMInterface == nil && vw.hasParent {
		vw.VMInterface = NewVirtualizationVMInterface()
		vw.placeholder = true
	}

	virtualMachine := VirtualizationVirtualMachineDataWrapper{VirtualMachine: vw.VMInterface.VirtualMachine, BaseDataWrapper: BaseDataWrapper{placeholder: vw.placeholder, hasParent: true, intended: vw.intended}}

	vmo, err := virtualMachine.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, vmo...)

	vw.VMInterface.VirtualMachine = virtualMachine.VirtualMachine

	if vw.VMInterface.Tags != nil {
		for _, t := range vw.VMInterface.Tags {
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
func (vw *VirtualizationVMInterfaceDataWrapper) DataType() string {
	return VirtualizationVMInterfaceObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (vw *VirtualizationVMInterfaceDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": vw.VMInterface.Name,
	}
	if vw.VMInterface.VirtualMachine != nil {
		params["virtual_machine__name"] = vw.VMInterface.VirtualMachine.Name

		if vw.VMInterface.VirtualMachine.Site != nil {
			params["virtual_machine__site__name"] = vw.VMInterface.VirtualMachine.Site.Name
		}
	}
	return params
}

// ID returns the ID of the data
func (vw *VirtualizationVMInterfaceDataWrapper) ID() int {
	return vw.VMInterface.ID
}

// Patch creates patches between the actual, intended and current data
func (vw *VirtualizationVMInterfaceDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*VirtualizationVMInterfaceDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range vw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualVirtualMachine := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", vw.VMInterface.VirtualMachine))
	intendedVirtualMachine := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", vw.VMInterface.VirtualMachine))

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

		vw.VMInterface.ID = intended.VMInterface.ID
		vw.VMInterface.Name = intended.VMInterface.Name

		if actualVirtualMachine.IsPlaceholder() && intended.VMInterface.VirtualMachine != nil {
			intendedVirtualMachine = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.VMInterface.VirtualMachine))
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
			if intended.VMInterface.VirtualMachine != nil {
				intendedVirtualMachineID = intended.VMInterface.VirtualMachine.ID
			}

			intended.VMInterface.VirtualMachine = &VirtualizationVirtualMachine{
				ID: intendedVirtualMachineID,
			}
		}

		vw.VMInterface.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		if vw.VMInterface.Enabled == nil {
			vw.VMInterface.Enabled = intended.VMInterface.Enabled
		}

		if vw.VMInterface.MTU == nil {
			vw.VMInterface.MTU = intended.VMInterface.MTU
		}

		if vw.VMInterface.MACAddress == nil {
			vw.VMInterface.MACAddress = intended.VMInterface.MACAddress
		}

		if vw.VMInterface.Description == nil {
			vw.VMInterface.Description = intended.VMInterface.Description
		}

		tagsToMerge := mergeTags(vw.VMInterface.Tags, intended.VMInterface.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VMInterface.Tags = tagsToMerge
		}

		for _, t := range vw.VMInterface.Tags {
			if t.ID == 0 {
				vw.objectsToReconcile = append(vw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(vw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		vw.SetDefaults()

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
		vw.VMInterface.VirtualMachine = virtualMachine

		vw.objectsToReconcile = append(vw.objectsToReconcile, virtualMachineObjectsToReconcile...)

		tagsToMerge := mergeTags(vw.VMInterface.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			vw.VMInterface.Tags = tagsToMerge
		}

		for _, t := range vw.VMInterface.Tags {
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
func (vw *VirtualizationVMInterfaceDataWrapper) SetDefaults() {}

// VirtualizationVirtualDiskDataWrapper represents a virtualization disk data wrapper
type VirtualizationVirtualDiskDataWrapper struct {
	BaseDataWrapper
	VirtualDisk *VirtualizationVirtualDisk
}

func (*VirtualizationVirtualDiskDataWrapper) comparableData() {}

// FromProtoEntity sets the data from a proto entity
func (vw *VirtualizationVirtualDiskDataWrapper) FromProtoEntity(entity *diodepb.Entity) error {
	virtualDisk, err := FromProtoVirtualDiskEntity(entity)
	if err != nil {
		return err
	}
	vw.VirtualDisk = virtualDisk
	return nil
}

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
	params := map[string]string{
		"q": vw.VirtualDisk.Name,
	}
	if vw.VirtualDisk.VirtualMachine != nil {
		params["virtual_machine__name"] = vw.VirtualDisk.VirtualMachine.Name

		if vw.VirtualDisk.VirtualMachine.Site != nil {
			params["virtual_machine__site__name"] = vw.VirtualDisk.VirtualMachine.Site.Name
		}
	}
	return params
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
		vw.SetDefaults()

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
