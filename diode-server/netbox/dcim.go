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
	ID         int             `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`
	Site       *DcimSite       `json:"site,omitempty"`
	Role       *DcimDeviceRole `json:"role,omitempty" mapstructure:"role"`
	DeviceType *DcimDeviceType `json:"device_type,omitempty" mapstructure:"device_type"`
	Platform   *DcimPlatform   `json:"platform,omitempty"`
	Serial     string          `json:"serial"`
}

// DcimDeviceRole represents a DCIM device role
type DcimDeviceRole struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Color string `json:"color,omitempty"`
}

// DcimDeviceType represents a DCIM device type
type DcimDeviceType struct {
	ID           int               `json:"id,omitempty"`
	Model        string            `json:"model,omitempty"`
	Slug         string            `json:"slug,omitempty"`
	Manufacturer *DcimManufacturer `json:"manufacturer,omitempty"`
}

// DcimInterface represents a DCIM interface
type DcimInterface struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimManufacturer represents a DCIM manufacturer
type DcimManufacturer struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimPlatform represents a DCIM platform
type DcimPlatform struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimSite represents a DCIM site
type DcimSite struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// NewDcimSite creates a new DCIM site placeholder
func NewDcimSite() *DcimSite {
	return &DcimSite{
		Name: "undefined",
		Slug: "undefined",
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
	return &DcimDeviceRole{
		Name:  "undefined",
		Slug:  "undefined",
		Color: "000000",
	}
}
