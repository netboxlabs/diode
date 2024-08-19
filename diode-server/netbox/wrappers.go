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
type BaseDataWrapper[T any] struct {
	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
	Field              *T
}

// comparableData is a no-op method to satisfy the ComparableData interface
func (bw *BaseDataWrapper[T]) comparableData() {}

// Data returns the Field
func (bw *BaseDataWrapper[T]) Data() any {
	return bw.Field
}

// IsPlaceholder returns true if the data is a placeholder
func (bw *BaseDataWrapper[T]) IsPlaceholder() bool {
	return bw.placeholder
}

// HasChanged returns true if the data has changed
func (bw *BaseDataWrapper[T]) HasChanged() bool {
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
	BaseDataWrapper[DcimDevice]
}

// IsValid returns true if the Device is not nil
func (dw *DcimDeviceDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimDeviceDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimDevice()
		dw.placeholder = true
	}

	// Ignore primary IP addresses for time being
	dw.Field.PrimaryIPv4 = nil
	dw.Field.PrimaryIPv6 = nil

	site := DcimSiteDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimSite]{Field: dw.Field.Site, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	dw.Field.Site = site.Field

	if dw.Field.Platform != nil {
		platform := DcimPlatformDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimPlatform]{Field: dw.Field.Platform, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

		po, err := platform.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, po...)

		dw.Field.Platform = platform.Field
	}

	deviceType := DcimDeviceTypeDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimDeviceType]{Field: dw.Field.DeviceType, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	dto, err := deviceType.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dto...)

	dw.Field.DeviceType = deviceType.Field

	deviceRole := DcimDeviceRoleDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimDeviceRole]{Field: dw.Field.Role, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	dro, err := deviceRole.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dro...)

	dw.Field.Role = deviceRole.Field

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
	if dw.Field.Site != nil {
		params["site__name"] = dw.Field.Site.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimDeviceDataWrapper) ID() int {
	return dw.Field.ID
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

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.Site))

	actualPlatform := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.Platform))
	intendedPlatform := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.Platform))

	actualDeviceType := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.DeviceType))
	intendedDeviceType := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.DeviceType))

	actualRole := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.Role))
	intendedRole := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.Role))

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

		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name

		if dw.Field.Status == nil {
			dw.Field.Status = intended.Field.Status
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.Comments == nil {
			dw.Field.Comments = intended.Field.Comments
		}

		if dw.Field.AssetTag == nil {
			dw.Field.AssetTag = intended.Field.AssetTag
		}

		if dw.Field.Serial == nil {
			dw.Field.Serial = intended.Field.Serial
		}

		if actualSite.IsPlaceholder() && intended.Field.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Site))
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
			if intended.Field.Site != nil {
				intendedSiteID = intended.Field.Site.ID
			}

			intended.Field.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		dw.Field.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		if actualPlatform != nil {
			if actualPlatform.IsPlaceholder() && intended.Field.Platform != nil {
				intendedPlatform = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Platform))
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
				if intended.Field.Platform != nil {
					intendedPlatformID = intended.Field.Platform.ID
				}

				intended.Field.Platform = &DcimPlatform{
					ID: intendedPlatformID,
				}
			}

			dw.Field.Platform = platform

			dw.objectsToReconcile = append(dw.objectsToReconcile, platformObjectsToReconcile...)
		} else {
			if intended.Field.Platform != nil {
				platformID := intended.Field.Platform.ID
				dw.Field.Platform = &DcimPlatform{
					ID: platformID,
				}
				intended.Field.Platform = &DcimPlatform{
					ID: platformID,
				}
			}
		}

		if actualDeviceType.IsPlaceholder() && intended.Field.DeviceType != nil {
			intendedDeviceType = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.DeviceType))
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
			if intended.Field.DeviceType != nil {
				intendedDeviceTypeID = intended.Field.DeviceType.ID
			}

			intended.Field.DeviceType = &DcimDeviceType{
				ID: intendedDeviceTypeID,
			}
		}

		dw.Field.DeviceType = deviceType

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceTypeObjectsToReconcile...)

		if actualRole.IsPlaceholder() && intended.Field.Role != nil {
			intendedRole = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Role))
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
			if intended.Field.Role != nil {
				intendedRoleID = intended.Field.Role.ID
			}

			intended.Field.Role = &DcimDeviceRole{
				ID: intendedRoleID,
			}
		}

		dw.Field.Role = role

		dw.objectsToReconcile = append(dw.objectsToReconcile, roleObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
		dw.Field.Site = site

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
			dw.Field.Platform = platform

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
		dw.Field.DeviceType = deviceType

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
		dw.Field.Role = role

		dw.objectsToReconcile = append(dw.objectsToReconcile, roleObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Status == nil {
		status := DcimDeviceStatusActive
		dw.Field.Status = &status
	}
}

// DcimDeviceRoleDataWrapper represents a DCIM device role data wrapper
type DcimDeviceRoleDataWrapper struct {
	BaseDataWrapper[DcimDeviceRole]
}

// IsValid returns true if the DeviceRole is not nil
func (dw *DcimDeviceRoleDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimDeviceRoleDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceRoleDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimDeviceRole()
		dw.placeholder = true
	}

	if dw.Field.Slug == "" {
		dw.Field.Slug = slug.Make(dw.Field.Name)
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimDeviceRoleDataWrapper) ID() int {
	return dw.Field.ID
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
		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name
		dw.Field.Slug = intended.Field.Slug

		if dw.IsPlaceholder() || dw.Field.Color == nil {
			dw.Field.Color = intended.Field.Color
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Color == nil {
		color := "000000"
		dw.Field.Color = &color
	}
}

// DcimDeviceTypeDataWrapper represents a DCIM device type data wrapper
type DcimDeviceTypeDataWrapper struct {
	BaseDataWrapper[DcimDeviceType]
}

// IsValid returns true if the DeviceType is not nil
func (dw *DcimDeviceTypeDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Model == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimDeviceTypeDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
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
		"q": dw.Field.Model,
	}
	if dw.Field.Manufacturer != nil {
		params["manufacturer__name"] = dw.Field.Manufacturer.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimDeviceTypeDataWrapper) ID() int {
	return dw.Field.ID
}

// NestedObjects returns all nested objects
func (dw *DcimDeviceTypeDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Model == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimDeviceType()
		dw.placeholder = true
	}

	if dw.Field.Slug == "" {
		dw.Field.Slug = slug.Make(dw.Field.Model)
	}

	manufacturer := DcimManufacturerDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimManufacturer]{Field: dw.Field.Manufacturer, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	mo, err := manufacturer.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, mo...)

	dw.Field.Manufacturer = manufacturer.Field

	if dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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

	actualManufacturerKey := fmt.Sprintf("%p", dw.Field.Manufacturer)
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

		dw.Field.ID = intended.Field.ID
		dw.Field.Model = intended.Field.Model
		dw.Field.Slug = intended.Field.Slug

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.Comments == nil {
			dw.Field.Comments = intended.Field.Comments
		}

		if dw.Field.PartNumber == nil {
			dw.Field.PartNumber = intended.Field.PartNumber
		}

		if actualManufacturer.IsPlaceholder() && intended.Field.Manufacturer != nil {
			intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Manufacturer))
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
			if intended.Field.Manufacturer != nil {
				intendedManufacturerID = intended.Field.Manufacturer.ID
			}

			intended.Field.Manufacturer = &DcimManufacturer{
				ID: intendedManufacturerID,
			}
		}

		dw.Field.Manufacturer = manufacturer

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
		dw.Field.Manufacturer = manufacturer

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	BaseDataWrapper[DcimInterface]
}

// IsValid returns true if the Interface is not nil
func (dw *DcimInterfaceDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	if dw.Field != nil {
		if err := dw.Field.Validate(); err != nil {
			return false
		}
	}

	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimInterfaceDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimInterfaceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimInterface()
		dw.placeholder = true
	}

	device := DcimDeviceDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimDevice]{Field: dw.Field.Device, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	do, err := device.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, do...)

	dw.Field.Device = device.Field

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
	if dw.Field.Device != nil {
		params["device__name"] = dw.Field.Device.Name

		if dw.Field.Device.Site != nil {
			params["device__site__name"] = dw.Field.Device.Site.Name
		}
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimInterfaceDataWrapper) ID() int {
	return dw.Field.ID
}

func (dw *DcimInterfaceDataWrapper) hash() string {
	var deviceName, siteName string
	if dw.Field.Device != nil {
		deviceName = dw.Field.Device.Name
		if dw.Field.Device.Site != nil {
			siteName = dw.Field.Device.Site.Name
		}
	}
	return slug.Make(fmt.Sprintf("%s-%s-%s", dw.Field.Name, deviceName, siteName))
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

	actualDevice := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.Device))
	intendedDevice := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.Device))

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

		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name

		if actualDevice.IsPlaceholder() && intended.Field.Device != nil {
			intendedDevice = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Device))
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
			if intended.Field.Device != nil {
				intendedDeviceID = intended.Field.Device.ID
			}

			intended.Field.Device = &DcimDevice{
				ID: intendedDeviceID,
			}
		}

		dw.Field.Device = device

		dw.objectsToReconcile = append(dw.objectsToReconcile, deviceObjectsToReconcile...)

		if dw.Field.Label == nil {
			dw.Field.Label = intended.Field.Label
		}

		if dw.Field.Type == nil {
			dw.Field.Type = intended.Field.Type
		}

		if dw.Field.Enabled == nil {
			dw.Field.Enabled = intended.Field.Enabled
		}

		if dw.Field.MTU == nil {
			dw.Field.MTU = intended.Field.MTU
		}

		if dw.Field.MACAddress == nil {
			dw.Field.MACAddress = intended.Field.MACAddress
		}

		if dw.Field.Speed == nil {
			dw.Field.Speed = intended.Field.Speed
		}

		if dw.Field.WWN == nil {
			dw.Field.WWN = intended.Field.WWN
		}

		if dw.Field.MgmtOnly == nil {
			dw.Field.MgmtOnly = intended.Field.MgmtOnly
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.MarkConnected == nil {
			dw.Field.MarkConnected = intended.Field.MarkConnected
		}

		if dw.Field.Mode == nil {
			dw.Field.Mode = intended.Field.Mode
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
		dw.Field.Device = device

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Type == nil {
		dw.Field.Type = &DefaultInterfaceType
	}
}

// DcimManufacturerDataWrapper represents a DCIM manufacturer data wrapper
type DcimManufacturerDataWrapper struct {
	BaseDataWrapper[DcimManufacturer]
}

// IsValid returns true if the Manufacturer is not nil
func (dw *DcimManufacturerDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimManufacturerDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimManufacturerDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimManufacturer()
		dw.placeholder = true
	}

	if dw.Field.Slug == "" {
		dw.Field.Slug = slug.Make(dw.Field.Name)
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimManufacturerDataWrapper) ID() int {
	return dw.Field.ID
}

// Patch creates patches between the actual, intended and current data
func (dw *DcimManufacturerDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimManufacturerDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name
		dw.Field.Slug = intended.Field.Slug

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}
	}

	for _, t := range dw.Field.Tags {
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
	BaseDataWrapper[DcimPlatform]
}

// IsValid returns true if the Platform is not nil
func (dw *DcimPlatformDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimPlatformDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimPlatformDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimPlatform()
		dw.placeholder = true
	}

	if dw.Field.Slug == "" {
		dw.Field.Slug = slug.Make(dw.Field.Name)
	}

	if dw.Field.Manufacturer != nil {
		manufacturer := DcimManufacturerDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimManufacturer]{Field: dw.Field.Manufacturer, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

		mo, err := manufacturer.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, mo...)

		dw.Field.Manufacturer = manufacturer.Field
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
	if dw.Field.Manufacturer != nil {
		params["manufacturer__name"] = dw.Field.Manufacturer.Name
	}
	return params
}

// ID returns the ID of the data
func (dw *DcimPlatformDataWrapper) ID() int {
	return dw.Field.ID
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

	actualManufacturerKey := fmt.Sprintf("%p", dw.Field.Manufacturer)
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

		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name
		dw.Field.Slug = intended.Field.Slug

		if actualManufacturer != nil {
			if actualManufacturer.IsPlaceholder() && intended.Field.Manufacturer != nil {
				intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Manufacturer))
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
				if intended.Field.Manufacturer != nil {
					intendedManufacturerID = intended.Field.Manufacturer.ID
				}

				intended.Field.Manufacturer = &DcimManufacturer{
					ID: intendedManufacturerID,
				}
			}

			dw.Field.Manufacturer = manufacturer

			dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)
		} else {
			if intended.Field.Manufacturer != nil {
				manufacturerID := intended.Field.Manufacturer.ID

				dw.Field.Manufacturer = &DcimManufacturer{
					ID: manufacturerID,
				}
				intended.Field.Manufacturer = &DcimManufacturer{
					ID: manufacturerID,
				}
			}
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
			dw.Field.Manufacturer = manufacturer

			dw.objectsToReconcile = append(dw.objectsToReconcile, manufacturerObjectsToReconcile...)
		}

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	BaseDataWrapper[DcimSite]
}

// IsValid returns true if the Site is not nil
func (dw *DcimSiteDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}
	return dw.Field != nil
}

// Normalise normalises the data
func (dw *DcimSiteDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *DcimSiteDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Name == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.Field = NewDcimSite()
		dw.placeholder = true
	}

	if dw.Field.Slug == "" {
		dw.Field.Slug = slug.Make(dw.Field.Name)
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Name,
	}
}

// ID returns the ID of the data
func (dw *DcimSiteDataWrapper) ID() int {
	return dw.Field.ID
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
		dw.Field.ID = intended.Field.ID
		dw.Field.Name = intended.Field.Name
		dw.Field.Slug = intended.Field.Slug

		if dw.Field.Status == nil {
			dw.Field.Status = intended.Field.Status
		}

		if dw.Field.Facility == nil {
			dw.Field.Facility = intended.Field.Facility
		}

		if dw.Field.TimeZone == nil {
			dw.Field.TimeZone = intended.Field.TimeZone
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.Comments == nil {
			dw.Field.Comments = intended.Field.Comments
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Status == nil {
		status := DcimSiteStatusActive
		dw.Field.Status = &status
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
