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
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimDeviceRole represents a DCIM device role
type DcimDeviceRole struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// DcimDeviceType represents a DCIM device type
type DcimDeviceType struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
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