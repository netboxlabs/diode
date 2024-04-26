package netbox

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gosimple/slug"
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

	// QueryString returns the query string needed to retrieve its object state
	QueryString() string

	// ID returns the ID of the data
	ID() int

	// IsPlaceholder returns true if the data is a placeholder
	IsPlaceholder() bool

	// SetDefaults sets the default values for the data
	SetDefaults()

	// Patch creates patches between the actual, intended and current data
	Patch(ComparableData, map[string]ComparableData) ([]ComparableData, error)
}

// DcimDeviceDataWrapper represents a DCIM device data wrapper
type DcimDeviceDataWrapper struct {
	Device *DcimDevice

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimDeviceDataWrapper) comparableData() {}

// Data returns the Device
func (actual *DcimDeviceDataWrapper) Data() any {
	return actual.Device
}

// IsValid returns true if the Device is not nil
func (actual *DcimDeviceDataWrapper) IsValid() bool {
	if actual.Device != nil && !actual.hasParent && actual.Device.Name == "" {
		actual.Device = nil
	}
	return actual.Device != nil
}

// Normalise normalises the data
func (actual *DcimDeviceDataWrapper) Normalise() {
	if actual.IsValid() && actual.Device.Tags != nil && len(actual.Device.Tags) == 0 {
		actual.Device.Tags = nil
	}
}

// NestedObjects returns all nested objects
func (actual *DcimDeviceDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.Device != nil && actual.hasParent && actual.Device.Name == "" {
		actual.Device = nil
	}

	objects := make([]ComparableData, 0)

	if actual.Device == nil && actual.hasParent {
		actual.Device = NewDcimDevice()
		actual.placeholder = true
	}

	site := DcimSiteDataWrapper{Site: actual.Device.Site, placeholder: actual.placeholder, hasParent: true}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	actual.Device.Site = site.Site

	platform := DcimPlatformDataWrapper{Platform: actual.Device.Platform, placeholder: actual.placeholder, hasParent: true}

	po, err := platform.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, po...)

	actual.Device.Platform = platform.Platform

	deviceType := DcimDeviceTypeDataWrapper{DeviceType: actual.Device.DeviceType, placeholder: actual.placeholder, hasParent: true}

	dto, err := deviceType.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dto...)

	actual.Device.DeviceType = deviceType.DeviceType

	deviceRole := DcimDeviceRoleDataWrapper{DeviceRole: actual.Device.Role, placeholder: actual.placeholder, hasParent: true}

	dro, err := deviceRole.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, dro...)

	actual.Device.Role = deviceRole.DeviceRole

	if actual.Device.Tags != nil {
		for _, t := range actual.Device.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// DataType returns the data type
func (actual *DcimDeviceDataWrapper) DataType() string {
	return DcimDeviceObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimDeviceDataWrapper) QueryString() string {
	return actual.Device.Name
}

// ID returns the ID of the data
func (actual *DcimDeviceDataWrapper) ID() int {
	return actual.Device.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimDeviceDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimDeviceDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range actual.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", actual.Device.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", actual.Device.Site))

	actualPlatform := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", actual.Device.Platform))
	intendedPlatform := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", actual.Device.Platform))

	actualDeviceType := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", actual.Device.DeviceType))
	intendedDeviceType := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", actual.Device.DeviceType))

	actualRole := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", actual.Device.Role))
	intendedRole := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", actual.Device.Role))

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

		actual.Device.ID = intended.Device.ID
		actual.Device.Name = intended.Device.Name

		if actual.Device.Status == nil {
			actual.Device.Status = intended.Device.Status
		}

		if actual.Device.Description == nil {
			actual.Device.Description = intended.Device.Description
		}

		if actual.Device.Comments == nil {
			actual.Device.Comments = intended.Device.Comments
		}

		if actual.Device.AssetTag == nil {
			actual.Device.AssetTag = intended.Device.AssetTag
		}

		if actualSite.IsPlaceholder() && intended.Device.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Site))
		}

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}
		actual.Device.Site = actualSite.Data().(*DcimSite)

		actual.objectsToReconcile = append(actual.objectsToReconcile, siteObjectsToReconcile...)

		if actualPlatform.IsPlaceholder() && intended.Device.Platform != nil {
			intendedPlatform = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Platform))
		}

		platformObjectsToReconcile, platformErr := actualPlatform.Patch(intendedPlatform, intendedNestedObjects)
		if platformErr != nil {
			return nil, platformErr
		}
		actual.Device.Platform = actualPlatform.Data().(*DcimPlatform)

		actual.objectsToReconcile = append(actual.objectsToReconcile, platformObjectsToReconcile...)

		if actualDeviceType.IsPlaceholder() && intended.Device.DeviceType != nil {
			intendedDeviceType = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.DeviceType))
		}

		deviceTypeObjectsToReconcile, deviceTypeErr := actualDeviceType.Patch(intendedDeviceType, intendedNestedObjects)
		if deviceTypeErr != nil {
			return nil, deviceTypeErr
		}
		actual.Device.DeviceType = actualDeviceType.Data().(*DcimDeviceType)

		actual.objectsToReconcile = append(actual.objectsToReconcile, deviceTypeObjectsToReconcile...)

		if actualRole.IsPlaceholder() && intended.Device.Role != nil {
			intendedRole = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Device.Role))
		}

		roleObjectsToReconcile, roleErr := actualRole.Patch(intendedRole, intendedNestedObjects)
		if roleErr != nil {
			return nil, roleErr
		}
		actual.Device.Role = actualRole.Data().(*DcimDeviceRole)

		actual.objectsToReconcile = append(actual.objectsToReconcile, roleObjectsToReconcile...)

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		actual.SetDefaults()

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}
		actual.Device.Site = actualSite.Data().(*DcimSite)

		actual.objectsToReconcile = append(actual.objectsToReconcile, siteObjectsToReconcile...)

		platformObjectsToReconcile, platformErr := actualPlatform.Patch(intendedPlatform, intendedNestedObjects)
		if platformErr != nil {
			return nil, platformErr
		}
		actual.Device.Platform = actualPlatform.Data().(*DcimPlatform)

		actual.objectsToReconcile = append(actual.objectsToReconcile, platformObjectsToReconcile...)

		deviceTypeObjectsToReconcile, deviceTypeErr := actualDeviceType.Patch(intendedDeviceType, intendedNestedObjects)
		if deviceTypeErr != nil {
			return nil, deviceTypeErr
		}
		actual.Device.DeviceType = actualDeviceType.Data().(*DcimDeviceType)

		actual.objectsToReconcile = append(actual.objectsToReconcile, deviceTypeObjectsToReconcile...)

		roleObjectsToReconcile, roleErr := actualRole.Patch(intendedRole, intendedNestedObjects)
		if roleErr != nil {
			return nil, roleErr
		}
		actual.Device.Role = actualRole.Data().(*DcimDeviceRole)

		actual.objectsToReconcile = append(actual.objectsToReconcile, roleObjectsToReconcile...)
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	dedupObjectsToReconcile, err := dedupObjectsToReconcile(actual.objectsToReconcile)
	if err != nil {
		return nil, err
	}
	actual.objectsToReconcile = dedupObjectsToReconcile

	return actual.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device
func (actual *DcimDeviceDataWrapper) SetDefaults() {
	if actual.Device.Status == nil {
		status := DcimDeviceStatusActive
		actual.Device.Status = &status
	}
}

// DcimDeviceRoleDataWrapper represents a DCIM device role data wrapper
type DcimDeviceRoleDataWrapper struct {
	DeviceRole *DcimDeviceRole

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimDeviceRoleDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (actual *DcimDeviceRoleDataWrapper) Data() any {
	return actual.DeviceRole
}

// IsValid returns true if the DeviceRole is not nil
func (actual *DcimDeviceRoleDataWrapper) IsValid() bool {
	if actual.DeviceRole != nil && !actual.hasParent && actual.DeviceRole.Name == "" {
		actual.DeviceRole = nil
	}
	return actual.DeviceRole != nil
}

// Normalise normalises the data
func (actual *DcimDeviceRoleDataWrapper) Normalise() {
	if actual.IsValid() && actual.DeviceRole.Tags != nil && len(actual.DeviceRole.Tags) == 0 {
		actual.DeviceRole.Tags = nil
	}
}

// NestedObjects returns all nested objects
func (actual *DcimDeviceRoleDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.DeviceRole != nil && actual.hasParent && actual.DeviceRole.Name == "" {
		actual.DeviceRole = nil
	}

	objects := make([]ComparableData, 0)

	if actual.DeviceRole == nil && actual.hasParent {
		actual.DeviceRole = NewDcimDeviceRole()
		actual.placeholder = true
	}

	if actual.DeviceRole.Slug == "" {
		actual.DeviceRole.Slug = slug.Make(actual.DeviceRole.Name)
	}

	if actual.DeviceRole.Tags != nil {
		for _, t := range actual.DeviceRole.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// DataType returns the data type
func (actual *DcimDeviceRoleDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimDeviceRoleDataWrapper) QueryString() string {
	return actual.DeviceRole.Name
}

// ID returns the ID of the data
func (actual *DcimDeviceRoleDataWrapper) ID() int {
	return actual.DeviceRole.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimDeviceRoleDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimDeviceRoleDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceRoleDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range actual.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	reconciliationRequired := true

	if intended != nil {
		actual.DeviceRole.ID = intended.DeviceRole.ID
		actual.DeviceRole.Name = intended.DeviceRole.Name
		actual.DeviceRole.Slug = intended.DeviceRole.Slug

		if actual.IsPlaceholder() || actual.DeviceRole.Color == nil {
			actual.DeviceRole.Color = intended.DeviceRole.Color
		}

		if actual.DeviceRole.Description == nil {
			actual.DeviceRole.Description = intended.DeviceRole.Description
		}

		tagsToMerge := mergeTags(actual.DeviceRole.Tags, intended.DeviceRole.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.DeviceRole.Tags = tagsToMerge
		}

		for _, t := range actual.DeviceRole.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		actual.SetDefaults()

		tagsToMerge := mergeTags(actual.DeviceRole.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.DeviceRole.Tags = tagsToMerge
		}

		for _, t := range actual.DeviceRole.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	return actual.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device role
func (actual *DcimDeviceRoleDataWrapper) SetDefaults() {
	if actual.DeviceRole.Color == nil {
		color := "000000"
		actual.DeviceRole.Color = &color
	}
}

// DcimDeviceTypeDataWrapper represents a DCIM device type data wrapper
type DcimDeviceTypeDataWrapper struct {
	DeviceType *DcimDeviceType

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimDeviceTypeDataWrapper) comparableData() {}

// Data returns the DeviceType
func (actual *DcimDeviceTypeDataWrapper) Data() any {
	return actual.DeviceType
}

// IsValid returns true if the DeviceType is not nil
func (actual *DcimDeviceTypeDataWrapper) IsValid() bool {
	if actual.DeviceType != nil && !actual.hasParent && actual.DeviceType.Model == "" {
		actual.DeviceType = nil
	}
	return actual.DeviceType != nil
}

// Normalise normalises the data
func (actual *DcimDeviceTypeDataWrapper) Normalise() {
	if actual.IsValid() && actual.DeviceType.Tags != nil && len(actual.DeviceType.Tags) == 0 {
		actual.DeviceType.Tags = nil
	}
}

// DataType returns the data type
func (actual *DcimDeviceTypeDataWrapper) DataType() string {
	return DcimDeviceTypeObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimDeviceTypeDataWrapper) QueryString() string {
	return actual.DeviceType.Model
}

// ID returns the ID of the data
func (actual *DcimDeviceTypeDataWrapper) ID() int {
	return actual.DeviceType.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimDeviceTypeDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// NestedObjects returns all nested objects
func (actual *DcimDeviceTypeDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.DeviceType != nil && actual.hasParent && actual.DeviceType.Model == "" {
		actual.DeviceType = nil
	}

	objects := make([]ComparableData, 0)

	if actual.DeviceType == nil && actual.hasParent {
		actual.DeviceType = NewDcimDeviceType()
		actual.placeholder = true
	}

	if actual.DeviceType.Slug == "" {
		actual.DeviceType.Slug = slug.Make(actual.DeviceType.Model)
	}

	manufacturer := DcimManufacturerDataWrapper{Manufacturer: actual.DeviceType.Manufacturer, placeholder: actual.placeholder, hasParent: true}

	mo, err := manufacturer.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, mo...)

	actual.DeviceType.Manufacturer = manufacturer.Manufacturer

	if actual.DeviceType.Tags != nil && len(actual.DeviceType.Tags) == 0 {
		actual.DeviceType.Tags = nil
	}

	if actual.DeviceType.Tags != nil {
		for _, t := range actual.DeviceType.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimDeviceTypeDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimDeviceTypeDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range actual.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualManufacturerKey := fmt.Sprintf("%p", actual.DeviceType.Manufacturer)
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

		actual.DeviceType.ID = intended.DeviceType.ID
		actual.DeviceType.Model = intended.DeviceType.Model
		actual.DeviceType.Slug = intended.DeviceType.Slug

		if actual.DeviceType.Description == nil {
			actual.DeviceType.Description = intended.DeviceType.Description
		}

		if actual.DeviceType.Comments == nil {
			actual.DeviceType.Comments = intended.DeviceType.Comments
		}

		if actual.DeviceType.PartNumber == nil {
			actual.DeviceType.PartNumber = intended.DeviceType.PartNumber
		}

		if actualManufacturer.IsPlaceholder() && intended.DeviceType.Manufacturer != nil {
			intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.DeviceType.Manufacturer))
		}

		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}
		actual.DeviceType.Manufacturer = actualManufacturer.Data().(*DcimManufacturer)

		tagsToMerge := mergeTags(actual.DeviceType.Tags, intended.DeviceType.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.DeviceType.Tags = tagsToMerge
		}

		for _, t := range actual.DeviceType.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actual.objectsToReconcile = append(actual.objectsToReconcile, manufacturerObjectsToReconcile...)

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}
		actual.DeviceType.Manufacturer = actualManufacturer.Data().(*DcimManufacturer)

		tagsToMerge := mergeTags(actual.DeviceType.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.DeviceType.Tags = tagsToMerge
		}

		for _, t := range actual.DeviceType.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actual.objectsToReconcile = append(actual.objectsToReconcile, manufacturerObjectsToReconcile...)
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	return actual.objectsToReconcile, nil
}

// SetDefaults sets the default values for the device type
func (actual *DcimDeviceTypeDataWrapper) SetDefaults() {}

// DcimInterfaceDataWrapper represents a DCIM interface data wrapper
type DcimInterfaceDataWrapper struct {
	Interface *DcimInterface

	placeholder bool
}

func (*DcimInterfaceDataWrapper) comparableData() {}

// Data returns the Interface
func (d *DcimInterfaceDataWrapper) Data() any {
	return d.Interface
}

// IsValid returns true if the Interface is not nil
func (d *DcimInterfaceDataWrapper) IsValid() bool {
	return d.Interface != nil
}

// Normalise normalises the data
func (d *DcimInterfaceDataWrapper) Normalise() {}

// NestedObjects returns all nested objects
func (d *DcimInterfaceDataWrapper) NestedObjects() ([]ComparableData, error) {
	return nil, nil
}

// DataType returns the data type
func (d *DcimInterfaceDataWrapper) DataType() string {
	return DcimInterfaceObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimInterfaceDataWrapper) QueryString() string {
	return d.Interface.Name
}

// ID returns the ID of the data
func (d *DcimInterfaceDataWrapper) ID() int {
	return d.Interface.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimInterfaceDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// Patch creates patches between the actual, intended and current data
func (d *DcimInterfaceDataWrapper) Patch(cmp ComparableData, _ map[string]ComparableData) ([]ComparableData, error) {
	d2, ok := cmp.(*DcimInterfaceDataWrapper)
	if !ok && d2 != nil {
		return nil, errors.New("invalid data type")
	}

	fmt.Printf("d: %#v\n", d)
	fmt.Printf("d2: %#v\n", d2)

	return nil, nil
}

// SetDefaults sets the default values for the interface
func (d *DcimInterfaceDataWrapper) SetDefaults() {}

// DcimManufacturerDataWrapper represents a DCIM manufacturer data wrapper
type DcimManufacturerDataWrapper struct {
	Manufacturer *DcimManufacturer

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimManufacturerDataWrapper) comparableData() {}

// Data returns the Manufacturer
func (actual *DcimManufacturerDataWrapper) Data() any {
	return actual.Manufacturer
}

// IsValid returns true if the Manufacturer is not nil
func (actual *DcimManufacturerDataWrapper) IsValid() bool {
	if actual.Manufacturer != nil && !actual.hasParent && actual.Manufacturer.Name == "" {
		actual.Manufacturer = nil
	}
	return actual.Manufacturer != nil
}

// Normalise normalises the data
func (actual *DcimManufacturerDataWrapper) Normalise() {
	if actual.IsValid() && actual.Manufacturer.Tags != nil && len(actual.Manufacturer.Tags) == 0 {
		actual.Manufacturer.Tags = nil
	}
}

// NestedObjects returns all nested objects
func (actual *DcimManufacturerDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.Manufacturer != nil && actual.hasParent && actual.Manufacturer.Name == "" {
		actual.Manufacturer = nil
	}

	objects := make([]ComparableData, 0)

	if actual.Manufacturer == nil && actual.hasParent {
		actual.Manufacturer = NewDcimManufacturer()
		actual.placeholder = true
	}

	if actual.Manufacturer.Slug == "" {
		actual.Manufacturer.Slug = slug.Make(actual.Manufacturer.Name)
	}

	if actual.Manufacturer.Tags != nil {
		for _, t := range actual.Manufacturer.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// DataType returns the data type
func (actual *DcimManufacturerDataWrapper) DataType() string {
	return DcimManufacturerObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimManufacturerDataWrapper) QueryString() string {
	return actual.Manufacturer.Name
}

// ID returns the ID of the data
func (actual *DcimManufacturerDataWrapper) ID() int {
	return actual.Manufacturer.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimManufacturerDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimManufacturerDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimManufacturerDataWrapper)

	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	reconciliationRequired := true

	if intended != nil {
		actual.Manufacturer.ID = intended.Manufacturer.ID
		actual.Manufacturer.Name = intended.Manufacturer.Name
		actual.Manufacturer.Slug = intended.Manufacturer.Slug

		if actual.Manufacturer.Description == nil {
			actual.Manufacturer.Description = intended.Manufacturer.Description
		}

		tagsToMerge := mergeTags(actual.Manufacturer.Tags, intended.Manufacturer.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Manufacturer.Tags = tagsToMerge
		}

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		tagsToMerge := mergeTags(actual.Manufacturer.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Manufacturer.Tags = tagsToMerge
		}
	}

	for _, t := range actual.Manufacturer.Tags {
		if t.ID == 0 {
			actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	return actual.objectsToReconcile, nil
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
func (actual *DcimManufacturerDataWrapper) SetDefaults() {}

// DcimPlatformDataWrapper represents a DCIM platform data wrapper
type DcimPlatformDataWrapper struct {
	Platform *DcimPlatform

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimPlatformDataWrapper) comparableData() {}

// Data returns the Platform
func (actual *DcimPlatformDataWrapper) Data() any {
	return actual.Platform
}

// IsValid returns true if the Platform is not nil
func (actual *DcimPlatformDataWrapper) IsValid() bool {
	if actual.Platform != nil && !actual.hasParent && actual.Platform.Name == "" {
		actual.Platform = nil
	}
	return actual.Platform != nil
}

// Normalise normalises the data
func (actual *DcimPlatformDataWrapper) Normalise() {
	if actual.IsValid() && actual.Platform.Tags != nil && len(actual.Platform.Tags) == 0 {
		actual.Platform.Tags = nil
	}
}

// NestedObjects returns all nested objects
func (actual *DcimPlatformDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.Platform != nil && actual.hasParent && actual.Platform.Name == "" {
		actual.Platform = nil
	}

	objects := make([]ComparableData, 0)

	if actual.Platform == nil && actual.hasParent {
		actual.Platform = NewDcimPlatform()
		actual.placeholder = true
	}

	if actual.Platform.Slug == "" {
		actual.Platform.Slug = slug.Make(actual.Platform.Name)
	}

	manufacturer := DcimManufacturerDataWrapper{Manufacturer: actual.Platform.Manufacturer, placeholder: actual.placeholder, hasParent: true}

	mo, err := manufacturer.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, mo...)

	actual.Platform.Manufacturer = manufacturer.Manufacturer

	if actual.Platform.Tags != nil {
		for _, t := range actual.Platform.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// DataType returns the data type
func (actual *DcimPlatformDataWrapper) DataType() string {
	return DcimPlatformObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimPlatformDataWrapper) QueryString() string {
	return actual.Platform.Name
}

// ID returns the ID of the data
func (actual *DcimPlatformDataWrapper) ID() int {
	return actual.Platform.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimPlatformDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimPlatformDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimPlatformDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range actual.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualManufacturerKey := fmt.Sprintf("%p", actual.Platform.Manufacturer)
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

		actual.Platform.ID = intended.Platform.ID
		actual.Platform.Name = intended.Platform.Name
		actual.Platform.Slug = intended.Platform.Slug

		if actualManufacturer.IsPlaceholder() && intended.Platform.Manufacturer != nil {
			intendedManufacturer = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Platform.Manufacturer))
		}

		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}
		actual.Platform.Manufacturer = actualManufacturer.Data().(*DcimManufacturer)

		actual.objectsToReconcile = append(actual.objectsToReconcile, manufacturerObjectsToReconcile...)

		if actual.Platform.Description == nil {
			actual.Platform.Description = intended.Platform.Description
		}

		tagsToMerge := mergeTags(actual.Platform.Tags, intended.Platform.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Platform.Tags = tagsToMerge
		}

		for _, t := range actual.Platform.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		manufacturerObjectsToReconcile, manufacturerErr := actualManufacturer.Patch(intendedManufacturer, intendedNestedObjects)
		if manufacturerErr != nil {
			return nil, manufacturerErr
		}
		actual.Platform.Manufacturer = actualManufacturer.Data().(*DcimManufacturer)

		tagsToMerge := mergeTags(actual.Platform.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Platform.Tags = tagsToMerge
		}

		for _, t := range actual.Platform.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actual.objectsToReconcile = append(actual.objectsToReconcile, manufacturerObjectsToReconcile...)
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	return actual.objectsToReconcile, nil
}

// SetDefaults sets the default values for the platform
func (actual *DcimPlatformDataWrapper) SetDefaults() {}

// DcimSiteDataWrapper represents a DCIM site data wrapper
type DcimSiteDataWrapper struct {
	Site *DcimSite

	placeholder        bool
	hasParent          bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*DcimSiteDataWrapper) comparableData() {}

// Data returns the Site
func (actual *DcimSiteDataWrapper) Data() any {
	return actual.Site
}

// IsValid returns true if the Site is not nil
func (actual *DcimSiteDataWrapper) IsValid() bool {
	if actual.Site != nil && !actual.hasParent && actual.Site.Name == "" {
		actual.Site = nil
	}
	return actual.Site != nil
}

// Normalise normalises the data
func (actual *DcimSiteDataWrapper) Normalise() {
	if actual.IsValid() && actual.Site.Tags != nil && len(actual.Site.Tags) == 0 {
		actual.Site.Tags = nil
	}
}

// NestedObjects returns all nested objects
func (actual *DcimSiteDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(actual.nestedObjects) > 0 {
		return actual.nestedObjects, nil
	}

	if actual.Site != nil && actual.hasParent && actual.Site.Name == "" {
		actual.Site = nil
	}

	objects := make([]ComparableData, 0)

	if actual.Site == nil && actual.hasParent {
		actual.Site = NewDcimSite()
		actual.placeholder = true
	}

	if actual.Site.Slug == "" {
		actual.Site.Slug = slug.Make(actual.Site.Name)
	}

	if actual.Site.Tags != nil {
		for _, t := range actual.Site.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	actual.nestedObjects = objects

	objects = append(objects, actual)

	return objects, nil
}

// DataType returns the data type
func (actual *DcimSiteDataWrapper) DataType() string {
	return DcimSiteObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (actual *DcimSiteDataWrapper) QueryString() string {
	return actual.Site.Name
}

// ID returns the ID of the data
func (actual *DcimSiteDataWrapper) ID() int {
	return actual.Site.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (actual *DcimSiteDataWrapper) IsPlaceholder() bool {
	return actual.placeholder
}

// Patch creates patches between the actual, intended and current data
func (actual *DcimSiteDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*DcimSiteDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range actual.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	reconciliationRequired := true

	if intended != nil {
		actual.Site.ID = intended.Site.ID
		actual.Site.Name = intended.Site.Name
		actual.Site.Slug = intended.Site.Slug

		if actual.Site.Status == nil {
			actual.Site.Status = intended.Site.Status
		}

		if actual.Site.Facility == nil {
			actual.Site.Facility = intended.Site.Facility
		}

		if actual.Site.TimeZone == nil {
			actual.Site.TimeZone = intended.Site.TimeZone
		}

		if actual.Site.Description == nil {
			actual.Site.Description = intended.Site.Description
		}

		if actual.Site.Comments == nil {
			actual.Site.Comments = intended.Site.Comments
		}

		tagsToMerge := mergeTags(actual.Site.Tags, intended.Site.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Site.Tags = tagsToMerge
		}

		for _, t := range actual.Site.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(actual.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		actual.SetDefaults()

		tagsToMerge := mergeTags(actual.Site.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			actual.Site.Tags = tagsToMerge
		}

		for _, t := range actual.Site.Tags {
			if t.ID == 0 {
				actual.objectsToReconcile = append(actual.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		actual.objectsToReconcile = append(actual.objectsToReconcile, actual)
	}

	return actual.objectsToReconcile, nil
}

// SetDefaults sets the default values for the site
func (actual *DcimSiteDataWrapper) SetDefaults() {
	if actual.Site.Status == nil {
		status := DcimSiteStatusActive
		actual.Site.Status = &status
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
func (d *TagDataWrapper) Data() any {
	return d.Tag
}

// IsValid returns true if the Tag is not nil
func (d *TagDataWrapper) IsValid() bool {
	return d.Tag != nil
}

// Normalise normalises the data
func (d *TagDataWrapper) Normalise() {}

// NestedObjects returns all nested objects
func (d *TagDataWrapper) NestedObjects() ([]ComparableData, error) {
	return nil, nil
}

// DataType returns the data type
func (d *TagDataWrapper) DataType() string {
	return ExtrasTagObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *TagDataWrapper) QueryString() string {
	return d.Tag.Name
}

// ID returns the ID of the data
func (d *TagDataWrapper) ID() int {
	return d.Tag.ID
}

// IsPlaceholder returns true if the data is a placeholder
func (d *TagDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// Patch creates patches between the actual, intended and current data
func (d *TagDataWrapper) Patch(cmp ComparableData, _ map[string]ComparableData) ([]ComparableData, error) {
	d2, ok := cmp.(*TagDataWrapper)
	if !ok && d2 != nil {
		return nil, errors.New("invalid data type")
	}

	fmt.Printf("d: %#v\n", d)
	fmt.Printf("d2: %#v\n", d2)

	return nil, nil
}

// SetDefaults sets the default values for the platform
func (d *TagDataWrapper) SetDefaults() {}

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
