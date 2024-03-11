package netbox

import "fmt"

type ComparableData interface {
	comparableData()
	Data() any

	IsValid() bool
}

type DcimDeviceDataWrapper struct {
	Device *DcimDevice
}

func (DcimDeviceDataWrapper) comparableData() {}

func (d DcimDeviceDataWrapper) Data() any {
	return d.Device
}

func (d DcimDeviceDataWrapper) IsValid() bool {
	return d.Device != nil
}

type DcimDeviceRoleDataWrapper struct {
	DeviceRole *DcimDeviceRole
}

func (DcimDeviceRoleDataWrapper) comparableData() {}

func (d DcimDeviceRoleDataWrapper) Data() any {
	return d.DeviceRole
}

func (d DcimDeviceRoleDataWrapper) IsValid() bool {
	return d.DeviceRole != nil
}

type DcimDeviceTypeDataWrapper struct {
	DeviceType *DcimDeviceType
}

func (DcimDeviceTypeDataWrapper) comparableData() {}

func (d DcimDeviceTypeDataWrapper) Data() any {
	return d.DeviceType
}

func (d DcimDeviceTypeDataWrapper) IsValid() bool {
	return d.DeviceType != nil
}

type DcimInterfaceDataWrapper struct {
	Interface *DcimInterface
}

func (DcimInterfaceDataWrapper) comparableData() {}

func (d DcimInterfaceDataWrapper) Data() any {
	return d.Interface
}

func (d DcimInterfaceDataWrapper) IsValid() bool {
	return d.Interface != nil
}

type DcimManufacturerDataWrapper struct {
	Manufacturer *DcimManufacturer
}

func (DcimManufacturerDataWrapper) comparableData() {}

func (d DcimManufacturerDataWrapper) Data() any {
	return d.Manufacturer
}

func (d DcimManufacturerDataWrapper) IsValid() bool {
	return d.Manufacturer != nil
}

type DcimPlatformDataWrapper struct {
	Platform *DcimPlatform
}

func (DcimPlatformDataWrapper) comparableData() {}

func (d DcimPlatformDataWrapper) Data() any {
	return d.Platform
}

func (d DcimPlatformDataWrapper) IsValid() bool {
	return d.Platform != nil
}

type DcimSiteDataWrapper struct {
	Site *DcimSite
}

func (DcimSiteDataWrapper) comparableData() {}

func (d DcimSiteDataWrapper) Data() any {
	return d.Site
}

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
