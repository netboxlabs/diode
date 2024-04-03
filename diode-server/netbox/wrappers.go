package netbox

import (
	"fmt"

	"github.com/gosimple/slug"
)

// ComparableData is an interface for NetBox comparable data
type ComparableData interface {
	comparableData()

	// Data returns the data
	Data() any

	// IsValid checks if the data is not nil
	IsValid() bool

	// AllData returns all the data
	AllData() []ComparableData

	// DataType returns the data type
	DataType() string

	// QueryString returns the query string needed to retrieve its object state
	QueryString() string

	// ID returns the ID of the data
	ID() int

	// ReplaceData replaces the data with the given data
	ReplaceData(data ComparableData)

	// IsPlaceholder returns true if the data is a placeholder
	IsPlaceholder() bool
}

// DcimDeviceDataWrapper represents a DCIM device data wrapper
type DcimDeviceDataWrapper struct {
	Device *DcimDevice

	placeholder bool
}

func (*DcimDeviceDataWrapper) comparableData() {}

// Data returns the Device
func (d *DcimDeviceDataWrapper) Data() any {
	return d.Device
}

// IsValid returns true if the Device is not nil
func (d *DcimDeviceDataWrapper) IsValid() bool {
	return d.Device != nil
}

// AllData returns all the data
func (d *DcimDeviceDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	for _, t := range NestedTypes(d.DataType()) {
		switch t {
		case DcimSiteObjectType:
			site := DcimSiteDataWrapper{Site: d.Device.Site, hasParent: true}

			siteData := site.AllData()

			d.Device.Site = site.Site

			result = append(result, siteData...)
		case DcimPlatformObjectType:
			platform := DcimPlatformDataWrapper{Platform: d.Device.Platform, hasParent: true}

			platformData := platform.AllData()

			d.Device.Platform = platform.Platform

			result = append(result, platformData...)
		case DcimDeviceTypeObjectType:
			deviceType := DcimDeviceTypeDataWrapper{DeviceType: d.Device.DeviceType, hasParent: true}

			deviceTypeData := deviceType.AllData()

			d.Device.DeviceType = deviceType.DeviceType

			result = append(result, deviceTypeData...)

		case DcimDeviceRoleObjectType:
			deviceRole := DcimDeviceRoleDataWrapper{DeviceRole: d.Device.Role, hasParent: true}

			deviceRoleData := deviceRole.AllData()

			d.Device.Role = deviceRole.DeviceRole

			result = append(result, deviceRoleData...)
		}
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimDeviceDataWrapper) DataType() string {
	return DcimDeviceObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimDeviceDataWrapper) QueryString() string {
	return d.Device.Name
}

// ID returns the ID of the data
func (d *DcimDeviceDataWrapper) ID() int {
	return d.Device.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimDeviceDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimDeviceDataWrapper); ok {
		*d.Device = *d2.Device
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimDeviceDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// DcimDeviceRoleDataWrapper represents a DCIM device role data wrapper
type DcimDeviceRoleDataWrapper struct {
	DeviceRole *DcimDeviceRole

	placeholder bool
	hasParent   bool
}

func (*DcimDeviceRoleDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (d *DcimDeviceRoleDataWrapper) Data() any {
	return d.DeviceRole
}

// IsValid returns true if the DeviceRole is not nil
func (d *DcimDeviceRoleDataWrapper) IsValid() bool {
	return d.DeviceRole != nil
}

// AllData returns all the data
func (d *DcimDeviceRoleDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	if d.DeviceRole == nil && d.hasParent {
		d.DeviceRole = NewDcimDeviceRole()
		d.placeholder = true
	}

	if d.DeviceRole.Slug == "" {
		d.DeviceRole.Slug = slug.Make(d.DeviceRole.Name)
	}

	if d.DeviceRole.Color == "" {
		d.DeviceRole.Color = "000000"
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimDeviceRoleDataWrapper) DataType() string {
	return DcimDeviceRoleObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimDeviceRoleDataWrapper) QueryString() string {
	return d.DeviceRole.Name
}

// ID returns the ID of the data
func (d *DcimDeviceRoleDataWrapper) ID() int {
	return d.DeviceRole.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimDeviceRoleDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimDeviceRoleDataWrapper); ok {
		*d.DeviceRole = *d2.DeviceRole
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimDeviceRoleDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// DcimDeviceTypeDataWrapper represents a DCIM device type data wrapper
type DcimDeviceTypeDataWrapper struct {
	DeviceType *DcimDeviceType

	placeholder bool
	hasParent   bool
}

func (*DcimDeviceTypeDataWrapper) comparableData() {}

// Data returns the DeviceType
func (d *DcimDeviceTypeDataWrapper) Data() any {
	return d.DeviceType
}

// IsValid returns true if the DeviceType is not nil
func (d *DcimDeviceTypeDataWrapper) IsValid() bool {
	return d.DeviceType != nil
}

// AllData returns all the data
func (d *DcimDeviceTypeDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	if d.DeviceType == nil && d.hasParent {
		d.DeviceType = NewDcimDeviceType()
		d.placeholder = true
	}

	if d.DeviceType.Slug == "" {
		d.DeviceType.Slug = slug.Make(d.DeviceType.Model)
	}

	for _, t := range NestedTypes(d.DataType()) {
		switch t {
		case DcimManufacturerObjectType:
			manufacturer := DcimManufacturerDataWrapper{Manufacturer: d.DeviceType.Manufacturer, hasParent: true}
			manufacturer.placeholder = d.placeholder

			manufacturerData := manufacturer.AllData()

			d.DeviceType.Manufacturer = manufacturer.Manufacturer

			result = append(result, manufacturerData...)
		}
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimDeviceTypeDataWrapper) DataType() string {
	return DcimDeviceTypeObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimDeviceTypeDataWrapper) QueryString() string {
	return d.DeviceType.Model
}

// ID returns the ID of the data
func (d *DcimDeviceTypeDataWrapper) ID() int {
	return d.DeviceType.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimDeviceTypeDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimDeviceTypeDataWrapper); ok {
		*d.DeviceType = *d2.DeviceType
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimDeviceTypeDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

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

// AllData returns all the data
func (d *DcimInterfaceDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	fmt.Printf("not implemented, data type: %s, required types: %#v\n", d.DataType(), NestedTypes(d.DataType()))

	result = append(result, d)

	return result
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

// ReplaceData replaces the data with the given data
func (d *DcimInterfaceDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimInterfaceDataWrapper); ok {
		*d.Interface = *d2.Interface
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimInterfaceDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// DcimManufacturerDataWrapper represents a DCIM manufacturer data wrapper
type DcimManufacturerDataWrapper struct {
	Manufacturer *DcimManufacturer

	placeholder bool
	hasParent   bool
}

func (*DcimManufacturerDataWrapper) comparableData() {}

// Data returns the Manufacturer
func (d *DcimManufacturerDataWrapper) Data() any {
	return d.Manufacturer
}

// IsValid returns true if the Manufacturer is not nil
func (d *DcimManufacturerDataWrapper) IsValid() bool {
	return d.Manufacturer != nil
}

// AllData returns all the data
func (d *DcimManufacturerDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	if d.Manufacturer == nil && d.hasParent {
		d.Manufacturer = NewDcimManufacturer()
		d.placeholder = true
	}

	if d.Manufacturer.Slug == "" {
		d.Manufacturer.Slug = slug.Make(d.Manufacturer.Name)
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimManufacturerDataWrapper) DataType() string {
	return DcimManufacturerObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimManufacturerDataWrapper) QueryString() string {
	return d.Manufacturer.Name
}

// ID returns the ID of the data
func (d *DcimManufacturerDataWrapper) ID() int {
	return d.Manufacturer.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimManufacturerDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimManufacturerDataWrapper); ok {
		*d.Manufacturer = *d2.Manufacturer
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimManufacturerDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// DcimPlatformDataWrapper represents a DCIM platform data wrapper
type DcimPlatformDataWrapper struct {
	Platform *DcimPlatform

	placeholder bool
	hasParent   bool
}

func (*DcimPlatformDataWrapper) comparableData() {}

// Data returns the Platform
func (d *DcimPlatformDataWrapper) Data() any {
	return d.Platform
}

// IsValid returns true if the Platform is not nil
func (d *DcimPlatformDataWrapper) IsValid() bool {
	return d.Platform != nil
}

// AllData returns all the data
func (d *DcimPlatformDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	if d.Platform == nil && d.hasParent {
		d.Platform = NewDcimPlatform()
		d.placeholder = true
	}

	if d.Platform.Slug == "" {
		d.Platform.Slug = slug.Make(d.Platform.Name)
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimPlatformDataWrapper) DataType() string {
	return DcimPlatformObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimPlatformDataWrapper) QueryString() string {
	return d.Platform.Name
}

// ID returns the ID of the data
func (d *DcimPlatformDataWrapper) ID() int {
	return d.Platform.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimPlatformDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimPlatformDataWrapper); ok {
		*d.Platform = *d2.Platform
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimPlatformDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

// DcimSiteDataWrapper represents a DCIM site data wrapper
type DcimSiteDataWrapper struct {
	Site *DcimSite

	placeholder bool
	hasParent   bool
}

func (*DcimSiteDataWrapper) comparableData() {}

// Data returns the Site
func (d *DcimSiteDataWrapper) Data() any {
	return d.Site
}

// IsValid returns true if the Site is not nil
func (d *DcimSiteDataWrapper) IsValid() bool {
	return d.Site != nil
}

// AllData returns all the data
func (d *DcimSiteDataWrapper) AllData() []ComparableData {
	result := make([]ComparableData, 0)

	if d.Site == nil && d.hasParent {
		d.Site = NewDcimSite()
		d.placeholder = true
	}

	if d.Site.Slug == "" {
		d.Site.Slug = slug.Make(d.Site.Name)
	}

	if d.Site.Status == "" {
		d.Site.Status = DcimSiteStatusActive
	}

	result = append(result, d)

	return result
}

// DataType returns the data type
func (d *DcimSiteDataWrapper) DataType() string {
	return DcimSiteObjectType
}

// QueryString returns the query string needed to retrieve its object state
func (d *DcimSiteDataWrapper) QueryString() string {
	return d.Site.Name
}

// ID returns the ID of the data
func (d *DcimSiteDataWrapper) ID() int {
	return d.Site.ID
}

// ReplaceData replaces the data with the given data
func (d *DcimSiteDataWrapper) ReplaceData(data ComparableData) {
	if d2, ok := data.(*DcimSiteDataWrapper); ok {
		*d.Site = *d2.Site
	}
}

// IsPlaceholder returns true if the data is a placeholder
func (d *DcimSiteDataWrapper) IsPlaceholder() bool {
	return d.placeholder
}

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
	default:
		return nil, fmt.Errorf("unsupported data type %s", dataType)
	}
}
