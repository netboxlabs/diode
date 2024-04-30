package netbox

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

const (
	// IpamIPAddressObjectType represents the IPAM IP address object type
	IpamIPAddressObjectType = "ipam.ipaddress"
)

var (
	// ErrInvalidIPAddressStatus is returned when the IP address status is invalid
	ErrInvalidIPAddressStatus = errors.New("invalid IP address status")

	// ErrInvalidIPAddressRole is returned when the IP address role is invalid
	ErrInvalidIPAddressRole = errors.New("invalid IP address role")

	// DefaultIPAddressStatus is the default status for an IP address
	DefaultIPAddressStatus = "active"
)

// IPAddressAssignedObject represents an assigned object for an IP address
type IPAddressAssignedObject interface {
	ipAddressAssignedObject()
}

// IPAddressInterface represents an assigned interface for an IP address
type IPAddressInterface struct {
	Interface *DcimInterface `json:"Interface,omitempty" mapstructure:"Interface"`
}

func (*IPAddressInterface) ipAddressAssignedObject() {}

// IpamIPAddress represents an IPAM IP address
type IpamIPAddress struct {
	ID             int                     `json:"id,omitempty"`
	Address        string                  `json:"address,omitempty"`
	AssignedObject IPAddressAssignedObject `json:"AssignedObject,omitempty" mapstructure:"AssignedObject"`
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

// IpamIPAddressAssignedObjectHookFunc returns a mapstructure decode hook function
// for IPAM IP address assigned objects.
func IpamIPAddressAssignedObjectHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Map {
			return data, nil
		}

		if t.Implements(reflect.TypeOf((*IPAddressAssignedObject)(nil)).Elem()) {
			for k := range data.(map[string]any) {
				if k == "Interface" {
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
