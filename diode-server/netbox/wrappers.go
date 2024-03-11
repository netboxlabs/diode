package netbox

import "fmt"

// ComparableData is an interface for NetBox comparable data
type ComparableData interface {
	comparableData()
	Data() any

	IsValid() bool
}

// DcimDeviceDataWrapper represents a DCIM device data wrapper
type DcimDeviceDataWrapper struct {
	Device *DcimDevice
}

func (DcimDeviceDataWrapper) comparableData() {}

// Data returns the Device
func (d DcimDeviceDataWrapper) Data() any {
	return d.Device
}

// IsValid returns true if the Device is not nil
func (d DcimDeviceDataWrapper) IsValid() bool {
	return d.Device != nil
}

// DcimDeviceRoleDataWrapper represents a DCIM device role data wrapper
type DcimDeviceRoleDataWrapper struct {
	DeviceRole *DcimDeviceRole
}

func (DcimDeviceRoleDataWrapper) comparableData() {}

// Data returns the DeviceRole
func (d DcimDeviceRoleDataWrapper) Data() any {
	return d.DeviceRole
}

// IsValid returns true if the DeviceRole is not nil
func (d DcimDeviceRoleDataWrapper) IsValid() bool {
	return d.DeviceRole != nil
}

// DcimDeviceTypeDataWrapper represents a DCIM device type data wrapper
type DcimDeviceTypeDataWrapper struct {
	DeviceType *DcimDeviceType
}

func (DcimDeviceTypeDataWrapper) comparableData() {}

// Data returns the DeviceType
func (d DcimDeviceTypeDataWrapper) Data() any {
	return d.DeviceType
}

// IsValid returns true if the DeviceType is not nil
func (d DcimDeviceTypeDataWrapper) IsValid() bool {
	return d.DeviceType != nil
}

// DcimInterfaceDataWrapper represents a DCIM interface data wrapper
type DcimInterfaceDataWrapper struct {
	Interface *DcimInterface
}

func (DcimInterfaceDataWrapper) comparableData() {}

// Data returns the Interface
func (d DcimInterfaceDataWrapper) Data() any {
	return d.Interface
}

// IsValid returns true if the Interface is not nil
func (d DcimInterfaceDataWrapper) IsValid() bool {
	return d.Interface != nil
}

// DcimManufacturerDataWrapper represents a DCIM manufacturer data wrapper
type DcimManufacturerDataWrapper struct {
	Manufacturer *DcimManufacturer
}

func (DcimManufacturerDataWrapper) comparableData() {}

// Data returns the Manufacturer
func (d DcimManufacturerDataWrapper) Data() any {
	return d.Manufacturer
}

// IsValid returns true if the Manufacturer is not nil
func (d DcimManufacturerDataWrapper) IsValid() bool {
	return d.Manufacturer != nil
}

// DcimPlatformDataWrapper represents a DCIM platform data wrapper
type DcimPlatformDataWrapper struct {
	Platform *DcimPlatform
}

func (DcimPlatformDataWrapper) comparableData() {}

// Data returns the Platform
func (d DcimPlatformDataWrapper) Data() any {
	return d.Platform
}

// IsValid returns true if the Platform is not nil
func (d DcimPlatformDataWrapper) IsValid() bool {
	return d.Platform != nil
}

// DcimSiteDataWrapper represents a DCIM site data wrapper
type DcimSiteDataWrapper struct {
	Site *DcimSite
}

func (DcimSiteDataWrapper) comparableData() {}

// Data returns the Site
func (d DcimSiteDataWrapper) Data() any {
	return d.Site
}

// IsValid returns true if the Site is not nil
func (d DcimSiteDataWrapper) IsValid() bool {
	return d.Site != nil
}

// NewDataWrapper creates a new data wrapper for the given data type
func NewDataWrapper(dataType string) (ComparableData, error) {
	switch dataType {
	case DcimDeviceObjectType:
		return DcimDeviceDataWrapper{}, nil
	case DcimDeviceRoleObjectType:
		return DcimDeviceRoleDataWrapper{}, nil
	case DcimDeviceTypeObjectType:
		return DcimDeviceTypeDataWrapper{}, nil
	case DcimInterfaceObjectType:
		return DcimInterfaceDataWrapper{}, nil
	case DcimManufacturerObjectType:
		return DcimManufacturerDataWrapper{}, nil
	case DcimPlatformObjectType:
		return DcimPlatformDataWrapper{}, nil
	case DcimSiteObjectType:
		return DcimSiteDataWrapper{}, nil
	default:
		return nil, fmt.Errorf("unsupported data type %s", dataType)
	}
}
