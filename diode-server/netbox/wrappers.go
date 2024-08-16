package netbox

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gosimple/slug"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/hashstructure/v2"
)

// ComparableData is an interface for NetBox comparable data
type ComparableData interface {
	comparableData()

	// Data returns the data
	Data() any

	// IsValid checks if the data is not nil
	IsValid() bool

	// Normalise normalises the data
	Normalise()

	// NestedObjects returns all nested objects
	NestedObjects() ([]ComparableData, error)

	// DataType returns the data type
	DataType() string

	// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
	ObjectStateQueryParams() map[string]string

	// ID returns the ID of the data
	ID() int

	// IsPlaceholder returns true if the data is a placeholder
	IsPlaceholder() bool

	// SetDefaults sets the default values for the data
	SetDefaults()

	// Patch creates patches between the actual, intended and current data
	Patch(ComparableData, map[string]ComparableData) ([]ComparableData, error)

	// HasChanged returns true if the data has changed
	HasChanged() bool
}

// BaseDataWrapper is the base struct for all data wrappers
type BaseDataWrapper struct {
	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

// IsPlaceholder returns true if the data is a placeholder
func (bw *BaseDataWrapper) IsPlaceholder() bool {
	return bw.placeholder
}

// HasChanged returns true if the data has changed
func (bw *BaseDataWrapper) HasChanged() bool {
	return bw.hasChanged
}

func copyData[T any](srcData *T) (*T, error) {
	var dstData T
	if err := copier.Copy(&dstData, srcData); err != nil {
		return nil, err
	}
	return &dstData, nil
}

// DcimDeviceDataWrapper represents a DCIM device data wrapper
type DcimDeviceDataWrapper struct {
	BaseDataWrapper
	Device *DcimDevice
}

func (*DcimDeviceDataWrapper) comparableData() {}

// Data returns the Device
func (dw *DcimDeviceDataWrapper) Data() any {
	return dw.Device
}

// IsValid returns true if the Device is not nil
func (dw *DcimDeviceDataWrapper) IsValid() bool {
	if dw.Device != nil && !dw.hasParent && dw.Device.Name == "" {
		dw.Device = nil
	}
	return dw.Device != nil
}

// Normalise normalises the data
func (dw *DcimDeviceDataWrapper) Normalise() {
	if dw.IsValid() && dw.Device.Tags != nil && len(dw.Device.Tags) == 0 {
		dw.Device.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Device != nil && dw.hasParent && dw.Device.Name == "" {
		dw.Device = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Device == nil && dw.intended {
		return objects, nil
	}

	if dw.Device == nil && dw.hasParent {
		dw.Device = NewDcimDevice()
		dw.placeholder = true
	}

	// Ignore primary IP addresses for time being
	dw.Device.PrimaryIPv4 = nil
	dw.Device.PrimaryIPv6 = nil

	site := DcimSiteDataWrapper{Site: dw.Device.Site, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	dw.Device.Site = site.Site

	if dw.Device.Platform != nil {
		platform := DcimPlatformDataWrapper{Platform: dw.Device.Platform, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

		po, err := platform.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, po...)

		dw.Device.Platform = platform.Platform
	}

	deviceType := DcimDeviceTypeDataWrapper{DeviceType: dw.Device.DeviceType, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	dto, err := deviceType.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dto...)

	dw.Device.DeviceType = deviceType.DeviceType

	deviceRole := DcimDeviceRoleDataWrapper{DeviceRole: dw.Device.Role, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	dro, err := deviceRole.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dro...)

	dw.Device.Role = deviceRole.DeviceRole

	if dw.Device.Tags != nil {
		for _, t := range dw.Device.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimDeviceDataWrapper) DataType() string {
	return DcimDeviceObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimDeviceDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": dw.Device.Name,
	}
	if dw.Device.Site != nil {
		params["site__name"] = dw.Device.Site.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimDeviceDataWrapper) ID() int {
	return dw.Device.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimDeviceDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Device.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Device.Site))

	actualPlatform := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Device.Platform))
	intendedPlatform := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Device.Platform))

	actualDeviceType := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Device.DeviceType))
	intendedDeviceType := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Device.DeviceType))

	actualRole := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Device.Role))
	intendedRole := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Device.Role))

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

		dw.Device.ID = intended.Device.ID
		dw.Device.Name = intended.Device.Name

		if dw.Device.Status == nil {
			dw.Device.Status = intended.Device.Status
		}

		if dw.Device.Description == nil {
			dw.Device.Description = intended.Device.Description
		}

		if dw.Device.Comments == nil {
			dw.Device.Comments = intended.Device.Comments
		}

		if dw.Device.AssetTag == nil {
			dw.Device.AssetTag = intended.Device.AssetTag
		}

		if dw.Device.Serial == nil {
			dw.Device.Serial = intended.Device.Serial
		}

		if actualSite.IsPlaceholder() && intended.Device.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Site))
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
			if intended.Device.Site != nil {
				intendedSiteID = intended.Device.Site.ID
			}

			intended.Device.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		dw.Device.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		if actualPlatform != nil {
			if actualPlatform.IsPlaceholder() && intended.Device.Platform != nil {
				intendedPlatform = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Platform))
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
				if intended.Device.Platform != nil {
					intendedPlatformID = intended.Device.Platform.ID
				}

				intended.Device.Platform = &DcimPlatform{
					ID: intendedPlatformID,
				}
			}

			dw.Device.Platform = platform

			dw.objectsToReconcile = append(dw.objectsToReconcile, platformObjectsToReconcile...)
		} else {
			if intended.Device.Platform != nil {
				platformID := intended.Device.Platform.ID
				dw.Device.Platform = &DcimPlatform{
					ID: platformID,
				}
				intended.Device.Platform = &DcimPlatform{
					ID: platformID,
				}
			}
		}

		if actualDeviceType.IsPlaceholder() && intended.Device.DeviceType != nil {
			intendedDeviceType = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.DeviceType))
		}

		deviceTypeObjectsToReconcile, deviceTypeErr := actualDeviceType.Patch(intendedDeviceType, intendedNestedObjects)
		if deviceTypeErr != nil {
			return nil, deviceTypeErr
		}

		deviceType, err := copyData(actualDeviceType.Data().(*DcimDeviceType))
		if err != nil {
			return nil, err
		}
		deviceType.Tags = nil

		if !actualDeviceType.HasChanged() {
			deviceType = &DcimDeviceType{
				ID: actualDeviceType.ID(),
			}

			intendedDeviceTypeID := intendedDeviceType.ID()
			if intended.Device.DeviceType != nil {
				intendedDeviceTypeID = intended.Device.DeviceType.ID
			}

			intended.Device.DeviceType = &DcimDeviceType{
				ID: intendedDeviceTypeID,
			}
		}

		dw.Device.DeviceType = deviceType

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceTypeObjectsToReconcile...)

		if actualRole.IsPlaceholder() && intended.Device.Role != nil {
			intendedRole = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Role))
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
			if intended.Device.Role != nil {
				intendedRoleID = intended.Device.Role.ID
			}

			intended.Device.Role = &DcimDeviceRole{
				ID: intendedRoleID,
			}
		}

		dw.Device.Role = role

		dw.objectsToReconcile = append(dw.objectsToReconcile, roleObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Device.Tags, intended.Device.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Device.Tags = tagsToMerge
		}

		for _, t := range dw.Device.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

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
		dw.Device.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

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
			dw.Device.Platform = platform

			dw.objectsToReconcile = append(dw.objectsToReconcile, platformObjectsToReconcile...)
		}

		deviceTypeObjectsToReconcile, deviceTypeErr := actualDeviceType.Patch(intendedDeviceType, intendedNestedObjects)
		if deviceTypeErr != nil {
			return nil, deviceTypeErr
		}

		deviceType, err := copyData(actualDeviceType.Data().(*DcimDeviceType))
		if err != nil {
			return nil, err
		}
		deviceType.Tags = nil

		if !actualDeviceType.HasChanged() {
			deviceType = &DcimDeviceType{
				ID: actualDeviceType.ID(),
			}
		}
		dw.Device.DeviceType = deviceType

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceTypeObjectsToReconcile...)

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
		dw.Device.Role = role

		dw.objectsToReconcile = append(dw.objectsToReconcile, roleObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Device.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Device.Tags = tagsToMerge
		}

		for _, t := range dw.Device.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	dedupObjectsToReconcile, err := dedupObjectsToReconcile(dw.objectsToReconcile)
	if err != nil {
		return nil, err
	}
	dw.objectsToReconcile = dedupObjectsToReconcile

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device
func (dw *DcimDeviceDataWrapper) SetDefaults() {
	if dw.Device.Status == nil {
		status := DcimDeviceStatusActive
		dw.Device.Status = &status
	}
}

// DcimDeviceRoleDataWrapper represents a DCIM device role data wrapper
type DcimDeviceRoleDataWrapper struct {
	BaseDataWrapper
	DeviceRole *DcimDeviceRole
}

func (*DcimDeviceRoleDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (dw *DcimDeviceRoleDataWrapper) Data() any {
	return dw.DeviceRole
}

// IsValid returns true if the DeviceRole is not nil
func (dw *DcimDeviceRoleDataWrapper) IsValid() bool {
	if dw.DeviceRole != nil && !dw.hasParent && dw.DeviceRole.Name == "" {
		dw.DeviceRole = nil
	}
	return dw.DeviceRole != nil
}

// Normalise normalises the data
func (dw *DcimDeviceRoleDataWrapper) Normalise() {
	if dw.IsValid() && dw.DeviceRole.Tags != nil && len(dw.DeviceRole.Tags) == 0 {
		dw.DeviceRole.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceRoleDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.DeviceRole != nil && dw.hasParent && dw.DeviceRole.Name == "" {
		dw.DeviceRole = nil
	}

	objects := make([]ComparableData, 0)

	if dw.DeviceRole == nil && dw.intended {
		return objects, nil
	}

	if dw.DeviceRole == nil && dw.hasParent {
		dw.DeviceRole = NewDcimDeviceRole()
		dw.placeholder = true
	}

	if dw.DeviceRole.Slug == "" {
		dw.DeviceRole.Slug = slug.Make(dw.DeviceRole.Name)
	}

	if dw.DeviceRole.Tags != nil {
		for _, t := range dw.DeviceRole.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimDeviceRoleDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimDeviceRoleDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.DeviceRole.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimDeviceRoleDataWrapper) ID() int {
	return dw.DeviceRole.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimDeviceRoleDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceRoleDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	reconciliationRequired := true

	if intended != nil {
		dw.DeviceRole.ID = intended.DeviceRole.ID
		dw.DeviceRole.Name = intended.DeviceRole.Name
		dw.DeviceRole.Slug = intended.DeviceRole.Slug

		if dw.IsPlaceholder() || dw.DeviceRole.Color == nil {
			dw.DeviceRole.Color = intended.DeviceRole.Color
		}

		if dw.DeviceRole.Description == nil {
			dw.DeviceRole.Description = intended.DeviceRole.Description
		}

		tagsToMerge := mergeTags(dw.DeviceRole.Tags, intended.DeviceRole.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.DeviceRole.Tags = tagsToMerge
		}

		for _, t := range dw.DeviceRole.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		tagsToMerge := mergeTags(dw.DeviceRole.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.DeviceRole.Tags = tagsToMerge
		}

		for _, t := range dw.DeviceRole.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device role
func (dw *DcimDeviceRoleDataWrapper) SetDefaults() {
	if dw.DeviceRole.Color == nil {
		color := "000000"
		dw.DeviceRole.Color = &color
	}
}

// DcimDeviceTypeDataWrapper represents a DCIM device type data wrapper
type DcimDeviceTypeDataWrapper struct {
	BaseDataWrapper
	DeviceType *DcimDeviceType
}

func (*DcimDeviceTypeDataWrapper) comparableData() {}

// Data returns the DeviceType
func (dw *DcimDeviceTypeDataWrapper) Data() any {
	return dw.DeviceType
}

// IsValid returns true if the DeviceType is not nil
func (dw *DcimDeviceTypeDataWrapper) IsValid() bool {
	if dw.DeviceType != nil && !dw.hasParent && dw.DeviceType.Model == "" {
		dw.DeviceType = nil
	}
	return dw.DeviceType != nil
}

// Normalise normalises the data
func (dw *DcimDeviceTypeDataWrapper) Normalise() {
	if dw.IsValid() && dw.DeviceType.Tags != nil && len(dw.DeviceType.Tags) == 0 {
		dw.DeviceType.Tags = nil
	}
	dw.intended = true
}

// DataType returns the data type
func (dw *DcimDeviceTypeDataWrapper) DataType() string {
	return DcimDeviceTypeObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimDeviceTypeDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": dw.DeviceType.Model,
	}
	if dw.DeviceType.Manufacturer != nil {
		params["manufacturer__name"] = dw.DeviceType.Manufacturer.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimDeviceTypeDataWrapper) ID() int {
	return dw.DeviceType.ID
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceTypeDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.DeviceType != nil && dw.hasParent && dw.DeviceType.Model == "" {
		dw.DeviceType = nil
	}

	objects := make([]ComparableData, 0)

	if dw.DeviceType == nil && dw.intended {
		return objects, nil
	}

	if dw.DeviceType == nil && dw.hasParent {
		dw.DeviceType = NewDcimDeviceType()
		dw.placeholder = true
	}

	if dw.DeviceType.Slug == "" {
		dw.DeviceType.Slug = slug.Make(dw.DeviceType.Model)
	}

	manufacturer := DcimManufacturerDataWrapper{Manufacturer: dw.DeviceType.Manufacturer, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	mo, err := manufacturer.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, mo...)

	dw.DeviceType.Manufacturer = manufacturer.Manufacturer

	if dw.DeviceType.Tags != nil && len(dw.DeviceType.Tags) == 0 {
		dw.DeviceType.Tags = nil
	}

	if dw.DeviceType.Tags != nil {
		for _, t := range dw.DeviceType.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimDeviceTypeDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceTypeDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualManufacturerKey := fmt.Sprintf("%p", dw.DeviceType.Manufacturer)
	actualManufacturer := extractFromObjectsMap(actualNestedObjectsMap, actualManufacturerKey)
	intendedManufacturer := extractFromObjectsMap(intendedNestedObjects, actualManufacturerKey)

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

		dw.DeviceType.ID = intended.DeviceType.ID
		dw.DeviceType.Model = intended.DeviceType.Model
		dw.DeviceType.Slug = intended.DeviceType.Slug

		if dw.DeviceType.Description == nil {
			dw.DeviceType.Description = intended.DeviceType.Description
		}

		if dw.DeviceType.Comments == nil {
			dw.DeviceType.Comments = intended.DeviceType.Comments
		}

		if dw.DeviceType.PartNumber == nil {
			dw.DeviceType.PartNumber = intended.DeviceType.PartNumber
		}

		if actualManufacturer.IsPlaceholder() && intended.DeviceType.Manufacturer != nil {
			intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.DeviceType.Manufacturer))
		}

		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}

		manufacturer, err := copyData(actualManufacturer.Data().(*DcimManufacturer))
		if err != nil {
			return nil, err
		}
		manufacturer.Tags = nil

		if !actualManufacturer.HasChanged() {
			manufacturer = &DcimManufacturer{
				ID: actualManufacturer.ID(),
			}

			intendedManufacturerID := intendedManufacturer.ID()
			if intended.DeviceType.Manufacturer != nil {
				intendedManufacturerID = intended.DeviceType.Manufacturer.ID
			}

			intended.DeviceType.Manufacturer = &DcimManufacturer{
				ID: intendedManufacturerID,
			}
		}

		dw.DeviceType.Manufacturer = manufacturer

		tagsToMerge := mergeTags(dw.DeviceType.Tags, intended.DeviceType.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.DeviceType.Tags = tagsToMerge
		}

		for _, t := range dw.DeviceType.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}

		manufacturer, err := copyData(actualManufacturer.Data().(*DcimManufacturer))
		if err != nil {
			return nil, err
		}
		manufacturer.Tags = nil

		if !actualManufacturer.HasChanged() {
			manufacturer = &DcimManufacturer{
				ID: actualManufacturer.ID(),
			}
		}
		dw.DeviceType.Manufacturer = manufacturer

		tagsToMerge := mergeTags(dw.DeviceType.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.DeviceType.Tags = tagsToMerge
		}

		for _, t := range dw.DeviceType.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (dw *DcimDeviceTypeDataWrapper) SetDefaults() {}

// DcimInterfaceDataWrapper represents a DCIM interface data wrapper
type DcimInterfaceDataWrapper struct {
	BaseDataWrapper
	Interface *DcimInterface
}

func (*DcimInterfaceDataWrapper) comparableData() {}

// Data returns the Interface
func (dw *DcimInterfaceDataWrapper) Data() any {
	return dw.Interface
}

// IsValid returns true if the Interface is not nil
func (dw *DcimInterfaceDataWrapper) IsValid() bool {
	if dw.Interface != nil && !dw.hasParent && dw.Interface.Name == "" {
		dw.Interface = nil
	}

	if dw.Interface != nil {
		if err := dw.Interface.Validate(); err != nil {
			return false
		}
	}

	return dw.Interface != nil
}

// Normalise normalises the data
func (dw *DcimInterfaceDataWrapper) Normalise() {
	if dw.IsValid() && dw.Interface.Tags != nil && len(dw.Interface.Tags) == 0 {
		dw.Interface.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimInterfaceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Interface != nil && dw.hasParent && dw.Interface.Name == "" {
		dw.Interface = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Interface == nil && dw.intended {
		return objects, nil
	}

	if dw.Interface == nil && dw.hasParent {
		dw.Interface = NewDcimInterface()
		dw.placeholder = true
	}

	device := DcimDeviceDataWrapper{Device: dw.Interface.Device, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	do, err := device.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, do...)

	dw.Interface.Device = device.Device

	if dw.Interface.Tags != nil {
		for _, t := range dw.Interface.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimInterfaceDataWrapper) DataType() string {
	return DcimInterfaceObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimInterfaceDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": dw.Interface.Name,
	}
	if dw.Interface.Device != nil {
		params["device__name"] = dw.Interface.Device.Name

		if dw.Interface.Device.Site != nil {
			params["device__site__name"] = dw.Interface.Device.Site.Name
		}
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimInterfaceDataWrapper) ID() int {
	return dw.Interface.ID
}

func (dw *DcimInterfaceDataWrapper) hash() string {
	var deviceName, siteName string
	if dw.Interface.Device != nil {
		deviceName = dw.Interface.Device.Name
		if dw.Interface.Device.Site != nil {
			siteName = dw.Interface.Device.Site.Name
		}
	}
	return slug.Make(fmt.Sprintf("%s-%s-%s", dw.Interface.Name, deviceName, siteName))
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimInterfaceDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimInterfaceDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualDevice := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Interface.Device))
	intendedDevice := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Interface.Device))

	reconciliationRequired := true

	if intended != nil && dw.hash() == intended.hash() {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		dw.Interface.ID = intended.Interface.ID
		dw.Interface.Name = intended.Interface.Name

		if actualDevice.IsPlaceholder() && intended.Interface.Device != nil {
			intendedDevice = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Interface.Device))
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
			if intended.Interface.Device != nil {
				intendedDeviceID = intended.Interface.Device.ID
			}

			intended.Interface.Device = &DcimDevice{
				ID: intendedDeviceID,
			}
		}

		dw.Interface.Device = device

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceObjectsToReconcile...)

		if dw.Interface.Label == nil {
			dw.Interface.Label = intended.Interface.Label
		}

		if dw.Interface.Type == nil {
			dw.Interface.Type = intended.Interface.Type
		}

		if dw.Interface.Enabled == nil {
			dw.Interface.Enabled = intended.Interface.Enabled
		}

		if dw.Interface.MTU == nil {
			dw.Interface.MTU = intended.Interface.MTU
		}

		if dw.Interface.MACAddress == nil {
			dw.Interface.MACAddress = intended.Interface.MACAddress
		}

		if dw.Interface.Speed == nil {
			dw.Interface.Speed = intended.Interface.Speed
		}

		if dw.Interface.WWN == nil {
			dw.Interface.WWN = intended.Interface.WWN
		}

		if dw.Interface.MgmtOnly == nil {
			dw.Interface.MgmtOnly = intended.Interface.MgmtOnly
		}

		if dw.Interface.Description == nil {
			dw.Interface.Description = intended.Interface.Description
		}

		if dw.Interface.MarkConnected == nil {
			dw.Interface.MarkConnected = intended.Interface.MarkConnected
		}

		if dw.Interface.Mode == nil {
			dw.Interface.Mode = intended.Interface.Mode
		}

		tagsToMerge := mergeTags(dw.Interface.Tags, intended.Interface.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Interface.Tags = tagsToMerge
		}

		for _, t := range dw.Interface.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

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
		dw.Interface.Device = device

		tagsToMerge := mergeTags(dw.Interface.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Interface.Tags = tagsToMerge
		}

		for _, t := range dw.Interface.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceObjectsToReconcile...)
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the interface
func (dw *DcimInterfaceDataWrapper) SetDefaults() {
	if dw.Interface.Type == nil {
		dw.Interface.Type = &DefaultInterfaceType
	}
}

// DcimManufacturerDataWrapper represents a DCIM manufacturer data wrapper
type DcimManufacturerDataWrapper struct {
	BaseDataWrapper
	Manufacturer *DcimManufacturer
}

func (*DcimManufacturerDataWrapper) comparableData() {}

// Data returns the Manufacturer
func (dw *DcimManufacturerDataWrapper) Data() any {
	return dw.Manufacturer
}

// IsValid returns true if the Manufacturer is not nil
func (dw *DcimManufacturerDataWrapper) IsValid() bool {
	if dw.Manufacturer != nil && !dw.hasParent && dw.Manufacturer.Name == "" {
		dw.Manufacturer = nil
	}
	return dw.Manufacturer != nil
}

// Normalise normalises the data
func (dw *DcimManufacturerDataWrapper) Normalise() {
	if dw.IsValid() && dw.Manufacturer.Tags != nil && len(dw.Manufacturer.Tags) == 0 {
		dw.Manufacturer.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimManufacturerDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Manufacturer != nil && dw.hasParent && dw.Manufacturer.Name == "" {
		dw.Manufacturer = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Manufacturer == nil && dw.intended {
		return objects, nil
	}

	if dw.Manufacturer == nil && dw.hasParent {
		dw.Manufacturer = NewDcimManufacturer()
		dw.placeholder = true
	}

	if dw.Manufacturer.Slug == "" {
		dw.Manufacturer.Slug = slug.Make(dw.Manufacturer.Name)
	}

	if dw.Manufacturer.Tags != nil {
		for _, t := range dw.Manufacturer.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimManufacturerDataWrapper) DataType() string {
	return DcimManufacturerObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimManufacturerDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.Manufacturer.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimManufacturerDataWrapper) ID() int {
	return dw.Manufacturer.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimManufacturerDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimManufacturerDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		dw.Manufacturer.ID = intended.Manufacturer.ID
		dw.Manufacturer.Name = intended.Manufacturer.Name
		dw.Manufacturer.Slug = intended.Manufacturer.Slug

		if dw.Manufacturer.Description == nil {
			dw.Manufacturer.Description = intended.Manufacturer.Description
		}

		tagsToMerge := mergeTags(dw.Manufacturer.Tags, intended.Manufacturer.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Manufacturer.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(dw.Manufacturer.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Manufacturer.Tags = tagsToMerge
		}
	}

	for _, t := range dw.Manufacturer.Tags {
		if t.ID == 0 {
			dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

func mergeTags(actualTags []*Tag, intendedTags []*Tag, intendedNestedObjects map[string]ComparableData) []*Tag {
	tagsToMerge := make([]*Tag, 0)
	tagsToCreate := make([]*Tag, 0)

	tagNamesToMerge := make([]string, 0)
	tagNamesToCreate := make([]string, 0)

	for _, t := range intendedTags {
		if !slices.Contains(tagNamesToMerge, t.Name) {
			tagNamesToMerge = append(tagNamesToMerge, t.Name)
			tagsToMerge = append(tagsToMerge, t)
		}
	}

	for _, t := range actualTags {
		tagKey := fmt.Sprintf("%p", t)
		tagWrapper := extractFromObjectsMap(intendedNestedObjects, tagKey)

		if !slices.Contains(tagNamesToMerge, t.Name) && tagWrapper != nil {
			tagNamesToMerge = append(tagNamesToMerge, t.Name)
			tagsToMerge = append(tagsToMerge, tagWrapper.Data().(*Tag))
			continue
		}

		if tagWrapper == nil {
			if !slices.Contains(tagNamesToCreate, t.Name) {
				tagNamesToCreate = append(tagNamesToCreate, t.Name)
				tagsToCreate = append(tagsToCreate, t)
			}
		}
	}

	return append(tagsToMerge, tagsToCreate...)
}

// SetDefaults sets the default values for the manufacturer
func (dw *DcimManufacturerDataWrapper) SetDefaults() {}

// DcimPlatformDataWrapper represents a DCIM platform data wrapper
type DcimPlatformDataWrapper struct {
	BaseDataWrapper
	Platform *DcimPlatform
}

func (*DcimPlatformDataWrapper) comparableData() {}

// Data returns the Platform
func (dw *DcimPlatformDataWrapper) Data() any {
	return dw.Platform
}

// IsValid returns true if the Platform is not nil
func (dw *DcimPlatformDataWrapper) IsValid() bool {
	if dw.Platform != nil && !dw.hasParent && dw.Platform.Name == "" {
		dw.Platform = nil
	}
	return dw.Platform != nil
}

// Normalise normalises the data
func (dw *DcimPlatformDataWrapper) Normalise() {
	if dw.IsValid() && dw.Platform.Tags != nil && len(dw.Platform.Tags) == 0 {
		dw.Platform.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimPlatformDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Platform != nil && dw.hasParent && dw.Platform.Name == "" {
		dw.Platform = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Platform == nil && dw.intended {
		return objects, nil
	}

	if dw.Platform == nil && dw.hasParent {
		dw.Platform = NewDcimPlatform()
		dw.placeholder = true
	}

	if dw.Platform.Slug == "" {
		dw.Platform.Slug = slug.Make(dw.Platform.Name)
	}

	if dw.Platform.Manufacturer != nil {
		manufacturer := DcimManufacturerDataWrapper{Manufacturer: dw.Platform.Manufacturer, BaseDataWrapper: BaseDataWrapper{placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

		mo, err := manufacturer.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, mo...)

		dw.Platform.Manufacturer = manufacturer.Manufacturer
	}

	if dw.Platform.Tags != nil {
		for _, t := range dw.Platform.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimPlatformDataWrapper) DataType() string {
	return DcimPlatformObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimPlatformDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": dw.Platform.Name,
	}
	if dw.Platform.Manufacturer != nil {
		params["manufacturer__name"] = dw.Platform.Manufacturer.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimPlatformDataWrapper) ID() int {
	return dw.Platform.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimPlatformDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimPlatformDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualManufacturerKey := fmt.Sprintf("%p", dw.Platform.Manufacturer)
	actualManufacturer := extractFromObjectsMap(actualNestedObjectsMap, actualManufacturerKey)
	intendedManufacturer := extractFromObjectsMap(intendedNestedObjects, actualManufacturerKey)

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

		dw.Platform.ID = intended.Platform.ID
		dw.Platform.Name = intended.Platform.Name
		dw.Platform.Slug = intended.Platform.Slug

		if actualManufacturer != nil {
			if actualManufacturer.IsPlaceholder() && intended.Platform.Manufacturer != nil {
				intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Platform.Manufacturer))
			}

			manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
			if manufacturerErr != nil {
				return nil, manufacturerErr
			}

			manufacturer, err := copyData(actualManufacturer.Data().(*DcimManufacturer))
			if err != nil {
				return nil, err
			}
			manufacturer.Tags = nil

			if !actualManufacturer.HasChanged() {
				manufacturer = &DcimManufacturer{
					ID: actualManufacturer.ID(),
				}

				intendedManufacturerID := intendedManufacturer.ID()
				if intended.Platform.Manufacturer != nil {
					intendedManufacturerID = intended.Platform.Manufacturer.ID
				}

				intended.Platform.Manufacturer = &DcimManufacturer{
					ID: intendedManufacturerID,
				}
			}

			dw.Platform.Manufacturer = manufacturer

			dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)
		} else {
			if intended.Platform.Manufacturer != nil {
				manufacturerID := intended.Platform.Manufacturer.ID

				dw.Platform.Manufacturer = &DcimManufacturer{
					ID: manufacturerID,
				}
				intended.Platform.Manufacturer = &DcimManufacturer{
					ID: manufacturerID,
				}
			}
		}

		if dw.Platform.Description == nil {
			dw.Platform.Description = intended.Platform.Description
		}

		tagsToMerge := mergeTags(dw.Platform.Tags, intended.Platform.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Platform.Tags = tagsToMerge
		}

		for _, t := range dw.Platform.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		if actualManufacturer != nil {
			manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
			if manufacturerErr != nil {
				return nil, manufacturerErr
			}

			manufacturer, err := copyData(actualManufacturer.Data().(*DcimManufacturer))
			if err != nil {
				return nil, err
			}
			manufacturer.Tags = nil

			if !actualManufacturer.HasChanged() {
				manufacturer = &DcimManufacturer{
					ID: actualManufacturer.ID(),
				}
			}
			dw.Platform.Manufacturer = manufacturer

			dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)
		}

		tagsToMerge := mergeTags(dw.Platform.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Platform.Tags = tagsToMerge
		}

		for _, t := range dw.Platform.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the platform
func (dw *DcimPlatformDataWrapper) SetDefaults() {}

// DcimSiteDataWrapper represents a DCIM site data wrapper
type DcimSiteDataWrapper struct {
	BaseDataWrapper
	Site *DcimSite
}

func (*DcimSiteDataWrapper) comparableData() {}

// Data returns the Site
func (dw *DcimSiteDataWrapper) Data() any {
	return dw.Site
}

// IsValid returns true if the Site is not nil
func (dw *DcimSiteDataWrapper) IsValid() bool {
	if dw.Site != nil && !dw.hasParent && dw.Site.Name == "" {
		dw.Site = nil
	}
	return dw.Site != nil
}

// Normalise normalises the data
func (dw *DcimSiteDataWrapper) Normalise() {
	if dw.IsValid() && dw.Site.Tags != nil && len(dw.Site.Tags) == 0 {
		dw.Site.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimSiteDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Site != nil && dw.hasParent && dw.Site.Name == "" {
		dw.Site = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Site == nil && dw.intended {
		return objects, nil
	}

	if dw.Site == nil && dw.hasParent {
		dw.Site = NewDcimSite()
		dw.placeholder = true
	}

	if dw.Site.Slug == "" {
		dw.Site.Slug = slug.Make(dw.Site.Name)
	}

	if dw.Site.Tags != nil {
		for _, t := range dw.Site.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *DcimSiteDataWrapper) DataType() string {
	return DcimSiteObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *DcimSiteDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.Site.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimSiteDataWrapper) ID() int {
	return dw.Site.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimSiteDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimSiteDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	reconciliationRequired := true

	if intended != nil {
		dw.Site.ID = intended.Site.ID
		dw.Site.Name = intended.Site.Name
		dw.Site.Slug = intended.Site.Slug

		if dw.Site.Status == nil {
			dw.Site.Status = intended.Site.Status
		}

		if dw.Site.Facility == nil {
			dw.Site.Facility = intended.Site.Facility
		}

		if dw.Site.TimeZone == nil {
			dw.Site.TimeZone = intended.Site.TimeZone
		}

		if dw.Site.Description == nil {
			dw.Site.Description = intended.Site.Description
		}

		if dw.Site.Comments == nil {
			dw.Site.Comments = intended.Site.Comments
		}

		tagsToMerge := mergeTags(dw.Site.Tags, intended.Site.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Site.Tags = tagsToMerge
		}

		for _, t := range dw.Site.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		tagsToMerge := mergeTags(dw.Site.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Site.Tags = tagsToMerge
		}

		for _, t := range dw.Site.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the site
func (dw *DcimSiteDataWrapper) SetDefaults() {
	if dw.Site.Status == nil {
		status := DcimSiteStatusActive
		dw.Site.Status = &status
	}
}

// TagDataWrapper represents a tag data wrapper
type TagDataWrapper struct {
	Tag *Tag

	placeholder bool
	hasParent   bool
}

func (*TagDataWrapper) comparableData() {}

// Data returns the Tag
func (dw *TagDataWrapper) Data() any {
	return dw.Tag
}

// IsValid returns true if the Tag is not nil
func (dw *TagDataWrapper) IsValid() bool {
	return dw.Tag != nil
}

// Normalise normalises the data
func (dw *TagDataWrapper) Normalise() {}

// NestedObjects returns all nested objects
func (dw *TagDataWrapper) NestedObjects() ([]ComparableData, error) {
	return nil, nil
}

// DataType returns the data type
func (dw *TagDataWrapper) DataType() string {
	return ExtrasTagObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *TagDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.Tag.Name,
	}
}

// ID returns the ID of the data
func (dw *TagDataWrapper) ID() int {
	return dw.Tag.ID
}

// HasChanged returns true if the data has changed
func (dw *TagDataWrapper) HasChanged() bool {
	return false
}

// IsPlaceholder returns true if the data is a placeholder
func (dw *TagDataWrapper) IsPlaceholder() bool {
	return dw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (dw *TagDataWrapper) Patch(cmp ComparableData, _ map[string]ComparableData) ([]ComparableData, error) {
	d2, ok := cmp.(*TagDataWrapper)
	if !ok && d2 != nil {
		return nil, errors.New("invalid data type")
	}

	return nil, nil
}

// SetDefaults sets the default values for the platform
func (dw *TagDataWrapper) SetDefaults() {}

// NewDataWrapper creates a new data wrapper for the given data type
func NewDataWrapper(dataType string) (ComparableData, error) {
	switch dataType {
	case DcimDeviceObjectType:
		return &DcimDeviceDataWrapper{}, nil
	case DcimDeviceRoleObjectType:
		return &DcimDeviceRoleDataWrapper{}, nil
	case DcimDeviceTypeObjectType:
		return &DcimDeviceTypeDataWrapper{}, nil
	case DcimInterfaceObjectType:
		return &DcimInterfaceDataWrapper{}, nil
	case DcimManufacturerObjectType:
		return &DcimManufacturerDataWrapper{}, nil
	case DcimPlatformObjectType:
		return &DcimPlatformDataWrapper{}, nil
	case DcimSiteObjectType:
		return &DcimSiteDataWrapper{}, nil
	case ExtrasTagObjectType:
		return &TagDataWrapper{}, nil
	case IpamIPAddressObjectType:
		return &IpamIPAddressDataWrapper{}, nil
	case IpamPrefixObjectType:
		return &IpamPrefixDataWrapper{}, nil
	default:
		return nil, fmt.Errorf("unsupported data type %s", dataType)
	}
}

func extractFromObjectsMap(objectsMap map[string]ComparableData, key string) ComparableData {
	if obj, ok := objectsMap[key]; ok {
		return obj
	}
	return nil
}

func dedupObjectsToReconcile(objects []ComparableData) ([]ComparableData, error) {
	hashes := make(map[uint64]struct{})
	dedupedObjectsToReconcile := make([]ComparableData, 0)
	for _, o := range objects {
		hash, err := hashstructure.Hash(o.Data(), hashstructure.FormatV2, nil)
		if err != nil {
			return nil, err
		}
		if _, ok := hashes[hash]; ok {
			continue
		}
		hashes[hash] = struct{}{}
		dedupedObjectsToReconcile = append(dedupedObjectsToReconcile, o)
	}

	return dedupedObjectsToReconcile, nil
}
