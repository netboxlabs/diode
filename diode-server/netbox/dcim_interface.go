package netbox

import "errors"

var (
	// ErrInvalidInterfaceType is returned when the interface type is invalid
	ErrInvalidInterfaceType = errors.New("invalid interface type")

	// ErrInvalidInterfaceMode is returned when the interface mode is invalid
	ErrInvalidInterfaceMode = errors.New("invalid interface mode")

	// DefaultInterfaceType is the default interface type
	DefaultInterfaceType = "other"
)

var interfaceTypesMap = map[string]struct{}{
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

var interfaceModesMap = map[string]struct{}{
	"access":     {},
	"tagged":     {},
	"tagged-all": {},
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
