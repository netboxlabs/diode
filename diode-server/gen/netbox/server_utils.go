// Generated code. DO NOT EDIT.
// Timestamp: 2025-11-04 12:25:51Z
package netbox

import (
	"fmt"

	pb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

const (
	ASNObjectType                           = "ipam.asn"
	ASNObjectTypeName                       = "ASN"
	ASNRangeObjectType                      = "ipam.asnrange"
	ASNRangeObjectTypeName                  = "ASN Range"
	AggregateObjectType                     = "ipam.aggregate"
	AggregateObjectTypeName                 = "Aggregate"
	CableObjectType                         = "dcim.cable"
	CableObjectTypeName                     = "Cable"
	CablePathObjectType                     = "dcim.cablepath"
	CablePathObjectTypeName                 = "Cable Path"
	CableTerminationObjectType              = "dcim.cabletermination"
	CableTerminationObjectTypeName          = "Cable Termination"
	CircuitObjectType                       = "circuits.circuit"
	CircuitObjectTypeName                   = "Circuit"
	CircuitGroupObjectType                  = "circuits.circuitgroup"
	CircuitGroupObjectTypeName              = "Circuit Group"
	CircuitGroupAssignmentObjectType        = "circuits.circuitgroupassignment"
	CircuitGroupAssignmentObjectTypeName    = "Circuit Group Assignment"
	CircuitTerminationObjectType            = "circuits.circuittermination"
	CircuitTerminationObjectTypeName        = "Circuit Termination"
	CircuitTypeObjectType                   = "circuits.circuittype"
	CircuitTypeObjectTypeName               = "Circuit Type"
	ClusterObjectType                       = "virtualization.cluster"
	ClusterObjectTypeName                   = "Cluster"
	ClusterGroupObjectType                  = "virtualization.clustergroup"
	ClusterGroupObjectTypeName              = "Cluster Group"
	ClusterTypeObjectType                   = "virtualization.clustertype"
	ClusterTypeObjectTypeName               = "Cluster Type"
	ConsolePortObjectType                   = "dcim.consoleport"
	ConsolePortObjectTypeName               = "Console Port"
	ConsoleServerPortObjectType             = "dcim.consoleserverport"
	ConsoleServerPortObjectTypeName         = "Console Server Port"
	ContactObjectType                       = "tenancy.contact"
	ContactObjectTypeName                   = "Contact"
	ContactAssignmentObjectType             = "tenancy.contactassignment"
	ContactAssignmentObjectTypeName         = "Contact Assignment"
	ContactGroupObjectType                  = "tenancy.contactgroup"
	ContactGroupObjectTypeName              = "Contact Group"
	ContactRoleObjectType                   = "tenancy.contactrole"
	ContactRoleObjectTypeName               = "Contact Role"
	DeviceObjectType                        = "dcim.device"
	DeviceObjectTypeName                    = "Device"
	DeviceBayObjectType                     = "dcim.devicebay"
	DeviceBayObjectTypeName                 = "Device Bay"
	DeviceRoleObjectType                    = "dcim.devicerole"
	DeviceRoleObjectTypeName                = "Device Role"
	DeviceTypeObjectType                    = "dcim.devicetype"
	DeviceTypeObjectTypeName                = "Device Type"
	FHRPGroupObjectType                     = "ipam.fhrpgroup"
	FHRPGroupObjectTypeName                 = "FHRP Group"
	FHRPGroupAssignmentObjectType           = "ipam.fhrpgroupassignment"
	FHRPGroupAssignmentObjectTypeName       = "FHRP Group Assignment"
	FrontPortObjectType                     = "dcim.frontport"
	FrontPortObjectTypeName                 = "Front Port"
	IKEPolicyObjectType                     = "vpn.ikepolicy"
	IKEPolicyObjectTypeName                 = "IKE Policy"
	IKEProposalObjectType                   = "vpn.ikeproposal"
	IKEProposalObjectTypeName               = "IKE Proposal"
	IPAddressObjectType                     = "ipam.ipaddress"
	IPAddressObjectTypeName                 = "IP"
	IPRangeObjectType                       = "ipam.iprange"
	IPRangeObjectTypeName                   = "IP Range"
	IPSecPolicyObjectType                   = "vpn.ipsecpolicy"
	IPSecPolicyObjectTypeName               = "IP Sec Policy"
	IPSecProfileObjectType                  = "vpn.ipsecprofile"
	IPSecProfileObjectTypeName              = "IP Sec Profile"
	IPSecProposalObjectType                 = "vpn.ipsecproposal"
	IPSecProposalObjectTypeName             = "IP Sec Proposal"
	InterfaceObjectType                     = "dcim.interface"
	InterfaceObjectTypeName                 = "Interface"
	InventoryItemObjectType                 = "dcim.inventoryitem"
	InventoryItemObjectTypeName             = "Inventory Item"
	InventoryItemRoleObjectType             = "dcim.inventoryitemrole"
	InventoryItemRoleObjectTypeName         = "Inventory Item Role"
	L2VPNObjectType                         = "vpn.l2vpn"
	L2VPNObjectTypeName                     = "L2VPN"
	L2VPNTerminationObjectType              = "vpn.l2vpntermination"
	L2VPNTerminationObjectTypeName          = "L2VPN Termination"
	LocationObjectType                      = "dcim.location"
	LocationObjectTypeName                  = "Location"
	MACAddressObjectType                    = "dcim.macaddress"
	MACAddressObjectTypeName                = "MAC Address"
	ManufacturerObjectType                  = "dcim.manufacturer"
	ManufacturerObjectTypeName              = "Manufacturer"
	ModuleObjectType                        = "dcim.module"
	ModuleObjectTypeName                    = "Module"
	ModuleBayObjectType                     = "dcim.modulebay"
	ModuleBayObjectTypeName                 = "Module Bay"
	ModuleTypeObjectType                    = "dcim.moduletype"
	ModuleTypeObjectTypeName                = "Module Type"
	PlatformObjectType                      = "dcim.platform"
	PlatformObjectTypeName                  = "Platform"
	PowerFeedObjectType                     = "dcim.powerfeed"
	PowerFeedObjectTypeName                 = "Power Feed"
	PowerOutletObjectType                   = "dcim.poweroutlet"
	PowerOutletObjectTypeName               = "Power Outlet"
	PowerPanelObjectType                    = "dcim.powerpanel"
	PowerPanelObjectTypeName                = "Power Panel"
	PowerPortObjectType                     = "dcim.powerport"
	PowerPortObjectTypeName                 = "Power Port"
	PrefixObjectType                        = "ipam.prefix"
	PrefixObjectTypeName                    = "Prefix"
	ProviderObjectType                      = "circuits.provider"
	ProviderObjectTypeName                  = "Provider"
	ProviderAccountObjectType               = "circuits.provideraccount"
	ProviderAccountObjectTypeName           = "Provider Account"
	ProviderNetworkObjectType               = "circuits.providernetwork"
	ProviderNetworkObjectTypeName           = "Provider Network"
	RIRObjectType                           = "ipam.rir"
	RIRObjectTypeName                       = "RIR"
	RackObjectType                          = "dcim.rack"
	RackObjectTypeName                      = "Rack"
	RackReservationObjectType               = "dcim.rackreservation"
	RackReservationObjectTypeName           = "Rack Reservation"
	RackRoleObjectType                      = "dcim.rackrole"
	RackRoleObjectTypeName                  = "Rack Role"
	RackTypeObjectType                      = "dcim.racktype"
	RackTypeObjectTypeName                  = "Rack Type"
	RearPortObjectType                      = "dcim.rearport"
	RearPortObjectTypeName                  = "Rear Port"
	RegionObjectType                        = "dcim.region"
	RegionObjectTypeName                    = "Region"
	RoleObjectType                          = "ipam.role"
	RoleObjectTypeName                      = "Role"
	RouteTargetObjectType                   = "ipam.routetarget"
	RouteTargetObjectTypeName               = "Route Target"
	ServiceObjectType                       = "ipam.service"
	ServiceObjectTypeName                   = "Service"
	SiteObjectType                          = "dcim.site"
	SiteObjectTypeName                      = "Site"
	SiteGroupObjectType                     = "dcim.sitegroup"
	SiteGroupObjectTypeName                 = "Site Group"
	TagObjectType                           = "extras.tag"
	TagObjectTypeName                       = "Tag"
	TenantObjectType                        = "tenancy.tenant"
	TenantObjectTypeName                    = "Tenant"
	TenantGroupObjectType                   = "tenancy.tenantgroup"
	TenantGroupObjectTypeName               = "Tenant Group"
	TunnelObjectType                        = "vpn.tunnel"
	TunnelObjectTypeName                    = "Tunnel"
	TunnelGroupObjectType                   = "vpn.tunnelgroup"
	TunnelGroupObjectTypeName               = "Tunnel Group"
	TunnelTerminationObjectType             = "vpn.tunneltermination"
	TunnelTerminationObjectTypeName         = "Tunnel Termination"
	VLANObjectType                          = "ipam.vlan"
	VLANObjectTypeName                      = "VLAN"
	VLANGroupObjectType                     = "ipam.vlangroup"
	VLANGroupObjectTypeName                 = "VLAN Group"
	VLANTranslationPolicyObjectType         = "ipam.vlantranslationpolicy"
	VLANTranslationPolicyObjectTypeName     = "VLAN Translation Policy"
	VLANTranslationRuleObjectType           = "ipam.vlantranslationrule"
	VLANTranslationRuleObjectTypeName       = "VLAN Translation Rule"
	VMInterfaceObjectType                   = "virtualization.vminterface"
	VMInterfaceObjectTypeName               = "VM Interface"
	VRFObjectType                           = "ipam.vrf"
	VRFObjectTypeName                       = "VRF"
	VirtualChassisObjectType                = "dcim.virtualchassis"
	VirtualChassisObjectTypeName            = "Virtual Chassis"
	VirtualCircuitObjectType                = "circuits.virtualcircuit"
	VirtualCircuitObjectTypeName            = "Virtual Circuit"
	VirtualCircuitTerminationObjectType     = "circuits.virtualcircuittermination"
	VirtualCircuitTerminationObjectTypeName = "Virtual Circuit Termination"
	VirtualCircuitTypeObjectType            = "circuits.virtualcircuittype"
	VirtualCircuitTypeObjectTypeName        = "Virtual Circuit Type"
	VirtualDeviceContextObjectType          = "dcim.virtualdevicecontext"
	VirtualDeviceContextObjectTypeName      = "Virtual Device Context"
	VirtualDiskObjectType                   = "virtualization.virtualdisk"
	VirtualDiskObjectTypeName               = "Virtual Disk"
	VirtualMachineObjectType                = "virtualization.virtualmachine"
	VirtualMachineObjectTypeName            = "Virtual Machine"
	WirelessLANObjectType                   = "wireless.wirelesslan"
	WirelessLANObjectTypeName               = "Wireless LAN"
	WirelessLANGroupObjectType              = "wireless.wirelesslangroup"
	WirelessLANGroupObjectTypeName          = "Wireless LAN Group"
	WirelessLinkObjectType                  = "wireless.wirelesslink"
	WirelessLinkObjectTypeName              = "Wireless Link"
	CustomFieldObjectType                   = "extras.customfield"
	CustomFieldObjectTypeName               = "Custom Field"
	CustomFieldChoiceSetObjectType          = "extras.customfieldchoiceset"
	CustomFieldChoiceSetObjectTypeName      = "Custom Field Choice Set"
	JournalEntryObjectType                  = "extras.journalentry"
	JournalEntryObjectTypeName              = "Journal Entry"
	ModuleTypeProfileObjectType             = "dcim.moduletypeprofile"
	ModuleTypeProfileObjectTypeName         = "Module Type Profile"
	CustomLinkObjectType                    = "extras.customlink"
	CustomLinkObjectTypeName                = "Custom Link"
)

func GetObjectType(entity *pb.Entity) (string, error) {
	switch entity.GetEntity().(type) {
	case *pb.Entity_Asn:
		return ASNObjectType, nil
	case *pb.Entity_AsnRange:
		return ASNRangeObjectType, nil
	case *pb.Entity_Aggregate:
		return AggregateObjectType, nil
	case *pb.Entity_Cable:
		return CableObjectType, nil
	case *pb.Entity_CablePath:
		return CablePathObjectType, nil
	case *pb.Entity_CableTermination:
		return CableTerminationObjectType, nil
	case *pb.Entity_Circuit:
		return CircuitObjectType, nil
	case *pb.Entity_CircuitGroup:
		return CircuitGroupObjectType, nil
	case *pb.Entity_CircuitGroupAssignment:
		return CircuitGroupAssignmentObjectType, nil
	case *pb.Entity_CircuitTermination:
		return CircuitTerminationObjectType, nil
	case *pb.Entity_CircuitType:
		return CircuitTypeObjectType, nil
	case *pb.Entity_Cluster:
		return ClusterObjectType, nil
	case *pb.Entity_ClusterGroup:
		return ClusterGroupObjectType, nil
	case *pb.Entity_ClusterType:
		return ClusterTypeObjectType, nil
	case *pb.Entity_ConsolePort:
		return ConsolePortObjectType, nil
	case *pb.Entity_ConsoleServerPort:
		return ConsoleServerPortObjectType, nil
	case *pb.Entity_Contact:
		return ContactObjectType, nil
	case *pb.Entity_ContactAssignment:
		return ContactAssignmentObjectType, nil
	case *pb.Entity_ContactGroup:
		return ContactGroupObjectType, nil
	case *pb.Entity_ContactRole:
		return ContactRoleObjectType, nil
	case *pb.Entity_Device:
		return DeviceObjectType, nil
	case *pb.Entity_DeviceBay:
		return DeviceBayObjectType, nil
	case *pb.Entity_DeviceRole:
		return DeviceRoleObjectType, nil
	case *pb.Entity_DeviceType:
		return DeviceTypeObjectType, nil
	case *pb.Entity_FhrpGroup:
		return FHRPGroupObjectType, nil
	case *pb.Entity_FhrpGroupAssignment:
		return FHRPGroupAssignmentObjectType, nil
	case *pb.Entity_FrontPort:
		return FrontPortObjectType, nil
	case *pb.Entity_IkePolicy:
		return IKEPolicyObjectType, nil
	case *pb.Entity_IkeProposal:
		return IKEProposalObjectType, nil
	case *pb.Entity_IpAddress:
		return IPAddressObjectType, nil
	case *pb.Entity_IpRange:
		return IPRangeObjectType, nil
	case *pb.Entity_IpSecPolicy:
		return IPSecPolicyObjectType, nil
	case *pb.Entity_IpSecProfile:
		return IPSecProfileObjectType, nil
	case *pb.Entity_IpSecProposal:
		return IPSecProposalObjectType, nil
	case *pb.Entity_Interface:
		return InterfaceObjectType, nil
	case *pb.Entity_InventoryItem:
		return InventoryItemObjectType, nil
	case *pb.Entity_InventoryItemRole:
		return InventoryItemRoleObjectType, nil
	case *pb.Entity_L2Vpn:
		return L2VPNObjectType, nil
	case *pb.Entity_L2VpnTermination:
		return L2VPNTerminationObjectType, nil
	case *pb.Entity_Location:
		return LocationObjectType, nil
	case *pb.Entity_MacAddress:
		return MACAddressObjectType, nil
	case *pb.Entity_Manufacturer:
		return ManufacturerObjectType, nil
	case *pb.Entity_Module:
		return ModuleObjectType, nil
	case *pb.Entity_ModuleBay:
		return ModuleBayObjectType, nil
	case *pb.Entity_ModuleType:
		return ModuleTypeObjectType, nil
	case *pb.Entity_Platform:
		return PlatformObjectType, nil
	case *pb.Entity_PowerFeed:
		return PowerFeedObjectType, nil
	case *pb.Entity_PowerOutlet:
		return PowerOutletObjectType, nil
	case *pb.Entity_PowerPanel:
		return PowerPanelObjectType, nil
	case *pb.Entity_PowerPort:
		return PowerPortObjectType, nil
	case *pb.Entity_Prefix:
		return PrefixObjectType, nil
	case *pb.Entity_Provider:
		return ProviderObjectType, nil
	case *pb.Entity_ProviderAccount:
		return ProviderAccountObjectType, nil
	case *pb.Entity_ProviderNetwork:
		return ProviderNetworkObjectType, nil
	case *pb.Entity_Rir:
		return RIRObjectType, nil
	case *pb.Entity_Rack:
		return RackObjectType, nil
	case *pb.Entity_RackReservation:
		return RackReservationObjectType, nil
	case *pb.Entity_RackRole:
		return RackRoleObjectType, nil
	case *pb.Entity_RackType:
		return RackTypeObjectType, nil
	case *pb.Entity_RearPort:
		return RearPortObjectType, nil
	case *pb.Entity_Region:
		return RegionObjectType, nil
	case *pb.Entity_Role:
		return RoleObjectType, nil
	case *pb.Entity_RouteTarget:
		return RouteTargetObjectType, nil
	case *pb.Entity_Service:
		return ServiceObjectType, nil
	case *pb.Entity_Site:
		return SiteObjectType, nil
	case *pb.Entity_SiteGroup:
		return SiteGroupObjectType, nil
	case *pb.Entity_Tag:
		return TagObjectType, nil
	case *pb.Entity_Tenant:
		return TenantObjectType, nil
	case *pb.Entity_TenantGroup:
		return TenantGroupObjectType, nil
	case *pb.Entity_Tunnel:
		return TunnelObjectType, nil
	case *pb.Entity_TunnelGroup:
		return TunnelGroupObjectType, nil
	case *pb.Entity_TunnelTermination:
		return TunnelTerminationObjectType, nil
	case *pb.Entity_Vlan:
		return VLANObjectType, nil
	case *pb.Entity_VlanGroup:
		return VLANGroupObjectType, nil
	case *pb.Entity_VlanTranslationPolicy:
		return VLANTranslationPolicyObjectType, nil
	case *pb.Entity_VlanTranslationRule:
		return VLANTranslationRuleObjectType, nil
	case *pb.Entity_VmInterface:
		return VMInterfaceObjectType, nil
	case *pb.Entity_Vrf:
		return VRFObjectType, nil
	case *pb.Entity_VirtualChassis:
		return VirtualChassisObjectType, nil
	case *pb.Entity_VirtualCircuit:
		return VirtualCircuitObjectType, nil
	case *pb.Entity_VirtualCircuitTermination:
		return VirtualCircuitTerminationObjectType, nil
	case *pb.Entity_VirtualCircuitType:
		return VirtualCircuitTypeObjectType, nil
	case *pb.Entity_VirtualDeviceContext:
		return VirtualDeviceContextObjectType, nil
	case *pb.Entity_VirtualDisk:
		return VirtualDiskObjectType, nil
	case *pb.Entity_VirtualMachine:
		return VirtualMachineObjectType, nil
	case *pb.Entity_WirelessLan:
		return WirelessLANObjectType, nil
	case *pb.Entity_WirelessLanGroup:
		return WirelessLANGroupObjectType, nil
	case *pb.Entity_WirelessLink:
		return WirelessLinkObjectType, nil
	case *pb.Entity_CustomField:
		return CustomFieldObjectType, nil
	case *pb.Entity_CustomFieldChoiceSet:
		return CustomFieldChoiceSetObjectType, nil
	case *pb.Entity_JournalEntry:
		return JournalEntryObjectType, nil
	case *pb.Entity_ModuleTypeProfile:
		return ModuleTypeProfileObjectType, nil
	case *pb.Entity_CustomLink:
		return CustomLinkObjectType, nil
	default:
		return "", fmt.Errorf("unknown entity type: %v", entity.GetEntity())
	}
}

func GetObjectTypeName(objectType string) (string, error) {
	switch objectType {
	case ASNObjectType:
		return ASNObjectTypeName, nil
	case ASNRangeObjectType:
		return ASNRangeObjectTypeName, nil
	case AggregateObjectType:
		return AggregateObjectTypeName, nil
	case CableObjectType:
		return CableObjectTypeName, nil
	case CablePathObjectType:
		return CablePathObjectTypeName, nil
	case CableTerminationObjectType:
		return CableTerminationObjectTypeName, nil
	case CircuitObjectType:
		return CircuitObjectTypeName, nil
	case CircuitGroupObjectType:
		return CircuitGroupObjectTypeName, nil
	case CircuitGroupAssignmentObjectType:
		return CircuitGroupAssignmentObjectTypeName, nil
	case CircuitTerminationObjectType:
		return CircuitTerminationObjectTypeName, nil
	case CircuitTypeObjectType:
		return CircuitTypeObjectTypeName, nil
	case ClusterObjectType:
		return ClusterObjectTypeName, nil
	case ClusterGroupObjectType:
		return ClusterGroupObjectTypeName, nil
	case ClusterTypeObjectType:
		return ClusterTypeObjectTypeName, nil
	case ConsolePortObjectType:
		return ConsolePortObjectTypeName, nil
	case ConsoleServerPortObjectType:
		return ConsoleServerPortObjectTypeName, nil
	case ContactObjectType:
		return ContactObjectTypeName, nil
	case ContactAssignmentObjectType:
		return ContactAssignmentObjectTypeName, nil
	case ContactGroupObjectType:
		return ContactGroupObjectTypeName, nil
	case ContactRoleObjectType:
		return ContactRoleObjectTypeName, nil
	case DeviceObjectType:
		return DeviceObjectTypeName, nil
	case DeviceBayObjectType:
		return DeviceBayObjectTypeName, nil
	case DeviceRoleObjectType:
		return DeviceRoleObjectTypeName, nil
	case DeviceTypeObjectType:
		return DeviceTypeObjectTypeName, nil
	case FHRPGroupObjectType:
		return FHRPGroupObjectTypeName, nil
	case FHRPGroupAssignmentObjectType:
		return FHRPGroupAssignmentObjectTypeName, nil
	case FrontPortObjectType:
		return FrontPortObjectTypeName, nil
	case IKEPolicyObjectType:
		return IKEPolicyObjectTypeName, nil
	case IKEProposalObjectType:
		return IKEProposalObjectTypeName, nil
	case IPAddressObjectType:
		return IPAddressObjectTypeName, nil
	case IPRangeObjectType:
		return IPRangeObjectTypeName, nil
	case IPSecPolicyObjectType:
		return IPSecPolicyObjectTypeName, nil
	case IPSecProfileObjectType:
		return IPSecProfileObjectTypeName, nil
	case IPSecProposalObjectType:
		return IPSecProposalObjectTypeName, nil
	case InterfaceObjectType:
		return InterfaceObjectTypeName, nil
	case InventoryItemObjectType:
		return InventoryItemObjectTypeName, nil
	case InventoryItemRoleObjectType:
		return InventoryItemRoleObjectTypeName, nil
	case L2VPNObjectType:
		return L2VPNObjectTypeName, nil
	case L2VPNTerminationObjectType:
		return L2VPNTerminationObjectTypeName, nil
	case LocationObjectType:
		return LocationObjectTypeName, nil
	case MACAddressObjectType:
		return MACAddressObjectTypeName, nil
	case ManufacturerObjectType:
		return ManufacturerObjectTypeName, nil
	case ModuleObjectType:
		return ModuleObjectTypeName, nil
	case ModuleBayObjectType:
		return ModuleBayObjectTypeName, nil
	case ModuleTypeObjectType:
		return ModuleTypeObjectTypeName, nil
	case PlatformObjectType:
		return PlatformObjectTypeName, nil
	case PowerFeedObjectType:
		return PowerFeedObjectTypeName, nil
	case PowerOutletObjectType:
		return PowerOutletObjectTypeName, nil
	case PowerPanelObjectType:
		return PowerPanelObjectTypeName, nil
	case PowerPortObjectType:
		return PowerPortObjectTypeName, nil
	case PrefixObjectType:
		return PrefixObjectTypeName, nil
	case ProviderObjectType:
		return ProviderObjectTypeName, nil
	case ProviderAccountObjectType:
		return ProviderAccountObjectTypeName, nil
	case ProviderNetworkObjectType:
		return ProviderNetworkObjectTypeName, nil
	case RIRObjectType:
		return RIRObjectTypeName, nil
	case RackObjectType:
		return RackObjectTypeName, nil
	case RackReservationObjectType:
		return RackReservationObjectTypeName, nil
	case RackRoleObjectType:
		return RackRoleObjectTypeName, nil
	case RackTypeObjectType:
		return RackTypeObjectTypeName, nil
	case RearPortObjectType:
		return RearPortObjectTypeName, nil
	case RegionObjectType:
		return RegionObjectTypeName, nil
	case RoleObjectType:
		return RoleObjectTypeName, nil
	case RouteTargetObjectType:
		return RouteTargetObjectTypeName, nil
	case ServiceObjectType:
		return ServiceObjectTypeName, nil
	case SiteObjectType:
		return SiteObjectTypeName, nil
	case SiteGroupObjectType:
		return SiteGroupObjectTypeName, nil
	case TagObjectType:
		return TagObjectTypeName, nil
	case TenantObjectType:
		return TenantObjectTypeName, nil
	case TenantGroupObjectType:
		return TenantGroupObjectTypeName, nil
	case TunnelObjectType:
		return TunnelObjectTypeName, nil
	case TunnelGroupObjectType:
		return TunnelGroupObjectTypeName, nil
	case TunnelTerminationObjectType:
		return TunnelTerminationObjectTypeName, nil
	case VLANObjectType:
		return VLANObjectTypeName, nil
	case VLANGroupObjectType:
		return VLANGroupObjectTypeName, nil
	case VLANTranslationPolicyObjectType:
		return VLANTranslationPolicyObjectTypeName, nil
	case VLANTranslationRuleObjectType:
		return VLANTranslationRuleObjectTypeName, nil
	case VMInterfaceObjectType:
		return VMInterfaceObjectTypeName, nil
	case VRFObjectType:
		return VRFObjectTypeName, nil
	case VirtualChassisObjectType:
		return VirtualChassisObjectTypeName, nil
	case VirtualCircuitObjectType:
		return VirtualCircuitObjectTypeName, nil
	case VirtualCircuitTerminationObjectType:
		return VirtualCircuitTerminationObjectTypeName, nil
	case VirtualCircuitTypeObjectType:
		return VirtualCircuitTypeObjectTypeName, nil
	case VirtualDeviceContextObjectType:
		return VirtualDeviceContextObjectTypeName, nil
	case VirtualDiskObjectType:
		return VirtualDiskObjectTypeName, nil
	case VirtualMachineObjectType:
		return VirtualMachineObjectTypeName, nil
	case WirelessLANObjectType:
		return WirelessLANObjectTypeName, nil
	case WirelessLANGroupObjectType:
		return WirelessLANGroupObjectTypeName, nil
	case WirelessLinkObjectType:
		return WirelessLinkObjectTypeName, nil
	case CustomFieldObjectType:
		return CustomFieldObjectTypeName, nil
	case CustomFieldChoiceSetObjectType:
		return CustomFieldChoiceSetObjectTypeName, nil
	case JournalEntryObjectType:
		return JournalEntryObjectTypeName, nil
	case ModuleTypeProfileObjectType:
		return ModuleTypeProfileObjectTypeName, nil
	case CustomLinkObjectType:
		return CustomLinkObjectTypeName, nil
	default:
		return "", fmt.Errorf("unknown object type: %v", objectType)
	}
}

func GetPrimaryValue(entity *pb.Entity) (string, error) {
	switch e := entity.GetEntity().(type) {
	case *pb.Entity_Asn:
		return fmt.Sprintf("%v", e.Asn.Asn), nil
	case *pb.Entity_AsnRange:
		return fmt.Sprintf("%v", e.AsnRange.Name), nil
	case *pb.Entity_Circuit:
		return fmt.Sprintf("%v", e.Circuit.Cid), nil
	case *pb.Entity_CircuitGroup:
		return fmt.Sprintf("%v", e.CircuitGroup.Name), nil
	case *pb.Entity_CircuitType:
		return fmt.Sprintf("%v", e.CircuitType.Name), nil
	case *pb.Entity_Cluster:
		return fmt.Sprintf("%v", e.Cluster.Name), nil
	case *pb.Entity_ClusterGroup:
		return fmt.Sprintf("%v", e.ClusterGroup.Name), nil
	case *pb.Entity_ClusterType:
		return fmt.Sprintf("%v", e.ClusterType.Name), nil
	case *pb.Entity_ConsolePort:
		return fmt.Sprintf("%v", e.ConsolePort.Name), nil
	case *pb.Entity_ConsoleServerPort:
		return fmt.Sprintf("%v", e.ConsoleServerPort.Name), nil
	case *pb.Entity_Contact:
		return fmt.Sprintf("%v", e.Contact.Name), nil
	case *pb.Entity_ContactGroup:
		return fmt.Sprintf("%v", e.ContactGroup.Name), nil
	case *pb.Entity_ContactRole:
		return fmt.Sprintf("%v", e.ContactRole.Name), nil
	case *pb.Entity_CustomField:
		return fmt.Sprintf("%v", e.CustomField.Name), nil
	case *pb.Entity_CustomFieldChoiceSet:
		return fmt.Sprintf("%v", e.CustomFieldChoiceSet.Name), nil
	case *pb.Entity_CustomLink:
		return fmt.Sprintf("%v", e.CustomLink.Name), nil
	case *pb.Entity_Device:
		return fmt.Sprintf("%v", e.Device.Name), nil
	case *pb.Entity_DeviceBay:
		return fmt.Sprintf("%v", e.DeviceBay.Name), nil
	case *pb.Entity_DeviceRole:
		return fmt.Sprintf("%v", e.DeviceRole.Name), nil
	case *pb.Entity_DeviceType:
		return fmt.Sprintf("%v", e.DeviceType.Model), nil
	case *pb.Entity_FhrpGroup:
		return fmt.Sprintf("%v", e.FhrpGroup.Name), nil
	case *pb.Entity_FrontPort:
		return fmt.Sprintf("%v", e.FrontPort.Name), nil
	case *pb.Entity_IkePolicy:
		return fmt.Sprintf("%v", e.IkePolicy.Name), nil
	case *pb.Entity_IkeProposal:
		return fmt.Sprintf("%v", e.IkeProposal.Name), nil
	case *pb.Entity_IpAddress:
		return fmt.Sprintf("%v", e.IpAddress.Address), nil
	case *pb.Entity_IpSecPolicy:
		return fmt.Sprintf("%v", e.IpSecPolicy.Name), nil
	case *pb.Entity_IpSecProfile:
		return fmt.Sprintf("%v", e.IpSecProfile.Name), nil
	case *pb.Entity_IpSecProposal:
		return fmt.Sprintf("%v", e.IpSecProposal.Name), nil
	case *pb.Entity_Interface:
		return fmt.Sprintf("%v", e.Interface.Name), nil
	case *pb.Entity_InventoryItem:
		return fmt.Sprintf("%v", e.InventoryItem.Name), nil
	case *pb.Entity_InventoryItemRole:
		return fmt.Sprintf("%v", e.InventoryItemRole.Name), nil
	case *pb.Entity_L2Vpn:
		return fmt.Sprintf("%v", e.L2Vpn.Name), nil
	case *pb.Entity_Location:
		return fmt.Sprintf("%v", e.Location.Name), nil
	case *pb.Entity_MacAddress:
		return fmt.Sprintf("%v", e.MacAddress.MacAddress), nil
	case *pb.Entity_Manufacturer:
		return fmt.Sprintf("%v", e.Manufacturer.Name), nil
	case *pb.Entity_ModuleBay:
		return fmt.Sprintf("%v", e.ModuleBay.Name), nil
	case *pb.Entity_ModuleType:
		return fmt.Sprintf("%v", e.ModuleType.Model), nil
	case *pb.Entity_ModuleTypeProfile:
		return fmt.Sprintf("%v", e.ModuleTypeProfile.Name), nil
	case *pb.Entity_Platform:
		return fmt.Sprintf("%v", e.Platform.Name), nil
	case *pb.Entity_PowerFeed:
		return fmt.Sprintf("%v", e.PowerFeed.Name), nil
	case *pb.Entity_PowerOutlet:
		return fmt.Sprintf("%v", e.PowerOutlet.Name), nil
	case *pb.Entity_PowerPanel:
		return fmt.Sprintf("%v", e.PowerPanel.Name), nil
	case *pb.Entity_PowerPort:
		return fmt.Sprintf("%v", e.PowerPort.Name), nil
	case *pb.Entity_Prefix:
		return fmt.Sprintf("%v", e.Prefix.Prefix), nil
	case *pb.Entity_Provider:
		return fmt.Sprintf("%v", e.Provider.Name), nil
	case *pb.Entity_ProviderAccount:
		return fmt.Sprintf("%v", e.ProviderAccount.Name), nil
	case *pb.Entity_ProviderNetwork:
		return fmt.Sprintf("%v", e.ProviderNetwork.Name), nil
	case *pb.Entity_Rir:
		return fmt.Sprintf("%v", e.Rir.Name), nil
	case *pb.Entity_Rack:
		return fmt.Sprintf("%v", e.Rack.Name), nil
	case *pb.Entity_RackRole:
		return fmt.Sprintf("%v", e.RackRole.Name), nil
	case *pb.Entity_RackType:
		return fmt.Sprintf("%v", e.RackType.Model), nil
	case *pb.Entity_RearPort:
		return fmt.Sprintf("%v", e.RearPort.Name), nil
	case *pb.Entity_Region:
		return fmt.Sprintf("%v", e.Region.Name), nil
	case *pb.Entity_Role:
		return fmt.Sprintf("%v", e.Role.Name), nil
	case *pb.Entity_RouteTarget:
		return fmt.Sprintf("%v", e.RouteTarget.Name), nil
	case *pb.Entity_Service:
		return fmt.Sprintf("%v", e.Service.Name), nil
	case *pb.Entity_Site:
		return fmt.Sprintf("%v", e.Site.Name), nil
	case *pb.Entity_SiteGroup:
		return fmt.Sprintf("%v", e.SiteGroup.Name), nil
	case *pb.Entity_Tag:
		return fmt.Sprintf("%v", e.Tag.Name), nil
	case *pb.Entity_Tenant:
		return fmt.Sprintf("%v", e.Tenant.Name), nil
	case *pb.Entity_TenantGroup:
		return fmt.Sprintf("%v", e.TenantGroup.Name), nil
	case *pb.Entity_Tunnel:
		return fmt.Sprintf("%v", e.Tunnel.Name), nil
	case *pb.Entity_TunnelGroup:
		return fmt.Sprintf("%v", e.TunnelGroup.Name), nil
	case *pb.Entity_Vlan:
		return fmt.Sprintf("%v", e.Vlan.Name), nil
	case *pb.Entity_VlanGroup:
		return fmt.Sprintf("%v", e.VlanGroup.Name), nil
	case *pb.Entity_VlanTranslationPolicy:
		return fmt.Sprintf("%v", e.VlanTranslationPolicy.Name), nil
	case *pb.Entity_VmInterface:
		return fmt.Sprintf("%v", e.VmInterface.Name), nil
	case *pb.Entity_Vrf:
		return fmt.Sprintf("%v", e.Vrf.Name), nil
	case *pb.Entity_VirtualChassis:
		return fmt.Sprintf("%v", e.VirtualChassis.Name), nil
	case *pb.Entity_VirtualCircuit:
		return fmt.Sprintf("%v", e.VirtualCircuit.Cid), nil
	case *pb.Entity_VirtualCircuitType:
		return fmt.Sprintf("%v", e.VirtualCircuitType.Name), nil
	case *pb.Entity_VirtualDeviceContext:
		return fmt.Sprintf("%v", e.VirtualDeviceContext.Name), nil
	case *pb.Entity_VirtualDisk:
		return fmt.Sprintf("%v", e.VirtualDisk.Name), nil
	case *pb.Entity_VirtualMachine:
		return fmt.Sprintf("%v", e.VirtualMachine.Name), nil
	case *pb.Entity_WirelessLan:
		return fmt.Sprintf("%v", e.WirelessLan.Ssid), nil
	case *pb.Entity_WirelessLanGroup:
		return fmt.Sprintf("%v", e.WirelessLanGroup.Name), nil
	default:
		return "", fmt.Errorf("unknown entity type: %v", entity.GetEntity())
	}
}
