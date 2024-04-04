package netbox

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
	Serial      *string           `json:"serial"`
	Description *string           `json:"description,omitempty"`
	Status      *DcimDeviceStatus `json:"status,omitempty"`
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
}

// DcimDeviceType represents a DCIM device type
type DcimDeviceType struct {
	ID           int               `json:"id,omitempty"`
	Model        string            `json:"model,omitempty"`
	Slug         string            `json:"slug,omitempty"`
	Manufacturer *DcimManufacturer `json:"manufacturer,omitempty"`
	Description  *string           `json:"description,omitempty"`
	PartNumber   *string           `json:"part_number,omitempty"  mapstructure:"part_number"`
}

// DcimInterface represents a DCIM interface
type DcimInterface struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimManufacturer represents a DCIM manufacturer
type DcimManufacturer struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
}

// DcimPlatform represents a DCIM platform
type DcimPlatform struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimSite represents a DCIM site
type DcimSite struct {
	ID          int             `json:"id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Status      *DcimSiteStatus `json:"status,omitempty"`
	Description *string         `json:"description,omitempty"`
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
	manufacturer := NewDcimManufacturer()
	return &DcimDeviceType{
		Model:        "undefined",
		Slug:         "undefined",
		Manufacturer: manufacturer,
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
