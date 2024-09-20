package netbox

import (
	"errors"
	"fmt"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

var (
	// ErrInvalidInterfaceType is returned when the interface type is invalid
	ErrInvalidInterfaceType = errors.New("invalid interface type")

	// ErrInvalidInterfaceMode is returned when the interface mode is invalid
	ErrInvalidInterfaceMode = errors.New("invalid interface mode")

	// DefaultInterfaceType is the default interface type
	DefaultInterfaceType = "other"

	interfaceTypesMap = map[string]struct{}{
		"virtual":                 {},
		"bridge":                  {},
		"lag":                     {},
		"100base-fx":              {},
		"100base-lfx":             {},
		"100base-tx":              {},
		"100base-t1":              {},
		"1000base-t":              {},
		"1000base-x-gbic":         {},
		"1000base-x-sfp":          {},
		"2.5gbase-t":              {},
		"5gbase-t":                {},
		"10gbase-t":               {},
		"10gbase-cx4":             {},
		"10gbase-x-sfpp":          {},
		"10gbase-x-xfp":           {},
		"10gbase-x-xenpak":        {},
		"10gbase-x-x2":            {},
		"25gbase-x-sfp28":         {},
		"50gbase-x-sfp56":         {},
		"40gbase-x-qsfpp":         {},
		"50gbase-x-sfp28":         {},
		"100gbase-x-cfp":          {},
		"100gbase-x-cfp2":         {},
		"100gbase-x-cfp4":         {},
		"100gbase-x-cxp":          {},
		"100gbase-x-cpak":         {},
		"100gbase-x-dsfp":         {},
		"100gbase-x-sfpdd":        {},
		"100gbase-x-qsfp28":       {},
		"100gbase-x-qsfpdd":       {},
		"200gbase-x-cfp2":         {},
		"200gbase-x-qsfp56":       {},
		"200gbase-x-qsfpdd":       {},
		"400gbase-x-cfp2":         {},
		"400gbase-x-qsfp112":      {},
		"400gbase-x-qsfpdd":       {},
		"400gbase-x-osfp":         {},
		"400gbase-x-osfp-rhs":     {},
		"400gbase-x-cdfp":         {},
		"400gbase-x-cfp8":         {},
		"800gbase-x-qsfpdd":       {},
		"800gbase-x-osfp":         {},
		"1000base-kx":             {},
		"10gbase-kr":              {},
		"10gbase-kx4":             {},
		"25gbase-kr":              {},
		"40gbase-kr4":             {},
		"50gbase-kr":              {},
		"100gbase-kp4":            {},
		"100gbase-kr2":            {},
		"100gbase-kr4":            {},
		"ieee802.11a":             {},
		"ieee802.11g":             {},
		"ieee802.11n":             {},
		"ieee802.11ac":            {},
		"ieee802.11ad":            {},
		"ieee802.11ax":            {},
		"ieee802.11ay":            {},
		"ieee802.15.1":            {},
		"other-wireless":          {},
		"gsm":                     {},
		"cdma":                    {},
		"lte":                     {},
		"sonet-oc3":               {},
		"sonet-oc12":              {},
		"sonet-oc48":              {},
		"sonet-oc192":             {},
		"sonet-oc768":             {},
		"sonet-oc1920":            {},
		"sonet-oc3840":            {},
		"1gfc-sfp":                {},
		"2gfc-sfp":                {},
		"4gfc-sfp":                {},
		"8gfc-sfpp":               {},
		"16gfc-sfpp":              {},
		"32gfc-sfp28":             {},
		"64gfc-qsfpp":             {},
		"128gfc-qsfp28":           {},
		"infiniband-sdr":          {},
		"infiniband-ddr":          {},
		"infiniband-qdr":          {},
		"infiniband-fdr10":        {},
		"infiniband-fdr":          {},
		"infiniband-edr":          {},
		"infiniband-hdr":          {},
		"infiniband-ndr":          {},
		"infiniband-xdr":          {},
		"t1":                      {},
		"e1":                      {},
		"t3":                      {},
		"e3":                      {},
		"xdsl":                    {},
		"docsis":                  {},
		"gpon":                    {},
		"xg-pon":                  {},
		"xgs-pon":                 {},
		"ng-pon2":                 {},
		"epon":                    {},
		"10g-epon":                {},
		"cisco-stackwise":         {},
		"cisco-stackwise-plus":    {},
		"cisco-flexstack":         {},
		"cisco-flexstack-plus":    {},
		"cisco-stackwise-80":      {},
		"cisco-stackwise-160":     {},
		"cisco-stackwise-320":     {},
		"cisco-stackwise-480":     {},
		"cisco-stackwise-1t":      {},
		"juniper-vcp":             {},
		"extreme-summitstack":     {},
		"extreme-summitstack-128": {},
		"extreme-summitstack-256": {},
		"extreme-summitstack-512": {},
		"other":                   {},
	}

	interfaceModesMap = map[string]struct{}{
		"access":     {},
		"tagged":     {},
		"tagged-all": {},
	}
)

const (
	// DcimDeviceObjectType represents the DCIM device object type
	DcimDeviceObjectType = "dcim.device"

	// DcimDeviceRoleObjectType represents the DCIM device role object type
	DcimDeviceRoleObjectType = "dcim.devicerole"

	// DcimDeviceTypeObjectType represents the DCIM device type object type
	DcimDeviceTypeObjectType = "dcim.devicetype"

	// DcimInterfaceObjectType represents the DCIM interface object type
	DcimInterfaceObjectType = "dcim.interface"

	// DcimManufacturerObjectType represents the DCIM manufacturer object type
	DcimManufacturerObjectType = "dcim.manufacturer"

	// DcimPlatformObjectType represents the DCIM platform object type
	DcimPlatformObjectType = "dcim.platform"

	// DcimSiteObjectType represents the DCIM site object type
	DcimSiteObjectType = "dcim.site"
)

// DcimDevice represents a DCIM device
type DcimDevice struct {
	ID          int               `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Site        *DcimSite         `json:"site,omitempty"`
	Role        *DcimDeviceRole   `json:"role,omitempty" mapstructure:"role"`
	DeviceType  *DcimDeviceType   `json:"device_type,omitempty" mapstructure:"device_type"`
	Platform    *DcimPlatform     `json:"platform,omitempty"`
	Serial      *string           `json:"serial,omitempty"`
	Description *string           `json:"description,omitempty"`
	Status      *DcimDeviceStatus `json:"status,omitempty"`
	AssetTag    *string           `json:"asset_tag,omitempty" mapstructure:"asset_tag"`
	PrimaryIPv4 *IpamIPAddress    `json:"primary_ip4,omitempty" mapstructure:"primary_ip4"`
	PrimaryIPv6 *IpamIPAddress    `json:"primary_ip6,omitempty" mapstructure:"primary_ip6"`
	Comments    *string           `json:"comments,omitempty"`
	Tags        []*Tag            `json:"tags,omitempty"`
}

// DcimDeviceStatus represents a DCIM device status
type DcimDeviceStatus string

const (
	// DcimDeviceStatusOffline represents the offline DCIM device status
	DcimDeviceStatusOffline DcimDeviceStatus = "offline"

	// DcimDeviceStatusActive represents the active DCIM device status
	DcimDeviceStatusActive DcimDeviceStatus = "active"

	// DcimDeviceStatusPlanned represents the planned DCIM device status
	DcimDeviceStatusPlanned DcimDeviceStatus = "planned"

	// DcimDeviceStatusStaged represents the staged DCIM device status
	DcimDeviceStatusStaged DcimDeviceStatus = "staged"

	// DcimDeviceStatusFailed represents the failed DCIM device status
	DcimDeviceStatusFailed DcimDeviceStatus = "failed"

	// DcimDeviceStatusInventory represents the inventory DCIM device status
	DcimDeviceStatusInventory DcimDeviceStatus = "inventory"

	// DcimDeviceStatusDecommissioning represents the decommissioning DCIM device status
	DcimDeviceStatusDecommissioning DcimDeviceStatus = "decommissioning"
)

// DcimDeviceRole represents a DCIM device role
type DcimDeviceRole struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        []*Tag  `json:"tags,omitempty"`
}

// DcimDeviceType represents a DCIM device type
type DcimDeviceType struct {
	ID           int               `json:"id,omitempty"`
	Model        string            `json:"model,omitempty"`
	Slug         string            `json:"slug,omitempty"`
	Manufacturer *DcimManufacturer `json:"manufacturer,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Comments     *string           `json:"comments,omitempty"`
	PartNumber   *string           `json:"part_number,omitempty" mapstructure:"part_number"`
	Tags         []*Tag            `json:"tags,omitempty"`
}

// DcimManufacturer represents a DCIM manufacturer
type DcimManufacturer struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        []*Tag  `json:"tags,omitempty"`
}

// DcimPlatform represents a DCIM platform
type DcimPlatform struct {
	ID           int               `json:"id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Slug         string            `json:"slug,omitempty"`
	Manufacturer *DcimManufacturer `json:"manufacturer,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Tags         []*Tag            `json:"tags,omitempty"`
}

// DcimSite represents a DCIM site
type DcimSite struct {
	ID          int             `json:"id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Status      *DcimSiteStatus `json:"status,omitempty"`
	Facility    *string         `json:"facility,omitempty"`
	TimeZone    *string         `json:"time_zone,omitempty" mapstructure:"time_zone"`
	Description *string         `json:"description,omitempty"`
	Comments    *string         `json:"comments,omitempty"`
	Tags        []*Tag          `json:"tags,omitempty"`
}

// DcimSiteStatus represents a DCIM site status
type DcimSiteStatus string

const (
	// DcimSiteStatusPlanned represents the planned DCIM site status
	DcimSiteStatusPlanned DcimSiteStatus = "planned"

	// DcimSiteStatusStaging represents the staging DCIM site status
	DcimSiteStatusStaging DcimSiteStatus = "staging"

	// DcimSiteStatusActive represents the active DCIM site status
	DcimSiteStatusActive DcimSiteStatus = "active"

	// DcimSiteStatusDecommissioning represents the decommissioning DCIM site status
	DcimSiteStatusDecommissioning DcimSiteStatus = "decommissioning"

	// DcimSiteStatusRetired represents the retired DCIM site status
	DcimSiteStatusRetired DcimSiteStatus = "retired"
)

// NewDcimSite creates a new DCIM site placeholder
func NewDcimSite() *DcimSite {
	status := DcimSiteStatusActive
	return &DcimSite{
		Name:   "undefined",
		Slug:   "undefined",
		Status: &status,
	}
}

// NewDcimPlatform creates a new DCIM platform placeholder
func NewDcimPlatform() *DcimPlatform {
	return &DcimPlatform{
		Name: "undefined",
		Slug: "undefined",
	}
}

// NewDcimManufacturer creates a new DCIM manufacturer placeholder
func NewDcimManufacturer() *DcimManufacturer {
	return &DcimManufacturer{
		Name: "undefined",
		Slug: "undefined",
	}
}

// NewDcimDeviceType creates a new DCIM device type placeholder
func NewDcimDeviceType() *DcimDeviceType {
	return &DcimDeviceType{
		Model:        "undefined",
		Slug:         "undefined",
		Manufacturer: NewDcimManufacturer(),
	}
}

// NewDcimDeviceRole creates a new DCIM device role placeholder
func NewDcimDeviceRole() *DcimDeviceRole {
	color := "000000"
	return &DcimDeviceRole{
		Name:  "undefined",
		Slug:  "undefined",
		Color: &color,
	}
}

// NewDcimDevice creates a new DCIM device placeholder
func NewDcimDevice() *DcimDevice {
	status := DcimDeviceStatusActive
	return &DcimDevice{
		Name:       "undefined",
		Site:       NewDcimSite(),
		Role:       NewDcimDeviceRole(),
		DeviceType: NewDcimDeviceType(),
		Platform:   NewDcimPlatform(),
		Status:     &status,
	}
}

// FromProtoDeviceEntity converts a diode device entity to a DCIM device
func FromProtoDeviceEntity(entity *diodepb.Entity) (*DcimDevice, error) {
	if entity == nil || entity.GetDevice() == nil {
		return nil, fmt.Errorf("entity is nil or not a device")
	}

	return FromProtoDevice(entity.GetDevice()), nil
}

// FromProtoDevice converts a diode device to a DCIM device
func FromProtoDevice(devicePb *diodepb.Device) *DcimDevice {
	if devicePb == nil {
		return nil
	}

	return &DcimDevice{
		Name:        devicePb.Name,
		Site:        FromProtoSite(devicePb.Site),
		Role:        FromProtoRole(devicePb.Role),
		DeviceType:  FromProtoDeviceType(devicePb.DeviceType),
		Platform:    FromProtoPlatform(devicePb.Platform),
		Serial:      devicePb.Serial,
		Description: devicePb.Description,
		Status:      (*DcimDeviceStatus)(&devicePb.Status),
		AssetTag:    devicePb.AssetTag,
		PrimaryIPv4: nil,
		PrimaryIPv6: nil,
		Comments:    devicePb.Comments,
		Tags:        FromProtoTags(devicePb.Tags),
	}
}

// FromProtoDeviceRoleEntity converts a diode device role entity to a DCIM device role
func FromProtoDeviceRoleEntity(entity *diodepb.Entity) (*DcimDeviceRole, error) {
	if entity == nil || entity.GetDeviceRole() == nil {
		return nil, fmt.Errorf("entity is nil or not a device role")
	}

	return FromProtoRole(entity.GetDeviceRole()), nil
}

// FromProtoDeviceTypeEntity converts a diode device type entity to a DCIM device type
func FromProtoDeviceTypeEntity(entity *diodepb.Entity) (*DcimDeviceType, error) {
	if entity == nil || entity.GetDeviceType() == nil {
		return nil, fmt.Errorf("entity is nil or not a device type")
	}

	return FromProtoDeviceType(entity.GetDeviceType()), nil
}

// FromProtoManufacturerEntity converts a diode manufacturer entity to a DCIM manufacturer
func FromProtoManufacturerEntity(entity *diodepb.Entity) (*DcimManufacturer, error) {
	if entity == nil || entity.GetManufacturer() == nil {
		return nil, fmt.Errorf("entity is nil or not a manufacturer")
	}

	return FromProtoManufacturer(entity.GetManufacturer()), nil
}

// FromProtoPlatformEntity converts a diode platform entity to a DCIM platform
func FromProtoPlatformEntity(entity *diodepb.Entity) (*DcimPlatform, error) {
	if entity == nil || entity.GetPlatform() == nil {
		return nil, fmt.Errorf("entity is nil or not a platform")
	}

	return FromProtoPlatform(entity.GetPlatform()), nil
}

// FromProtoSiteEntity converts a diode site entity to a DCIM site
func FromProtoSiteEntity(entity *diodepb.Entity) (*DcimSite, error) {
	if entity == nil || entity.GetSite() == nil {
		return nil, fmt.Errorf("entity is nil or not a site")
	}

	return FromProtoSite(entity.GetSite()), nil
}

// FromProtoSite converts a diode site to a DCIM site
func FromProtoSite(sitePb *diodepb.Site) *DcimSite {
	if sitePb == nil {
		return nil
	}

	return &DcimSite{
		Name:        sitePb.Name,
		Slug:        sitePb.Slug,
		Status:      (*DcimSiteStatus)(&sitePb.Status),
		Facility:    sitePb.Facility,
		TimeZone:    sitePb.TimeZone,
		Description: sitePb.Description,
		Comments:    sitePb.Comments,
		Tags:        FromProtoTags(sitePb.Tags),
	}
}

// FromProtoRole converts a diode role to a DCIM device role
func FromProtoRole(rolePb *diodepb.Role) *DcimDeviceRole {
	if rolePb == nil {
		return nil
	}

	var color *string
	if rolePb.Color != "" {
		color = &rolePb.Color
	}

	return &DcimDeviceRole{
		Name:        rolePb.Name,
		Slug:        rolePb.Slug,
		Color:       color,
		Description: rolePb.Description,
		Tags:        FromProtoTags(rolePb.Tags),
	}
}

// FromProtoDeviceType converts a diode device type to a DCIM device type
func FromProtoDeviceType(deviceTypePb *diodepb.DeviceType) *DcimDeviceType {
	if deviceTypePb == nil {
		return nil
	}

	return &DcimDeviceType{
		Model:        deviceTypePb.Model,
		Slug:         deviceTypePb.Slug,
		Manufacturer: FromProtoManufacturer(deviceTypePb.Manufacturer),
		Description:  deviceTypePb.Description,
		Comments:     deviceTypePb.Comments,
		PartNumber:   deviceTypePb.PartNumber,
		Tags:         FromProtoTags(deviceTypePb.Tags),
	}
}

// FromProtoManufacturer converts a diode manufacturer to a DCIM manufacturer
func FromProtoManufacturer(manufacturerPb *diodepb.Manufacturer) *DcimManufacturer {
	if manufacturerPb == nil {
		return nil
	}

	return &DcimManufacturer{
		Name:        manufacturerPb.Name,
		Slug:        manufacturerPb.Slug,
		Description: manufacturerPb.Description,
		Tags:        FromProtoTags(manufacturerPb.Tags),
	}
}

// FromProtoPlatform converts a diode platform to a DCIM platform
func FromProtoPlatform(platformPb *diodepb.Platform) *DcimPlatform {
	if platformPb == nil {
		return nil
	}

	return &DcimPlatform{
		Name:         platformPb.Name,
		Slug:         platformPb.Slug,
		Manufacturer: FromProtoManufacturer(platformPb.Manufacturer),
		Description:  platformPb.Description,
		Tags:         FromProtoTags(platformPb.Tags),
	}
}

// DcimInterface represents a DCIM interface
type DcimInterface struct {
	ID            int         `json:"id,omitempty"`
	Device        *DcimDevice `json:"device,omitempty"`
	Name          string      `json:"name,omitempty"`
	Label         *string     `json:"label,omitempty"`
	Type          *string     `json:"type,omitempty"`
	Enabled       *bool       `json:"enabled,omitempty"`
	MTU           *int        `json:"mtu,omitempty"`
	MACAddress    *string     `json:"mac_address,omitempty" mapstructure:"mac_address,omitempty"`
	Speed         *int        `json:"speed,omitempty"`
	WWN           *string     `json:"wwn,omitempty"`
	MgmtOnly      *bool       `json:"mgmt_only,omitempty" mapstructure:"mgmt_only,omitempty"`
	Description   *string     `json:"description,omitempty"`
	MarkConnected *bool       `json:"mark_connected,omitempty" mapstructure:"mark_connected,omitempty"`
	Mode          *string     `json:"mode,omitempty"`
	Tags          []*Tag      `json:"tags,omitempty"`
}

func validateInterfaceType(t string) bool {
	_, ok := interfaceTypesMap[t]
	return ok
}

func validateInterfaceMode(m string) bool {
	if m == "" {
		return true
	}
	_, ok := interfaceModesMap[m]
	return ok
}

// Validate checks if the DCIM interface is valid
func (i *DcimInterface) Validate() error {
	if i.Type != nil && !validateInterfaceType(*i.Type) {
		return ErrInvalidInterfaceType
	}
	if i.Mode != nil && !validateInterfaceMode(*i.Mode) {
		return ErrInvalidInterfaceMode
	}
	return nil
}

// NewDcimInterface creates a new DCIM interface placeholder
func NewDcimInterface() *DcimInterface {
	return &DcimInterface{
		Device: NewDcimDevice(),
		Name:   "undefined",
	}
}

// FromProtoInterfaceEntity converts a diode interface entity to a DCIM interface
func FromProtoInterfaceEntity(entity *diodepb.Entity) (*DcimInterface, error) {
	if entity == nil || entity.GetInterface() == nil {
		return nil, fmt.Errorf("entity is nil or not an interface")
	}

	return FromProtoInterface(entity.GetInterface()), nil
}

// FromProtoInterface converts a diode interface to a DCIM interface
func FromProtoInterface(interfacePb *diodepb.Interface) *DcimInterface {
	if interfacePb == nil {
		return nil
	}

	var interfaceType *string
	if interfacePb.Type != "" {
		interfaceType = &interfacePb.Type
	}

	var mode *string
	if interfacePb.Mode != "" {
		mode = &interfacePb.Mode
	}

	return &DcimInterface{
		Name:          interfacePb.Name,
		Device:        FromProtoDevice(interfacePb.Device),
		Label:         interfacePb.Label,
		Type:          interfaceType,
		Enabled:       interfacePb.Enabled,
		MTU:           int32PtrToIntPtr(interfacePb.Mtu),
		MACAddress:    interfacePb.MacAddress,
		Speed:         int32PtrToIntPtr(interfacePb.Speed),
		WWN:           interfacePb.Wwn,
		MgmtOnly:      interfacePb.MgmtOnly,
		Description:   interfacePb.Description,
		MarkConnected: interfacePb.MarkConnected,
		Mode:          mode,
		Tags:          FromProtoTags(interfacePb.Tags),
	}
}
