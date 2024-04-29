package netbox

import (
	"errors"
	"github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb"
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

// IpamIPAddress represents an IPAM IP address
type IpamIPAddress struct {
	ID             int                     `json:"id,omitempty"`
	Address        string                  `json:"address,omitempty"`
	AssignedObject *diodepb.AssignedObject `json:"assigned_object,omitempty" mapstructure:"assigned_object"`
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
