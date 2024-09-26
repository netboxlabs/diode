package netbox

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

const (
	// IpamIPAddressObjectType represents the IPAM IP address object type
	IpamIPAddressObjectType = "ipam.ipaddress"

	// IpamPrefixObjectType represents the IPAM Prefix object type
	IpamPrefixObjectType = "ipam.prefix"
)

var (
	// ErrInvalidIPAddressStatus is returned when the IP address status is invalid
	ErrInvalidIPAddressStatus = errors.New("invalid IP address status")

	// ErrInvalidIPAddressRole is returned when the IP address role is invalid
	ErrInvalidIPAddressRole = errors.New("invalid IP address role")

	// DefaultIPAddressStatus is the default status for an IP address
	DefaultIPAddressStatus = "active"

	// ErrInvalidIPrefixStatus is returned when the IPAM Prefix status is invalid
	ErrInvalidIPrefixStatus = errors.New("invalid prefix status")

	// DefaultPrefixStatus is the default status for the IpamPrefix
	DefaultPrefixStatus = "active"
)

// IPAddressAssignedObject represents an assigned object for an IP address
type IPAddressAssignedObject interface {
	ipAddressAssignedObject()
}

// IPAddressInterface represents an assigned interface for an IP address
type IPAddressInterface struct {
	Interface *DcimInterface `json:"interface,omitempty" mapstructure:"Interface"`
}

func (*IPAddressInterface) ipAddressAssignedObject() {}

// IpamIPAddress represents an IPAM IP address
type IpamIPAddress struct {
	ID             int                     `json:"id,omitempty"`
	Address        string                  `json:"address,omitempty"`
	AssignedObject IPAddressAssignedObject `json:"assigned_object,omitempty" mapstructure:"AssignedObject"`
	Status         *string                 `json:"status,omitempty"`
	Role           *string                 `json:"role,omitempty"`
	DNSName        *string                 `json:"dns_name,omitempty" mapstructure:"dns_name"`
	Description    *string                 `json:"description,omitempty"`
	Comments       *string                 `json:"comments,omitempty"`
	Tags           []*Tag                  `json:"tags,omitempty"`
}

var ipAddressStatusMap = map[string]struct{}{
	"active":     {},
	"reserved":   {},
	"deprecated": {},
	"dhcp":       {},
	"slaac":      {},
}

var ipAddressRoleMap = map[string]struct{}{
	"loopback":  {},
	"secondary": {},
	"anycast":   {},
	"vip":       {},
	"vrrp":      {},
	"hsrp":      {},
	"glbp":      {},
	"carp":      {},
}

func validateIPAddressStatus(s string) bool {
	_, ok := ipAddressStatusMap[s]
	return ok
}

func validateIPAddressRole(r string) bool {
	_, ok := ipAddressRoleMap[r]
	return ok
}

// Validate checks if the IPAM IP address is valid
func (ip *IpamIPAddress) Validate() error {
	if ip.Status != nil && !validateIPAddressStatus(*ip.Status) {
		return ErrInvalidIPAddressStatus
	}
	if ip.Role != nil && !validateIPAddressRole(*ip.Role) {
		return ErrInvalidIPAddressRole
	}
	return nil
}

// IpamIPAddressAssignedObjectMatchName returns true if the mapstructure key matches the field name
func IpamIPAddressAssignedObjectMatchName(mapKey, fieldName string) bool {
	return mapKey == fieldName || strcase.ToSnake(mapKey) == strcase.ToSnake(fieldName)
}

// IpamIPAddressAssignedObjectHookFunc returns a mapstructure decode hook function
// for IPAM IP address assigned objects.
func IpamIPAddressAssignedObjectHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Map {
			return data, nil
		}

		if t.Implements(reflect.TypeOf((*IPAddressAssignedObject)(nil)).Elem()) {
			for k := range data.(map[string]any) {
				if strings.ToLower(k) == "interface" {
					var ipInterface IPAddressInterface
					if err := mapstructure.Decode(data, &ipInterface); err != nil {
						return nil, fmt.Errorf("failed to decode ingest entity %w", err)
					}
					return &ipInterface, nil
				}
			}
		}

		return data, nil
	}
}

// FromProtoIPAddressEntity converts a diode IP address entity to a IPAM IP address
func FromProtoIPAddressEntity(entity *diodepb.Entity) (*IpamIPAddress, error) {
	if entity == nil || entity.GetIpAddress() == nil {
		return nil, fmt.Errorf("entity is nil or not an IP address")
	}

	return FromProtoIPAddress(entity.GetIpAddress()), nil
}

// FromProtoIPAddress converts a diode IP address to a IPAM IP address
func FromProtoIPAddress(ipaddressPb *diodepb.IPAddress) *IpamIPAddress {
	if ipaddressPb == nil {
		return nil
	}

	var status *string
	if ipaddressPb.Status != "" {
		status = &ipaddressPb.Status
	}

	var role *string
	if ipaddressPb.Role != "" {
		role = &ipaddressPb.Role
	}

	return &IpamIPAddress{
		Address:        ipaddressPb.Address,
		AssignedObject: FromProtoIPAddressAssignedObject(ipaddressPb.AssignedObject),
		Status:         status,
		Role:           role,
		DNSName:        ipaddressPb.DnsName,
		Description:    ipaddressPb.Description,
		Comments:       ipaddressPb.Comments,
		Tags:           FromProtoTags(ipaddressPb.Tags),
	}
}

// FromProtoIPAddressAssignedObject converts a diode IP address assigned object to an IPAM IP address assigned object
func FromProtoIPAddressAssignedObject(assignedObjectPb any) IPAddressAssignedObject {
	if assignedObjectPb == nil {
		return nil
	}

	switch assignedObjectPb := assignedObjectPb.(type) {
	case *diodepb.IPAddress_Interface:
		return FromProtoIPAddressInterface(assignedObjectPb)
	default:
		return nil
	}
}

// FromProtoIPAddressInterface converts a diode IP address interface to an IPAM IP address interface
func FromProtoIPAddressInterface(interfacePb *diodepb.IPAddress_Interface) *IPAddressInterface {
	if interfacePb == nil {
		return nil
	}

	return &IPAddressInterface{
		Interface: FromProtoInterface(interfacePb.Interface),
	}
}

// IpamPrefix represents an IPAM Prefix
type IpamPrefix struct {
	ID           int       `json:"id,omitempty"`
	Prefix       string    `json:"prefix,omitempty"`
	Site         *DcimSite `json:"site,omitempty"`
	Status       *string   `json:"status,omitempty"`
	IsPool       *bool     `json:"is_pool,omitempty"`
	MarkUtilized *bool     `json:"mark_utilized,omitempty" mapstructure:"mark_utilized"`
	Description  *string   `json:"description,omitempty"`
	Comments     *string   `json:"comments,omitempty"`
	Tags         []*Tag    `json:"tags,omitempty"`
}

var prefixStatusMap = map[string]struct{}{
	"active":     {},
	"container":  {},
	"reserved":   {},
	"deprecated": {},
}

func validatePrefixStatus(r string) bool {
	_, ok := prefixStatusMap[r]
	return ok
}

// Validate checks if the IPAM prefix is valid
func (p *IpamPrefix) Validate() error {
	if p.Status != nil && !validatePrefixStatus(*p.Status) {
		return ErrInvalidIPrefixStatus
	}
	return nil
}

// FromProtoPrefixEntity converts a diode prefix entity to a IPAM prefix
func FromProtoPrefixEntity(entity *diodepb.Entity) (*IpamPrefix, error) {
	if entity == nil || entity.GetPrefix() == nil {
		return nil, fmt.Errorf("entity is nil or not a prefix")
	}

	return FromProtoPrefix(entity.GetPrefix()), nil
}

// FromProtoPrefix converts a diode prefix to a IPAM prefix
func FromProtoPrefix(prefixPb *diodepb.Prefix) *IpamPrefix {
	if prefixPb == nil {
		return nil
	}

	var status *string
	if prefixPb.Status != "" {
		status = &prefixPb.Status
	}

	return &IpamPrefix{
		Prefix:       prefixPb.Prefix,
		Site:         FromProtoSite(prefixPb.Site),
		Status:       status,
		IsPool:       prefixPb.IsPool,
		MarkUtilized: prefixPb.MarkUtilized,
		Description:  prefixPb.Description,
		Comments:     prefixPb.Comments,
		Tags:         FromProtoTags(prefixPb.Tags),
	}
}
