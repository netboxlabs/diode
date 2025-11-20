# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [diode/v1/ingester.proto](#diode_v1_ingester-proto)
    - [ASN](#diode-v1-ASN)
    - [ASN.CustomFieldsEntry](#diode-v1-ASN-CustomFieldsEntry)
    - [ASNRange](#diode-v1-ASNRange)
    - [ASNRange.CustomFieldsEntry](#diode-v1-ASNRange-CustomFieldsEntry)
    - [Aggregate](#diode-v1-Aggregate)
    - [Aggregate.CustomFieldsEntry](#diode-v1-Aggregate-CustomFieldsEntry)
    - [Cable](#diode-v1-Cable)
    - [Cable.CustomFieldsEntry](#diode-v1-Cable-CustomFieldsEntry)
    - [CablePath](#diode-v1-CablePath)
    - [CableTermination](#diode-v1-CableTermination)
    - [Circuit](#diode-v1-Circuit)
    - [Circuit.CustomFieldsEntry](#diode-v1-Circuit-CustomFieldsEntry)
    - [CircuitGroup](#diode-v1-CircuitGroup)
    - [CircuitGroup.CustomFieldsEntry](#diode-v1-CircuitGroup-CustomFieldsEntry)
    - [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment)
    - [CircuitTermination](#diode-v1-CircuitTermination)
    - [CircuitTermination.CustomFieldsEntry](#diode-v1-CircuitTermination-CustomFieldsEntry)
    - [CircuitType](#diode-v1-CircuitType)
    - [CircuitType.CustomFieldsEntry](#diode-v1-CircuitType-CustomFieldsEntry)
    - [Cluster](#diode-v1-Cluster)
    - [Cluster.CustomFieldsEntry](#diode-v1-Cluster-CustomFieldsEntry)
    - [ClusterGroup](#diode-v1-ClusterGroup)
    - [ClusterGroup.CustomFieldsEntry](#diode-v1-ClusterGroup-CustomFieldsEntry)
    - [ClusterType](#diode-v1-ClusterType)
    - [ClusterType.CustomFieldsEntry](#diode-v1-ClusterType-CustomFieldsEntry)
    - [ConsolePort](#diode-v1-ConsolePort)
    - [ConsolePort.CustomFieldsEntry](#diode-v1-ConsolePort-CustomFieldsEntry)
    - [ConsoleServerPort](#diode-v1-ConsoleServerPort)
    - [ConsoleServerPort.CustomFieldsEntry](#diode-v1-ConsoleServerPort-CustomFieldsEntry)
    - [Contact](#diode-v1-Contact)
    - [Contact.CustomFieldsEntry](#diode-v1-Contact-CustomFieldsEntry)
    - [ContactAssignment](#diode-v1-ContactAssignment)
    - [ContactAssignment.CustomFieldsEntry](#diode-v1-ContactAssignment-CustomFieldsEntry)
    - [ContactGroup](#diode-v1-ContactGroup)
    - [ContactGroup.CustomFieldsEntry](#diode-v1-ContactGroup-CustomFieldsEntry)
    - [ContactRole](#diode-v1-ContactRole)
    - [ContactRole.CustomFieldsEntry](#diode-v1-ContactRole-CustomFieldsEntry)
    - [CustomField](#diode-v1-CustomField)
    - [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet)
    - [CustomFieldObjectReference](#diode-v1-CustomFieldObjectReference)
    - [CustomFieldValue](#diode-v1-CustomFieldValue)
    - [CustomLink](#diode-v1-CustomLink)
    - [Device](#diode-v1-Device)
    - [Device.CustomFieldsEntry](#diode-v1-Device-CustomFieldsEntry)
    - [DeviceBay](#diode-v1-DeviceBay)
    - [DeviceBay.CustomFieldsEntry](#diode-v1-DeviceBay-CustomFieldsEntry)
    - [DeviceRole](#diode-v1-DeviceRole)
    - [DeviceRole.CustomFieldsEntry](#diode-v1-DeviceRole-CustomFieldsEntry)
    - [DeviceType](#diode-v1-DeviceType)
    - [DeviceType.CustomFieldsEntry](#diode-v1-DeviceType-CustomFieldsEntry)
    - [Entity](#diode-v1-Entity)
    - [FHRPGroup](#diode-v1-FHRPGroup)
    - [FHRPGroup.CustomFieldsEntry](#diode-v1-FHRPGroup-CustomFieldsEntry)
    - [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment)
    - [FrontPort](#diode-v1-FrontPort)
    - [FrontPort.CustomFieldsEntry](#diode-v1-FrontPort-CustomFieldsEntry)
    - [GenericObject](#diode-v1-GenericObject)
    - [IKEPolicy](#diode-v1-IKEPolicy)
    - [IKEPolicy.CustomFieldsEntry](#diode-v1-IKEPolicy-CustomFieldsEntry)
    - [IKEProposal](#diode-v1-IKEProposal)
    - [IKEProposal.CustomFieldsEntry](#diode-v1-IKEProposal-CustomFieldsEntry)
    - [IPAddress](#diode-v1-IPAddress)
    - [IPAddress.CustomFieldsEntry](#diode-v1-IPAddress-CustomFieldsEntry)
    - [IPRange](#diode-v1-IPRange)
    - [IPRange.CustomFieldsEntry](#diode-v1-IPRange-CustomFieldsEntry)
    - [IPSecPolicy](#diode-v1-IPSecPolicy)
    - [IPSecPolicy.CustomFieldsEntry](#diode-v1-IPSecPolicy-CustomFieldsEntry)
    - [IPSecProfile](#diode-v1-IPSecProfile)
    - [IPSecProfile.CustomFieldsEntry](#diode-v1-IPSecProfile-CustomFieldsEntry)
    - [IPSecProposal](#diode-v1-IPSecProposal)
    - [IPSecProposal.CustomFieldsEntry](#diode-v1-IPSecProposal-CustomFieldsEntry)
    - [IngestRequest](#diode-v1-IngestRequest)
    - [IngestResponse](#diode-v1-IngestResponse)
    - [Interface](#diode-v1-Interface)
    - [Interface.CustomFieldsEntry](#diode-v1-Interface-CustomFieldsEntry)
    - [InventoryItem](#diode-v1-InventoryItem)
    - [InventoryItem.CustomFieldsEntry](#diode-v1-InventoryItem-CustomFieldsEntry)
    - [InventoryItemRole](#diode-v1-InventoryItemRole)
    - [InventoryItemRole.CustomFieldsEntry](#diode-v1-InventoryItemRole-CustomFieldsEntry)
    - [JournalEntry](#diode-v1-JournalEntry)
    - [JournalEntry.CustomFieldsEntry](#diode-v1-JournalEntry-CustomFieldsEntry)
    - [L2VPN](#diode-v1-L2VPN)
    - [L2VPN.CustomFieldsEntry](#diode-v1-L2VPN-CustomFieldsEntry)
    - [L2VPNTermination](#diode-v1-L2VPNTermination)
    - [L2VPNTermination.CustomFieldsEntry](#diode-v1-L2VPNTermination-CustomFieldsEntry)
    - [Location](#diode-v1-Location)
    - [Location.CustomFieldsEntry](#diode-v1-Location-CustomFieldsEntry)
    - [MACAddress](#diode-v1-MACAddress)
    - [MACAddress.CustomFieldsEntry](#diode-v1-MACAddress-CustomFieldsEntry)
    - [Manufacturer](#diode-v1-Manufacturer)
    - [Manufacturer.CustomFieldsEntry](#diode-v1-Manufacturer-CustomFieldsEntry)
    - [Module](#diode-v1-Module)
    - [Module.CustomFieldsEntry](#diode-v1-Module-CustomFieldsEntry)
    - [ModuleBay](#diode-v1-ModuleBay)
    - [ModuleBay.CustomFieldsEntry](#diode-v1-ModuleBay-CustomFieldsEntry)
    - [ModuleType](#diode-v1-ModuleType)
    - [ModuleType.CustomFieldsEntry](#diode-v1-ModuleType-CustomFieldsEntry)
    - [ModuleTypeProfile](#diode-v1-ModuleTypeProfile)
    - [ModuleTypeProfile.CustomFieldsEntry](#diode-v1-ModuleTypeProfile-CustomFieldsEntry)
    - [Platform](#diode-v1-Platform)
    - [Platform.CustomFieldsEntry](#diode-v1-Platform-CustomFieldsEntry)
    - [PowerFeed](#diode-v1-PowerFeed)
    - [PowerFeed.CustomFieldsEntry](#diode-v1-PowerFeed-CustomFieldsEntry)
    - [PowerOutlet](#diode-v1-PowerOutlet)
    - [PowerOutlet.CustomFieldsEntry](#diode-v1-PowerOutlet-CustomFieldsEntry)
    - [PowerPanel](#diode-v1-PowerPanel)
    - [PowerPanel.CustomFieldsEntry](#diode-v1-PowerPanel-CustomFieldsEntry)
    - [PowerPort](#diode-v1-PowerPort)
    - [PowerPort.CustomFieldsEntry](#diode-v1-PowerPort-CustomFieldsEntry)
    - [Prefix](#diode-v1-Prefix)
    - [Prefix.CustomFieldsEntry](#diode-v1-Prefix-CustomFieldsEntry)
    - [Provider](#diode-v1-Provider)
    - [Provider.CustomFieldsEntry](#diode-v1-Provider-CustomFieldsEntry)
    - [ProviderAccount](#diode-v1-ProviderAccount)
    - [ProviderAccount.CustomFieldsEntry](#diode-v1-ProviderAccount-CustomFieldsEntry)
    - [ProviderNetwork](#diode-v1-ProviderNetwork)
    - [ProviderNetwork.CustomFieldsEntry](#diode-v1-ProviderNetwork-CustomFieldsEntry)
    - [RIR](#diode-v1-RIR)
    - [RIR.CustomFieldsEntry](#diode-v1-RIR-CustomFieldsEntry)
    - [Rack](#diode-v1-Rack)
    - [Rack.CustomFieldsEntry](#diode-v1-Rack-CustomFieldsEntry)
    - [RackReservation](#diode-v1-RackReservation)
    - [RackReservation.CustomFieldsEntry](#diode-v1-RackReservation-CustomFieldsEntry)
    - [RackRole](#diode-v1-RackRole)
    - [RackRole.CustomFieldsEntry](#diode-v1-RackRole-CustomFieldsEntry)
    - [RackType](#diode-v1-RackType)
    - [RackType.CustomFieldsEntry](#diode-v1-RackType-CustomFieldsEntry)
    - [RearPort](#diode-v1-RearPort)
    - [RearPort.CustomFieldsEntry](#diode-v1-RearPort-CustomFieldsEntry)
    - [Region](#diode-v1-Region)
    - [Region.CustomFieldsEntry](#diode-v1-Region-CustomFieldsEntry)
    - [Role](#diode-v1-Role)
    - [Role.CustomFieldsEntry](#diode-v1-Role-CustomFieldsEntry)
    - [RouteTarget](#diode-v1-RouteTarget)
    - [RouteTarget.CustomFieldsEntry](#diode-v1-RouteTarget-CustomFieldsEntry)
    - [Service](#diode-v1-Service)
    - [Service.CustomFieldsEntry](#diode-v1-Service-CustomFieldsEntry)
    - [Site](#diode-v1-Site)
    - [Site.CustomFieldsEntry](#diode-v1-Site-CustomFieldsEntry)
    - [SiteGroup](#diode-v1-SiteGroup)
    - [SiteGroup.CustomFieldsEntry](#diode-v1-SiteGroup-CustomFieldsEntry)
    - [Tag](#diode-v1-Tag)
    - [Tenant](#diode-v1-Tenant)
    - [Tenant.CustomFieldsEntry](#diode-v1-Tenant-CustomFieldsEntry)
    - [TenantGroup](#diode-v1-TenantGroup)
    - [TenantGroup.CustomFieldsEntry](#diode-v1-TenantGroup-CustomFieldsEntry)
    - [Tunnel](#diode-v1-Tunnel)
    - [Tunnel.CustomFieldsEntry](#diode-v1-Tunnel-CustomFieldsEntry)
    - [TunnelGroup](#diode-v1-TunnelGroup)
    - [TunnelGroup.CustomFieldsEntry](#diode-v1-TunnelGroup-CustomFieldsEntry)
    - [TunnelTermination](#diode-v1-TunnelTermination)
    - [TunnelTermination.CustomFieldsEntry](#diode-v1-TunnelTermination-CustomFieldsEntry)
    - [VLAN](#diode-v1-VLAN)
    - [VLAN.CustomFieldsEntry](#diode-v1-VLAN-CustomFieldsEntry)
    - [VLANGroup](#diode-v1-VLANGroup)
    - [VLANGroup.CustomFieldsEntry](#diode-v1-VLANGroup-CustomFieldsEntry)
    - [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy)
    - [VLANTranslationRule](#diode-v1-VLANTranslationRule)
    - [VMInterface](#diode-v1-VMInterface)
    - [VMInterface.CustomFieldsEntry](#diode-v1-VMInterface-CustomFieldsEntry)
    - [VRF](#diode-v1-VRF)
    - [VRF.CustomFieldsEntry](#diode-v1-VRF-CustomFieldsEntry)
    - [VirtualChassis](#diode-v1-VirtualChassis)
    - [VirtualChassis.CustomFieldsEntry](#diode-v1-VirtualChassis-CustomFieldsEntry)
    - [VirtualCircuit](#diode-v1-VirtualCircuit)
    - [VirtualCircuit.CustomFieldsEntry](#diode-v1-VirtualCircuit-CustomFieldsEntry)
    - [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination)
    - [VirtualCircuitTermination.CustomFieldsEntry](#diode-v1-VirtualCircuitTermination-CustomFieldsEntry)
    - [VirtualCircuitType](#diode-v1-VirtualCircuitType)
    - [VirtualCircuitType.CustomFieldsEntry](#diode-v1-VirtualCircuitType-CustomFieldsEntry)
    - [VirtualDeviceContext](#diode-v1-VirtualDeviceContext)
    - [VirtualDeviceContext.CustomFieldsEntry](#diode-v1-VirtualDeviceContext-CustomFieldsEntry)
    - [VirtualDisk](#diode-v1-VirtualDisk)
    - [VirtualDisk.CustomFieldsEntry](#diode-v1-VirtualDisk-CustomFieldsEntry)
    - [VirtualMachine](#diode-v1-VirtualMachine)
    - [VirtualMachine.CustomFieldsEntry](#diode-v1-VirtualMachine-CustomFieldsEntry)
    - [WirelessLAN](#diode-v1-WirelessLAN)
    - [WirelessLAN.CustomFieldsEntry](#diode-v1-WirelessLAN-CustomFieldsEntry)
    - [WirelessLANGroup](#diode-v1-WirelessLANGroup)
    - [WirelessLANGroup.CustomFieldsEntry](#diode-v1-WirelessLANGroup-CustomFieldsEntry)
    - [WirelessLink](#diode-v1-WirelessLink)
    - [WirelessLink.CustomFieldsEntry](#diode-v1-WirelessLink-CustomFieldsEntry)
  
    - [File-level Extensions](#diode_v1_ingester-proto-extensions)
  
    - [IngesterService](#diode-v1-IngesterService)
  
- [diode/v1/reconciler.proto](#diode_v1_reconciler-proto)
    - [Change](#diode-v1-Change)
    - [ChangeSet](#diode-v1-ChangeSet)
    - [Deviation](#diode-v1-Deviation)
    - [DeviationError](#diode-v1-DeviationError)
    - [IngestionLog](#diode-v1-IngestionLog)
    - [IngestionMetrics](#diode-v1-IngestionMetrics)
    - [RetrieveDeviationByIDRequest](#diode-v1-RetrieveDeviationByIDRequest)
    - [RetrieveDeviationByIDResponse](#diode-v1-RetrieveDeviationByIDResponse)
    - [RetrieveDeviationsRequest](#diode-v1-RetrieveDeviationsRequest)
    - [RetrieveDeviationsResponse](#diode-v1-RetrieveDeviationsResponse)
    - [RetrieveIngestionLogsRequest](#diode-v1-RetrieveIngestionLogsRequest)
    - [RetrieveIngestionLogsResponse](#diode-v1-RetrieveIngestionLogsResponse)
  
    - [State](#diode-v1-State)
  
    - [ReconcilerService](#diode-v1-ReconcilerService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="diode_v1_ingester-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/ingester.proto



<a name="diode-v1-ASN"></a>

### ASN



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| asn | [int64](#int64) |  |  |
| rir | [RIR](#diode-v1-RIR) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ASN.CustomFieldsEntry](#diode-v1-ASN-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ASN-CustomFieldsEntry"></a>

### ASN.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ASNRange"></a>

### ASNRange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| rir | [RIR](#diode-v1-RIR) |  |  |
| start | [int64](#int64) |  |  |
| end | [int64](#int64) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ASNRange.CustomFieldsEntry](#diode-v1-ASNRange-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ASNRange-CustomFieldsEntry"></a>

### ASNRange.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Aggregate"></a>

### Aggregate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prefix | [string](#string) |  |  |
| rir | [RIR](#diode-v1-RIR) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| date_added | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Aggregate.CustomFieldsEntry](#diode-v1-Aggregate-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Aggregate-CustomFieldsEntry"></a>

### Aggregate.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Cable"></a>

### Cable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) | optional |  |
| a_terminations | [GenericObject](#diode-v1-GenericObject) | repeated |  |
| b_terminations | [GenericObject](#diode-v1-GenericObject) | repeated |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| label | [string](#string) | optional |  |
| color | [string](#string) | optional |  |
| length | [double](#double) | optional |  |
| length_unit | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Cable.CustomFieldsEntry](#diode-v1-Cable-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Cable-CustomFieldsEntry"></a>

### Cable.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-CablePath"></a>

### CablePath



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| is_active | [bool](#bool) | optional |  |
| is_complete | [bool](#bool) | optional |  |
| is_split | [bool](#bool) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CableTermination"></a>

### CableTermination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cable | [Cable](#diode-v1-Cable) |  |  |
| cable_end | [string](#string) |  |  |
| termination_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| termination_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| termination_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| termination_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| termination_interface | [Interface](#diode-v1-Interface) |  |  |
| termination_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| termination_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| termination_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| termination_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Circuit"></a>

### Circuit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cid | [string](#string) |  |  |
| provider | [Provider](#diode-v1-Provider) |  |  |
| provider_account | [ProviderAccount](#diode-v1-ProviderAccount) | optional |  |
| type | [CircuitType](#diode-v1-CircuitType) |  |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| install_date | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional |  |
| termination_date | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional |  |
| commit_rate | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| distance | [double](#double) | optional |  |
| distance_unit | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| assignments | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) | repeated |  |
| custom_fields | [Circuit.CustomFieldsEntry](#diode-v1-Circuit-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Circuit-CustomFieldsEntry"></a>

### Circuit.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-CircuitGroup"></a>

### CircuitGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [CircuitGroup.CustomFieldsEntry](#diode-v1-CircuitGroup-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CircuitGroup-CustomFieldsEntry"></a>

### CircuitGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-CircuitGroupAssignment"></a>

### CircuitGroupAssignment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| member_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| member_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| priority | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CircuitTermination"></a>

### CircuitTermination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| circuit | [Circuit](#diode-v1-Circuit) |  |  |
| term_side | [string](#string) |  |  |
| termination_location | [Location](#diode-v1-Location) |  |  |
| termination_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| termination_region | [Region](#diode-v1-Region) |  |  |
| termination_site | [Site](#diode-v1-Site) |  |  |
| termination_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| port_speed | [int64](#int64) | optional |  |
| upstream_speed | [int64](#int64) | optional |  |
| xconnect_id | [string](#string) | optional |  |
| pp_info | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [CircuitTermination.CustomFieldsEntry](#diode-v1-CircuitTermination-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CircuitTermination-CustomFieldsEntry"></a>

### CircuitTermination.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-CircuitType"></a>

### CircuitType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [CircuitType.CustomFieldsEntry](#diode-v1-CircuitType-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CircuitType-CustomFieldsEntry"></a>

### CircuitType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Cluster"></a>

### Cluster



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| type | [ClusterType](#diode-v1-ClusterType) |  |  |
| group | [ClusterGroup](#diode-v1-ClusterGroup) | optional |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| scope_location | [Location](#diode-v1-Location) |  |  |
| scope_region | [Region](#diode-v1-Region) |  |  |
| scope_site | [Site](#diode-v1-Site) |  |  |
| scope_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Cluster.CustomFieldsEntry](#diode-v1-Cluster-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Cluster-CustomFieldsEntry"></a>

### Cluster.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ClusterGroup"></a>

### ClusterGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ClusterGroup.CustomFieldsEntry](#diode-v1-ClusterGroup-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ClusterGroup-CustomFieldsEntry"></a>

### ClusterGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ClusterType"></a>

### ClusterType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ClusterType.CustomFieldsEntry](#diode-v1-ClusterType-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ClusterType-CustomFieldsEntry"></a>

### ClusterType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ConsolePort"></a>

### ConsolePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) | optional |  |
| speed | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ConsolePort.CustomFieldsEntry](#diode-v1-ConsolePort-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ConsolePort-CustomFieldsEntry"></a>

### ConsolePort.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ConsoleServerPort"></a>

### ConsoleServerPort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) | optional |  |
| speed | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ConsoleServerPort.CustomFieldsEntry](#diode-v1-ConsoleServerPort-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ConsoleServerPort-CustomFieldsEntry"></a>

### ConsoleServerPort.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Contact"></a>

### Contact



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [ContactGroup](#diode-v1-ContactGroup) | optional | **Deprecated.**  |
| name | [string](#string) |  |  |
| title | [string](#string) | optional |  |
| phone | [string](#string) | optional |  |
| email | [string](#string) | optional |  |
| address | [string](#string) | optional |  |
| link | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Contact.CustomFieldsEntry](#diode-v1-Contact-CustomFieldsEntry) | repeated |  |
| groups | [ContactGroup](#diode-v1-ContactGroup) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Contact-CustomFieldsEntry"></a>

### Contact.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ContactAssignment"></a>

### ContactAssignment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_asn | [ASN](#diode-v1-ASN) |  |  |
| object_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| object_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| object_cable | [Cable](#diode-v1-Cable) |  |  |
| object_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| object_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| object_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| object_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| object_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| object_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| object_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| object_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| object_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| object_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| object_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| object_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| object_contact | [Contact](#diode-v1-Contact) |  |  |
| object_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| object_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| object_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| object_device | [Device](#diode-v1-Device) |  |  |
| object_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| object_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| object_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| object_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| object_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| object_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| object_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| object_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| object_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| object_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| object_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| object_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| object_interface | [Interface](#diode-v1-Interface) |  |  |
| object_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| object_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| object_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| object_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| object_location | [Location](#diode-v1-Location) |  |  |
| object_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| object_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| object_module | [Module](#diode-v1-Module) |  |  |
| object_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| object_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| object_platform | [Platform](#diode-v1-Platform) |  |  |
| object_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| object_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| object_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| object_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| object_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| object_provider | [Provider](#diode-v1-Provider) |  |  |
| object_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| object_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| object_rir | [RIR](#diode-v1-RIR) |  |  |
| object_rack | [Rack](#diode-v1-Rack) |  |  |
| object_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| object_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| object_rack_type | [RackType](#diode-v1-RackType) |  |  |
| object_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| object_region | [Region](#diode-v1-Region) |  |  |
| object_role | [Role](#diode-v1-Role) |  |  |
| object_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| object_service | [Service](#diode-v1-Service) |  |  |
| object_site | [Site](#diode-v1-Site) |  |  |
| object_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| object_tag | [Tag](#diode-v1-Tag) |  |  |
| object_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| object_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| object_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| object_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| object_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| object_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| object_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| object_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| object_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| object_vrf | [VRF](#diode-v1-VRF) |  |  |
| object_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| object_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| object_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| object_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| object_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| object_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| object_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| object_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| object_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| object_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| object_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| object_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| object_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| object_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| object_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |
| contact | [Contact](#diode-v1-Contact) |  |  |
| role | [ContactRole](#diode-v1-ContactRole) | optional |  |
| priority | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ContactAssignment.CustomFieldsEntry](#diode-v1-ContactAssignment-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ContactAssignment-CustomFieldsEntry"></a>

### ContactAssignment.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ContactGroup"></a>

### ContactGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| parent | [ContactGroup](#diode-v1-ContactGroup) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ContactGroup.CustomFieldsEntry](#diode-v1-ContactGroup-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ContactGroup-CustomFieldsEntry"></a>

### ContactGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ContactRole"></a>

### ContactRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ContactRole.CustomFieldsEntry](#diode-v1-ContactRole-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ContactRole-CustomFieldsEntry"></a>

### ContactRole.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-CustomField"></a>

### CustomField



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  |  |
| related_object_type | [string](#string) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| group_name | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| required | [bool](#bool) | optional |  |
| unique | [bool](#bool) | optional |  |
| search_weight | [int64](#int64) | optional |  |
| filter_logic | [string](#string) | optional |  |
| ui_visible | [string](#string) | optional |  |
| ui_editable | [string](#string) | optional |  |
| is_cloneable | [bool](#bool) | optional |  |
| default | [string](#string) | optional |  |
| related_object_filter | [string](#string) | optional |  |
| weight | [int64](#int64) | optional |  |
| validation_minimum | [double](#double) | optional |  |
| validation_maximum | [double](#double) | optional |  |
| validation_regex | [string](#string) | optional |  |
| choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) | optional |  |
| comments | [string](#string) | optional |  |
| object_types | [string](#string) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CustomFieldChoiceSet"></a>

### CustomFieldChoiceSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| base_choices | [string](#string) | optional |  |
| order_alphabetically | [bool](#bool) | optional |  |
| extra_choices | [string](#string) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-CustomFieldObjectReference"></a>

### CustomFieldObjectReference



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| asn | [ASN](#diode-v1-ASN) |  |  |
| asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| cable | [Cable](#diode-v1-Cable) |  |  |
| cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| circuit | [Circuit](#diode-v1-Circuit) |  |  |
| circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| cluster | [Cluster](#diode-v1-Cluster) |  |  |
| cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| contact | [Contact](#diode-v1-Contact) |  |  |
| contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| device | [Device](#diode-v1-Device) |  |  |
| device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| interface | [Interface](#diode-v1-Interface) |  |  |
| inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| location | [Location](#diode-v1-Location) |  |  |
| mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| module | [Module](#diode-v1-Module) |  |  |
| module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| platform | [Platform](#diode-v1-Platform) |  |  |
| power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| prefix | [Prefix](#diode-v1-Prefix) |  |  |
| provider | [Provider](#diode-v1-Provider) |  |  |
| provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| rir | [RIR](#diode-v1-RIR) |  |  |
| rack | [Rack](#diode-v1-Rack) |  |  |
| rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| rack_type | [RackType](#diode-v1-RackType) |  |  |
| rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| region | [Region](#diode-v1-Region) |  |  |
| role | [Role](#diode-v1-Role) |  |  |
| route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| service | [Service](#diode-v1-Service) |  |  |
| site | [Site](#diode-v1-Site) |  |  |
| site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| tag | [Tag](#diode-v1-Tag) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) |  |  |
| tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| vlan | [VLAN](#diode-v1-VLAN) |  |  |
| vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| vrf | [VRF](#diode-v1-VRF) |  |  |
| virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |






<a name="diode-v1-CustomFieldValue"></a>

### CustomFieldValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiple_selection | [string](#string) | repeated |  |
| multiple_objects | [CustomFieldObjectReference](#diode-v1-CustomFieldObjectReference) | repeated |  |
| text | [string](#string) |  |  |
| long_text | [string](#string) |  |  |
| integer | [int64](#int64) |  |  |
| decimal | [double](#double) |  |  |
| boolean | [bool](#bool) |  |  |
| date | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| datetime | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| url | [string](#string) |  |  |
| json | [string](#string) |  |  |
| selection | [string](#string) |  |  |
| object | [CustomFieldObjectReference](#diode-v1-CustomFieldObjectReference) |  |  |






<a name="diode-v1-CustomLink"></a>

### CustomLink



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| enabled | [bool](#bool) | optional |  |
| link_text | [string](#string) |  |  |
| link_url | [string](#string) |  |  |
| weight | [int64](#int64) | optional |  |
| group_name | [string](#string) | optional |  |
| button_class | [string](#string) | optional |  |
| new_window | [bool](#bool) | optional |  |
| object_types | [string](#string) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Device"></a>

### Device



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional |  |
| device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| platform | [Platform](#diode-v1-Platform) | optional |  |
| serial | [string](#string) | optional |  |
| asset_tag | [string](#string) | optional |  |
| site | [Site](#diode-v1-Site) |  |  |
| location | [Location](#diode-v1-Location) | optional |  |
| rack | [Rack](#diode-v1-Rack) | optional |  |
| position | [double](#double) | optional |  |
| face | [string](#string) | optional |  |
| latitude | [double](#double) | optional |  |
| longitude | [double](#double) | optional |  |
| status | [string](#string) | optional |  |
| airflow | [string](#string) | optional |  |
| primary_ip4 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| primary_ip6 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| oob_ip | [IPAddress](#diode-v1-IPAddress) | optional |  |
| cluster | [Cluster](#diode-v1-Cluster) | optional |  |
| virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) | optional |  |
| vc_position | [int64](#int64) | optional |  |
| vc_priority | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Device.CustomFieldsEntry](#diode-v1-Device-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Device-CustomFieldsEntry"></a>

### Device.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-DeviceBay"></a>

### DeviceBay



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| installed_device | [Device](#diode-v1-Device) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [DeviceBay.CustomFieldsEntry](#diode-v1-DeviceBay-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-DeviceBay-CustomFieldsEntry"></a>

### DeviceBay.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-DeviceRole"></a>

### DeviceRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| vm_role | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [DeviceRole.CustomFieldsEntry](#diode-v1-DeviceRole-CustomFieldsEntry) | repeated |  |
| parent | [DeviceRole](#diode-v1-DeviceRole) | optional |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-DeviceRole-CustomFieldsEntry"></a>

### DeviceRole.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-DeviceType"></a>

### DeviceType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| default_platform | [Platform](#diode-v1-Platform) | optional |  |
| model | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| part_number | [string](#string) | optional |  |
| u_height | [double](#double) | optional |  |
| exclude_from_utilization | [bool](#bool) | optional |  |
| is_full_depth | [bool](#bool) | optional |  |
| subdevice_role | [string](#string) | optional |  |
| airflow | [string](#string) | optional |  |
| weight | [double](#double) | optional |  |
| weight_unit | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [DeviceType.CustomFieldsEntry](#diode-v1-DeviceType-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-DeviceType-CustomFieldsEntry"></a>

### DeviceType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Entity"></a>

### Entity



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| asn | [ASN](#diode-v1-ASN) |  |  |
| asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| cable | [Cable](#diode-v1-Cable) |  |  |
| cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| circuit | [Circuit](#diode-v1-Circuit) |  |  |
| circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| cluster | [Cluster](#diode-v1-Cluster) |  |  |
| cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| contact | [Contact](#diode-v1-Contact) |  |  |
| contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| device | [Device](#diode-v1-Device) |  |  |
| device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| interface | [Interface](#diode-v1-Interface) |  |  |
| inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| location | [Location](#diode-v1-Location) |  |  |
| mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| module | [Module](#diode-v1-Module) |  |  |
| module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| platform | [Platform](#diode-v1-Platform) |  |  |
| power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| prefix | [Prefix](#diode-v1-Prefix) |  |  |
| provider | [Provider](#diode-v1-Provider) |  |  |
| provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| rir | [RIR](#diode-v1-RIR) |  |  |
| rack | [Rack](#diode-v1-Rack) |  |  |
| rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| rack_type | [RackType](#diode-v1-RackType) |  |  |
| rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| region | [Region](#diode-v1-Region) |  |  |
| role | [Role](#diode-v1-Role) |  |  |
| route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| service | [Service](#diode-v1-Service) |  |  |
| site | [Site](#diode-v1-Site) |  |  |
| site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| tag | [Tag](#diode-v1-Tag) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) |  |  |
| tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| vlan | [VLAN](#diode-v1-VLAN) |  |  |
| vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| vrf | [VRF](#diode-v1-VRF) |  |  |
| virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |






<a name="diode-v1-FHRPGroup"></a>

### FHRPGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) | optional |  |
| protocol | [string](#string) |  |  |
| group_id | [int64](#int64) |  |  |
| auth_type | [string](#string) | optional |  |
| auth_key | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [FHRPGroup.CustomFieldsEntry](#diode-v1-FHRPGroup-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-FHRPGroup-CustomFieldsEntry"></a>

### FHRPGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-FHRPGroupAssignment"></a>

### FHRPGroupAssignment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| interface_asn | [ASN](#diode-v1-ASN) |  |  |
| interface_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| interface_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| interface_cable | [Cable](#diode-v1-Cable) |  |  |
| interface_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| interface_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| interface_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| interface_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| interface_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| interface_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| interface_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| interface_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| interface_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| interface_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| interface_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| interface_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| interface_contact | [Contact](#diode-v1-Contact) |  |  |
| interface_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| interface_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| interface_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| interface_device | [Device](#diode-v1-Device) |  |  |
| interface_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| interface_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| interface_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| interface_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| interface_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| interface_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| interface_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| interface_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| interface_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| interface_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| interface_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| interface_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| interface_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| interface_interface | [Interface](#diode-v1-Interface) |  |  |
| interface_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| interface_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| interface_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| interface_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| interface_location | [Location](#diode-v1-Location) |  |  |
| interface_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| interface_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| interface_module | [Module](#diode-v1-Module) |  |  |
| interface_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| interface_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| interface_platform | [Platform](#diode-v1-Platform) |  |  |
| interface_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| interface_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| interface_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| interface_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| interface_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| interface_provider | [Provider](#diode-v1-Provider) |  |  |
| interface_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| interface_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| interface_rir | [RIR](#diode-v1-RIR) |  |  |
| interface_rack | [Rack](#diode-v1-Rack) |  |  |
| interface_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| interface_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| interface_rack_type | [RackType](#diode-v1-RackType) |  |  |
| interface_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| interface_region | [Region](#diode-v1-Region) |  |  |
| interface_role | [Role](#diode-v1-Role) |  |  |
| interface_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| interface_service | [Service](#diode-v1-Service) |  |  |
| interface_site | [Site](#diode-v1-Site) |  |  |
| interface_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| interface_tag | [Tag](#diode-v1-Tag) |  |  |
| interface_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| interface_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| interface_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| interface_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| interface_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| interface_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| interface_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| interface_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| interface_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| interface_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| interface_vrf | [VRF](#diode-v1-VRF) |  |  |
| interface_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| interface_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| interface_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| interface_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| interface_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| interface_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| interface_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| interface_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| interface_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| interface_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| interface_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| interface_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| interface_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| interface_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| interface_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |
| priority | [int64](#int64) |  |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-FrontPort"></a>

### FrontPort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| rear_port_position | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [FrontPort.CustomFieldsEntry](#diode-v1-FrontPort-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-FrontPort-CustomFieldsEntry"></a>

### FrontPort.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-GenericObject"></a>

### GenericObject



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_asn | [ASN](#diode-v1-ASN) |  |  |
| object_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| object_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| object_cable | [Cable](#diode-v1-Cable) |  |  |
| object_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| object_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| object_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| object_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| object_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| object_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| object_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| object_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| object_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| object_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| object_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| object_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| object_contact | [Contact](#diode-v1-Contact) |  |  |
| object_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| object_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| object_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| object_device | [Device](#diode-v1-Device) |  |  |
| object_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| object_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| object_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| object_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| object_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| object_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| object_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| object_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| object_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| object_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| object_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| object_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| object_interface | [Interface](#diode-v1-Interface) |  |  |
| object_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| object_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| object_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| object_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| object_location | [Location](#diode-v1-Location) |  |  |
| object_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| object_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| object_module | [Module](#diode-v1-Module) |  |  |
| object_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| object_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| object_platform | [Platform](#diode-v1-Platform) |  |  |
| object_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| object_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| object_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| object_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| object_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| object_provider | [Provider](#diode-v1-Provider) |  |  |
| object_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| object_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| object_rir | [RIR](#diode-v1-RIR) |  |  |
| object_rack | [Rack](#diode-v1-Rack) |  |  |
| object_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| object_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| object_rack_type | [RackType](#diode-v1-RackType) |  |  |
| object_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| object_region | [Region](#diode-v1-Region) |  |  |
| object_role | [Role](#diode-v1-Role) |  |  |
| object_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| object_service | [Service](#diode-v1-Service) |  |  |
| object_site | [Site](#diode-v1-Site) |  |  |
| object_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| object_tag | [Tag](#diode-v1-Tag) |  |  |
| object_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| object_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| object_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| object_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| object_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| object_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| object_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| object_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| object_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| object_vrf | [VRF](#diode-v1-VRF) |  |  |
| object_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| object_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| object_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| object_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| object_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| object_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| object_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| object_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| object_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| object_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| object_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| object_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| object_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| object_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| object_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |






<a name="diode-v1-IKEPolicy"></a>

### IKEPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| version | [int64](#int64) |  |  |
| mode | [string](#string) | optional |  |
| preshared_key | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IKEPolicy.CustomFieldsEntry](#diode-v1-IKEPolicy-CustomFieldsEntry) | repeated |  |
| proposals | [IKEProposal](#diode-v1-IKEProposal) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IKEPolicy-CustomFieldsEntry"></a>

### IKEPolicy.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IKEProposal"></a>

### IKEProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| authentication_method | [string](#string) |  |  |
| encryption_algorithm | [string](#string) |  |  |
| authentication_algorithm | [string](#string) | optional |  |
| group | [int64](#int64) |  |  |
| sa_lifetime | [int64](#int64) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IKEProposal.CustomFieldsEntry](#diode-v1-IKEProposal-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IKEProposal-CustomFieldsEntry"></a>

### IKEProposal.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IPAddress"></a>

### IPAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| vrf | [VRF](#diode-v1-VRF) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| status | [string](#string) | optional |  |
| role | [string](#string) | optional |  |
| assigned_object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| assigned_object_interface | [Interface](#diode-v1-Interface) |  |  |
| assigned_object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| nat_inside | [IPAddress](#diode-v1-IPAddress) | optional |  |
| dns_name | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IPAddress.CustomFieldsEntry](#diode-v1-IPAddress-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IPAddress-CustomFieldsEntry"></a>

### IPAddress.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IPRange"></a>

### IPRange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start_address | [string](#string) |  |  |
| end_address | [string](#string) |  |  |
| vrf | [VRF](#diode-v1-VRF) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| status | [string](#string) | optional |  |
| role | [Role](#diode-v1-Role) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| mark_utilized | [bool](#bool) | optional |  |
| custom_fields | [IPRange.CustomFieldsEntry](#diode-v1-IPRange-CustomFieldsEntry) | repeated |  |
| mark_populated | [bool](#bool) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IPRange-CustomFieldsEntry"></a>

### IPRange.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IPSecPolicy"></a>

### IPSecPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| pfs_group | [int64](#int64) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IPSecPolicy.CustomFieldsEntry](#diode-v1-IPSecPolicy-CustomFieldsEntry) | repeated |  |
| proposals | [IPSecProposal](#diode-v1-IPSecProposal) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IPSecPolicy-CustomFieldsEntry"></a>

### IPSecPolicy.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IPSecProfile"></a>

### IPSecProfile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| mode | [string](#string) |  |  |
| ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| ipsec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IPSecProfile.CustomFieldsEntry](#diode-v1-IPSecProfile-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IPSecProfile-CustomFieldsEntry"></a>

### IPSecProfile.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IPSecProposal"></a>

### IPSecProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| encryption_algorithm | [string](#string) | optional |  |
| authentication_algorithm | [string](#string) | optional |  |
| sa_lifetime_seconds | [int64](#int64) | optional |  |
| sa_lifetime_data | [int64](#int64) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [IPSecProposal.CustomFieldsEntry](#diode-v1-IPSecProposal-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IPSecProposal-CustomFieldsEntry"></a>

### IPSecProposal.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-IngestRequest"></a>

### IngestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream | [string](#string) |  |  |
| entities | [Entity](#diode-v1-Entity) | repeated |  |
| id | [string](#string) |  |  |
| producer_app_name | [string](#string) |  |  |
| producer_app_version | [string](#string) |  |  |
| sdk_name | [string](#string) |  |  |
| sdk_version | [string](#string) |  |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-IngestResponse"></a>

### IngestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| errors | [string](#string) | repeated |  |






<a name="diode-v1-Interface"></a>

### Interface



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) |  |  |
| enabled | [bool](#bool) | optional |  |
| parent | [Interface](#diode-v1-Interface) | optional |  |
| bridge | [Interface](#diode-v1-Interface) | optional |  |
| lag | [Interface](#diode-v1-Interface) | optional |  |
| mtu | [int64](#int64) | optional |  |
| primary_mac_address | [MACAddress](#diode-v1-MACAddress) | optional |  |
| speed | [int64](#int64) | optional |  |
| duplex | [string](#string) | optional |  |
| wwn | [string](#string) | optional |  |
| mgmt_only | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| mode | [string](#string) | optional |  |
| rf_role | [string](#string) | optional |  |
| rf_channel | [string](#string) | optional |  |
| poe_mode | [string](#string) | optional |  |
| poe_type | [string](#string) | optional |  |
| rf_channel_frequency | [double](#double) | optional |  |
| rf_channel_width | [double](#double) | optional |  |
| tx_power | [int64](#int64) | optional |  |
| untagged_vlan | [VLAN](#diode-v1-VLAN) | optional |  |
| qinq_svlan | [VLAN](#diode-v1-VLAN) | optional |  |
| vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| vrf | [VRF](#diode-v1-VRF) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Interface.CustomFieldsEntry](#diode-v1-Interface-CustomFieldsEntry) | repeated |  |
| vdcs | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) | repeated |  |
| tagged_vlans | [VLAN](#diode-v1-VLAN) | repeated |  |
| wireless_lans | [WirelessLAN](#diode-v1-WirelessLAN) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Interface-CustomFieldsEntry"></a>

### Interface.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-InventoryItem"></a>

### InventoryItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| parent | [InventoryItem](#diode-v1-InventoryItem) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| status | [string](#string) | optional |  |
| role | [InventoryItemRole](#diode-v1-InventoryItemRole) | optional |  |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) | optional |  |
| part_id | [string](#string) | optional |  |
| serial | [string](#string) | optional |  |
| asset_tag | [string](#string) | optional |  |
| discovered | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| component_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| component_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| component_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| component_interface | [Interface](#diode-v1-Interface) |  |  |
| component_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| component_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| component_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [InventoryItem.CustomFieldsEntry](#diode-v1-InventoryItem-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-InventoryItem-CustomFieldsEntry"></a>

### InventoryItem.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-InventoryItemRole"></a>

### InventoryItemRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [InventoryItemRole.CustomFieldsEntry](#diode-v1-InventoryItemRole-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-InventoryItemRole-CustomFieldsEntry"></a>

### InventoryItemRole.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-JournalEntry"></a>

### JournalEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| assigned_object_asn | [ASN](#diode-v1-ASN) |  |  |
| assigned_object_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| assigned_object_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| assigned_object_cable | [Cable](#diode-v1-Cable) |  |  |
| assigned_object_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| assigned_object_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| assigned_object_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| assigned_object_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| assigned_object_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| assigned_object_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| assigned_object_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| assigned_object_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| assigned_object_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| assigned_object_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| assigned_object_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| assigned_object_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| assigned_object_contact | [Contact](#diode-v1-Contact) |  |  |
| assigned_object_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| assigned_object_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| assigned_object_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| assigned_object_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| assigned_object_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| assigned_object_device | [Device](#diode-v1-Device) |  |  |
| assigned_object_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| assigned_object_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| assigned_object_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| assigned_object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| assigned_object_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| assigned_object_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| assigned_object_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| assigned_object_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| assigned_object_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| assigned_object_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| assigned_object_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| assigned_object_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| assigned_object_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| assigned_object_interface | [Interface](#diode-v1-Interface) |  |  |
| assigned_object_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| assigned_object_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| assigned_object_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| assigned_object_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| assigned_object_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| assigned_object_location | [Location](#diode-v1-Location) |  |  |
| assigned_object_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| assigned_object_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| assigned_object_module | [Module](#diode-v1-Module) |  |  |
| assigned_object_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| assigned_object_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| assigned_object_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| assigned_object_platform | [Platform](#diode-v1-Platform) |  |  |
| assigned_object_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| assigned_object_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| assigned_object_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| assigned_object_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| assigned_object_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| assigned_object_provider | [Provider](#diode-v1-Provider) |  |  |
| assigned_object_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| assigned_object_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| assigned_object_rir | [RIR](#diode-v1-RIR) |  |  |
| assigned_object_rack | [Rack](#diode-v1-Rack) |  |  |
| assigned_object_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| assigned_object_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| assigned_object_rack_type | [RackType](#diode-v1-RackType) |  |  |
| assigned_object_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| assigned_object_region | [Region](#diode-v1-Region) |  |  |
| assigned_object_role | [Role](#diode-v1-Role) |  |  |
| assigned_object_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| assigned_object_service | [Service](#diode-v1-Service) |  |  |
| assigned_object_site | [Site](#diode-v1-Site) |  |  |
| assigned_object_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| assigned_object_tag | [Tag](#diode-v1-Tag) |  |  |
| assigned_object_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| assigned_object_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| assigned_object_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| assigned_object_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| assigned_object_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| assigned_object_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| assigned_object_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| assigned_object_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| assigned_object_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| assigned_object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| assigned_object_vrf | [VRF](#diode-v1-VRF) |  |  |
| assigned_object_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| assigned_object_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| assigned_object_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| assigned_object_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| assigned_object_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| assigned_object_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| assigned_object_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| assigned_object_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| assigned_object_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| assigned_object_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| assigned_object_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |
| kind | [string](#string) | optional |  |
| comments | [string](#string) |  |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [JournalEntry.CustomFieldsEntry](#diode-v1-JournalEntry-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-JournalEntry-CustomFieldsEntry"></a>

### JournalEntry.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-L2VPN"></a>

### L2VPN



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [int64](#int64) | optional |  |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| type | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [L2VPN.CustomFieldsEntry](#diode-v1-L2VPN-CustomFieldsEntry) | repeated |  |
| import_targets | [RouteTarget](#diode-v1-RouteTarget) | repeated |  |
| export_targets | [RouteTarget](#diode-v1-RouteTarget) | repeated |  |
| status | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-L2VPN-CustomFieldsEntry"></a>

### L2VPN.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-L2VPNTermination"></a>

### L2VPNTermination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| assigned_object_interface | [Interface](#diode-v1-Interface) |  |  |
| assigned_object_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| assigned_object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| assigned_object_asn | [ASN](#diode-v1-ASN) |  |  |
| assigned_object_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| assigned_object_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| assigned_object_cable | [Cable](#diode-v1-Cable) |  |  |
| assigned_object_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| assigned_object_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| assigned_object_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| assigned_object_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| assigned_object_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| assigned_object_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| assigned_object_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| assigned_object_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| assigned_object_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| assigned_object_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| assigned_object_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| assigned_object_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| assigned_object_contact | [Contact](#diode-v1-Contact) |  |  |
| assigned_object_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| assigned_object_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| assigned_object_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| assigned_object_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| assigned_object_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| assigned_object_device | [Device](#diode-v1-Device) |  |  |
| assigned_object_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| assigned_object_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| assigned_object_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| assigned_object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| assigned_object_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| assigned_object_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| assigned_object_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| assigned_object_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| assigned_object_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| assigned_object_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| assigned_object_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| assigned_object_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| assigned_object_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| assigned_object_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| assigned_object_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| assigned_object_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| assigned_object_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| assigned_object_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| assigned_object_location | [Location](#diode-v1-Location) |  |  |
| assigned_object_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| assigned_object_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| assigned_object_module | [Module](#diode-v1-Module) |  |  |
| assigned_object_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| assigned_object_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| assigned_object_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| assigned_object_platform | [Platform](#diode-v1-Platform) |  |  |
| assigned_object_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| assigned_object_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| assigned_object_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| assigned_object_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| assigned_object_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| assigned_object_provider | [Provider](#diode-v1-Provider) |  |  |
| assigned_object_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| assigned_object_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| assigned_object_rir | [RIR](#diode-v1-RIR) |  |  |
| assigned_object_rack | [Rack](#diode-v1-Rack) |  |  |
| assigned_object_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| assigned_object_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| assigned_object_rack_type | [RackType](#diode-v1-RackType) |  |  |
| assigned_object_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| assigned_object_region | [Region](#diode-v1-Region) |  |  |
| assigned_object_role | [Role](#diode-v1-Role) |  |  |
| assigned_object_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| assigned_object_service | [Service](#diode-v1-Service) |  |  |
| assigned_object_site | [Site](#diode-v1-Site) |  |  |
| assigned_object_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| assigned_object_tag | [Tag](#diode-v1-Tag) |  |  |
| assigned_object_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| assigned_object_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| assigned_object_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| assigned_object_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| assigned_object_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| assigned_object_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| assigned_object_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| assigned_object_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| assigned_object_vrf | [VRF](#diode-v1-VRF) |  |  |
| assigned_object_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| assigned_object_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| assigned_object_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| assigned_object_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| assigned_object_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| assigned_object_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| assigned_object_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| assigned_object_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| assigned_object_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| assigned_object_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| assigned_object_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [L2VPNTermination.CustomFieldsEntry](#diode-v1-L2VPNTermination-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-L2VPNTermination-CustomFieldsEntry"></a>

### L2VPNTermination.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Location"></a>

### Location



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| site | [Site](#diode-v1-Site) |  |  |
| parent | [Location](#diode-v1-Location) | optional |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| facility | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Location.CustomFieldsEntry](#diode-v1-Location-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Location-CustomFieldsEntry"></a>

### Location.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-MACAddress"></a>

### MACAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mac_address | [string](#string) |  |  |
| assigned_object_interface | [Interface](#diode-v1-Interface) |  |  |
| assigned_object_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [MACAddress.CustomFieldsEntry](#diode-v1-MACAddress-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-MACAddress-CustomFieldsEntry"></a>

### MACAddress.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Manufacturer"></a>

### Manufacturer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Manufacturer.CustomFieldsEntry](#diode-v1-Manufacturer-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Manufacturer-CustomFieldsEntry"></a>

### Manufacturer.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Module"></a>

### Module



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| status | [string](#string) | optional |  |
| serial | [string](#string) | optional |  |
| asset_tag | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Module.CustomFieldsEntry](#diode-v1-Module-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Module-CustomFieldsEntry"></a>

### Module.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ModuleBay"></a>

### ModuleBay



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| installed_module | [Module](#diode-v1-Module) | optional |  |
| label | [string](#string) | optional |  |
| position | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ModuleBay.CustomFieldsEntry](#diode-v1-ModuleBay-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ModuleBay-CustomFieldsEntry"></a>

### ModuleBay.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ModuleType"></a>

### ModuleType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| model | [string](#string) |  |  |
| part_number | [string](#string) | optional |  |
| airflow | [string](#string) | optional |  |
| weight | [double](#double) | optional |  |
| weight_unit | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ModuleType.CustomFieldsEntry](#diode-v1-ModuleType-CustomFieldsEntry) | repeated |  |
| profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) | optional |  |
| attributes | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ModuleType-CustomFieldsEntry"></a>

### ModuleType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ModuleTypeProfile"></a>

### ModuleTypeProfile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| schema | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ModuleTypeProfile.CustomFieldsEntry](#diode-v1-ModuleTypeProfile-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ModuleTypeProfile-CustomFieldsEntry"></a>

### ModuleTypeProfile.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Platform"></a>

### Platform



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Platform.CustomFieldsEntry](#diode-v1-Platform-CustomFieldsEntry) | repeated |  |
| parent | [Platform](#diode-v1-Platform) | optional |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Platform-CustomFieldsEntry"></a>

### Platform.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-PowerFeed"></a>

### PowerFeed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| rack | [Rack](#diode-v1-Rack) | optional |  |
| name | [string](#string) |  |  |
| status | [string](#string) | optional |  |
| type | [string](#string) | optional |  |
| supply | [string](#string) | optional |  |
| phase | [string](#string) | optional |  |
| voltage | [int64](#int64) | optional |  |
| amperage | [int64](#int64) | optional |  |
| max_utilization | [int64](#int64) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [PowerFeed.CustomFieldsEntry](#diode-v1-PowerFeed-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-PowerFeed-CustomFieldsEntry"></a>

### PowerFeed.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-PowerOutlet"></a>

### PowerOutlet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) | optional |  |
| color | [string](#string) | optional |  |
| power_port | [PowerPort](#diode-v1-PowerPort) | optional |  |
| feed_leg | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [PowerOutlet.CustomFieldsEntry](#diode-v1-PowerOutlet-CustomFieldsEntry) | repeated |  |
| status | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-PowerOutlet-CustomFieldsEntry"></a>

### PowerOutlet.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-PowerPanel"></a>

### PowerPanel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| site | [Site](#diode-v1-Site) |  |  |
| location | [Location](#diode-v1-Location) | optional |  |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [PowerPanel.CustomFieldsEntry](#diode-v1-PowerPanel-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-PowerPanel-CustomFieldsEntry"></a>

### PowerPanel.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-PowerPort"></a>

### PowerPort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) | optional |  |
| maximum_draw | [int64](#int64) | optional |  |
| allocated_draw | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [PowerPort.CustomFieldsEntry](#diode-v1-PowerPort-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-PowerPort-CustomFieldsEntry"></a>

### PowerPort.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Prefix"></a>

### Prefix



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prefix | [string](#string) |  |  |
| vrf | [VRF](#diode-v1-VRF) | optional |  |
| scope_location | [Location](#diode-v1-Location) |  |  |
| scope_region | [Region](#diode-v1-Region) |  |  |
| scope_site | [Site](#diode-v1-Site) |  |  |
| scope_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| vlan | [VLAN](#diode-v1-VLAN) | optional |  |
| status | [string](#string) | optional |  |
| role | [Role](#diode-v1-Role) | optional |  |
| is_pool | [bool](#bool) | optional |  |
| mark_utilized | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Prefix.CustomFieldsEntry](#diode-v1-Prefix-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Prefix-CustomFieldsEntry"></a>

### Prefix.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Provider"></a>

### Provider



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Provider.CustomFieldsEntry](#diode-v1-Provider-CustomFieldsEntry) | repeated |  |
| accounts | [ProviderAccount](#diode-v1-ProviderAccount) | repeated |  |
| asns | [ASN](#diode-v1-ASN) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Provider-CustomFieldsEntry"></a>

### Provider.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ProviderAccount"></a>

### ProviderAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [Provider](#diode-v1-Provider) |  |  |
| name | [string](#string) | optional |  |
| account | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ProviderAccount.CustomFieldsEntry](#diode-v1-ProviderAccount-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ProviderAccount-CustomFieldsEntry"></a>

### ProviderAccount.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-ProviderNetwork"></a>

### ProviderNetwork



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [Provider](#diode-v1-Provider) |  |  |
| name | [string](#string) |  |  |
| service_id | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [ProviderNetwork.CustomFieldsEntry](#diode-v1-ProviderNetwork-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-ProviderNetwork-CustomFieldsEntry"></a>

### ProviderNetwork.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RIR"></a>

### RIR



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| is_private | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RIR.CustomFieldsEntry](#diode-v1-RIR-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RIR-CustomFieldsEntry"></a>

### RIR.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Rack"></a>

### Rack



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| facility_id | [string](#string) | optional |  |
| site | [Site](#diode-v1-Site) |  |  |
| location | [Location](#diode-v1-Location) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| status | [string](#string) | optional |  |
| role | [RackRole](#diode-v1-RackRole) | optional |  |
| serial | [string](#string) | optional |  |
| asset_tag | [string](#string) | optional |  |
| rack_type | [RackType](#diode-v1-RackType) | optional |  |
| form_factor | [string](#string) | optional |  |
| width | [int64](#int64) | optional |  |
| u_height | [int64](#int64) | optional |  |
| starting_unit | [int64](#int64) | optional |  |
| weight | [double](#double) | optional |  |
| max_weight | [int64](#int64) | optional |  |
| weight_unit | [string](#string) | optional |  |
| desc_units | [bool](#bool) | optional |  |
| outer_width | [int64](#int64) | optional |  |
| outer_depth | [int64](#int64) | optional |  |
| outer_unit | [string](#string) | optional |  |
| mounting_depth | [int64](#int64) | optional |  |
| airflow | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Rack.CustomFieldsEntry](#diode-v1-Rack-CustomFieldsEntry) | repeated |  |
| outer_height | [int64](#int64) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Rack-CustomFieldsEntry"></a>

### Rack.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RackReservation"></a>

### RackReservation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rack | [Rack](#diode-v1-Rack) |  |  |
| units | [int64](#int64) | repeated |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| description | [string](#string) |  |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RackReservation.CustomFieldsEntry](#diode-v1-RackReservation-CustomFieldsEntry) | repeated |  |
| status | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RackReservation-CustomFieldsEntry"></a>

### RackReservation.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RackRole"></a>

### RackRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RackRole.CustomFieldsEntry](#diode-v1-RackRole-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RackRole-CustomFieldsEntry"></a>

### RackRole.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RackType"></a>

### RackType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| model | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| form_factor | [string](#string) | optional |  |
| width | [int64](#int64) | optional |  |
| u_height | [int64](#int64) | optional |  |
| starting_unit | [int64](#int64) | optional |  |
| desc_units | [bool](#bool) | optional |  |
| outer_width | [int64](#int64) | optional |  |
| outer_depth | [int64](#int64) | optional |  |
| outer_unit | [string](#string) | optional |  |
| weight | [double](#double) | optional |  |
| max_weight | [int64](#int64) | optional |  |
| weight_unit | [string](#string) | optional |  |
| mounting_depth | [int64](#int64) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RackType.CustomFieldsEntry](#diode-v1-RackType-CustomFieldsEntry) | repeated |  |
| outer_height | [int64](#int64) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RackType-CustomFieldsEntry"></a>

### RackType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RearPort"></a>

### RearPort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) |  |  |
| module | [Module](#diode-v1-Module) | optional |  |
| name | [string](#string) |  |  |
| label | [string](#string) | optional |  |
| type | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| positions | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| mark_connected | [bool](#bool) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RearPort.CustomFieldsEntry](#diode-v1-RearPort-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RearPort-CustomFieldsEntry"></a>

### RearPort.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Region"></a>

### Region



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| parent | [Region](#diode-v1-Region) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Region.CustomFieldsEntry](#diode-v1-Region-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Region-CustomFieldsEntry"></a>

### Region.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Role"></a>

### Role



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| weight | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Role.CustomFieldsEntry](#diode-v1-Role-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Role-CustomFieldsEntry"></a>

### Role.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-RouteTarget"></a>

### RouteTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [RouteTarget.CustomFieldsEntry](#diode-v1-RouteTarget-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-RouteTarget-CustomFieldsEntry"></a>

### RouteTarget.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Service"></a>

### Service



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device | [Device](#diode-v1-Device) | optional | **Deprecated.**  |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) | optional | **Deprecated.**  |
| name | [string](#string) |  |  |
| protocol | [string](#string) | optional |  |
| ports | [int64](#int64) | repeated |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Service.CustomFieldsEntry](#diode-v1-Service-CustomFieldsEntry) | repeated |  |
| ipaddresses | [IPAddress](#diode-v1-IPAddress) | repeated |  |
| parent_object_device | [Device](#diode-v1-Device) |  |  |
| parent_object_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| parent_object_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Service-CustomFieldsEntry"></a>

### Service.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Site"></a>

### Site



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| status | [string](#string) | optional |  |
| region | [Region](#diode-v1-Region) | optional |  |
| group | [SiteGroup](#diode-v1-SiteGroup) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| facility | [string](#string) | optional |  |
| time_zone | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| physical_address | [string](#string) | optional |  |
| shipping_address | [string](#string) | optional |  |
| latitude | [double](#double) | optional |  |
| longitude | [double](#double) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Site.CustomFieldsEntry](#diode-v1-Site-CustomFieldsEntry) | repeated |  |
| asns | [ASN](#diode-v1-ASN) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Site-CustomFieldsEntry"></a>

### Site.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-SiteGroup"></a>

### SiteGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| parent | [SiteGroup](#diode-v1-SiteGroup) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [SiteGroup.CustomFieldsEntry](#diode-v1-SiteGroup-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-SiteGroup-CustomFieldsEntry"></a>

### SiteGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Tag"></a>

### Tag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| weight | [int64](#int64) | optional |  |
| object_types | [string](#string) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Tenant"></a>

### Tenant



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| group | [TenantGroup](#diode-v1-TenantGroup) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Tenant.CustomFieldsEntry](#diode-v1-Tenant-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Tenant-CustomFieldsEntry"></a>

### Tenant.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-TenantGroup"></a>

### TenantGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| parent | [TenantGroup](#diode-v1-TenantGroup) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [TenantGroup.CustomFieldsEntry](#diode-v1-TenantGroup-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-TenantGroup-CustomFieldsEntry"></a>

### TenantGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-Tunnel"></a>

### Tunnel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| status | [string](#string) |  |  |
| group | [TunnelGroup](#diode-v1-TunnelGroup) | optional |  |
| encapsulation | [string](#string) |  |  |
| ipsec_profile | [IPSecProfile](#diode-v1-IPSecProfile) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| tunnel_id | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [Tunnel.CustomFieldsEntry](#diode-v1-Tunnel-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-Tunnel-CustomFieldsEntry"></a>

### Tunnel.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-TunnelGroup"></a>

### TunnelGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [TunnelGroup.CustomFieldsEntry](#diode-v1-TunnelGroup-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-TunnelGroup-CustomFieldsEntry"></a>

### TunnelGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-TunnelTermination"></a>

### TunnelTermination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| role | [string](#string) |  |  |
| termination_asn | [ASN](#diode-v1-ASN) |  |  |
| termination_asn_range | [ASNRange](#diode-v1-ASNRange) |  |  |
| termination_aggregate | [Aggregate](#diode-v1-Aggregate) |  |  |
| termination_cable | [Cable](#diode-v1-Cable) |  |  |
| termination_cable_path | [CablePath](#diode-v1-CablePath) |  |  |
| termination_cable_termination | [CableTermination](#diode-v1-CableTermination) |  |  |
| termination_circuit | [Circuit](#diode-v1-Circuit) |  |  |
| termination_circuit_group | [CircuitGroup](#diode-v1-CircuitGroup) |  |  |
| termination_circuit_group_assignment | [CircuitGroupAssignment](#diode-v1-CircuitGroupAssignment) |  |  |
| termination_circuit_termination | [CircuitTermination](#diode-v1-CircuitTermination) |  |  |
| termination_circuit_type | [CircuitType](#diode-v1-CircuitType) |  |  |
| termination_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| termination_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| termination_cluster_type | [ClusterType](#diode-v1-ClusterType) |  |  |
| termination_console_port | [ConsolePort](#diode-v1-ConsolePort) |  |  |
| termination_console_server_port | [ConsoleServerPort](#diode-v1-ConsoleServerPort) |  |  |
| termination_contact | [Contact](#diode-v1-Contact) |  |  |
| termination_contact_assignment | [ContactAssignment](#diode-v1-ContactAssignment) |  |  |
| termination_contact_group | [ContactGroup](#diode-v1-ContactGroup) |  |  |
| termination_contact_role | [ContactRole](#diode-v1-ContactRole) |  |  |
| termination_device | [Device](#diode-v1-Device) |  |  |
| termination_device_bay | [DeviceBay](#diode-v1-DeviceBay) |  |  |
| termination_device_role | [DeviceRole](#diode-v1-DeviceRole) |  |  |
| termination_device_type | [DeviceType](#diode-v1-DeviceType) |  |  |
| termination_fhrp_group | [FHRPGroup](#diode-v1-FHRPGroup) |  |  |
| termination_fhrp_group_assignment | [FHRPGroupAssignment](#diode-v1-FHRPGroupAssignment) |  |  |
| termination_front_port | [FrontPort](#diode-v1-FrontPort) |  |  |
| termination_ike_policy | [IKEPolicy](#diode-v1-IKEPolicy) |  |  |
| termination_ike_proposal | [IKEProposal](#diode-v1-IKEProposal) |  |  |
| termination_ip_address | [IPAddress](#diode-v1-IPAddress) |  |  |
| termination_ip_range | [IPRange](#diode-v1-IPRange) |  |  |
| termination_ip_sec_policy | [IPSecPolicy](#diode-v1-IPSecPolicy) |  |  |
| termination_ip_sec_profile | [IPSecProfile](#diode-v1-IPSecProfile) |  |  |
| termination_ip_sec_proposal | [IPSecProposal](#diode-v1-IPSecProposal) |  |  |
| termination_interface | [Interface](#diode-v1-Interface) |  |  |
| termination_inventory_item | [InventoryItem](#diode-v1-InventoryItem) |  |  |
| termination_inventory_item_role | [InventoryItemRole](#diode-v1-InventoryItemRole) |  |  |
| termination_l2vpn | [L2VPN](#diode-v1-L2VPN) |  |  |
| termination_l2vpn_termination | [L2VPNTermination](#diode-v1-L2VPNTermination) |  |  |
| termination_location | [Location](#diode-v1-Location) |  |  |
| termination_mac_address | [MACAddress](#diode-v1-MACAddress) |  |  |
| termination_manufacturer | [Manufacturer](#diode-v1-Manufacturer) |  |  |
| termination_module | [Module](#diode-v1-Module) |  |  |
| termination_module_bay | [ModuleBay](#diode-v1-ModuleBay) |  |  |
| termination_module_type | [ModuleType](#diode-v1-ModuleType) |  |  |
| termination_platform | [Platform](#diode-v1-Platform) |  |  |
| termination_power_feed | [PowerFeed](#diode-v1-PowerFeed) |  |  |
| termination_power_outlet | [PowerOutlet](#diode-v1-PowerOutlet) |  |  |
| termination_power_panel | [PowerPanel](#diode-v1-PowerPanel) |  |  |
| termination_power_port | [PowerPort](#diode-v1-PowerPort) |  |  |
| termination_prefix | [Prefix](#diode-v1-Prefix) |  |  |
| termination_provider | [Provider](#diode-v1-Provider) |  |  |
| termination_provider_account | [ProviderAccount](#diode-v1-ProviderAccount) |  |  |
| termination_provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| termination_rir | [RIR](#diode-v1-RIR) |  |  |
| termination_rack | [Rack](#diode-v1-Rack) |  |  |
| termination_rack_reservation | [RackReservation](#diode-v1-RackReservation) |  |  |
| termination_rack_role | [RackRole](#diode-v1-RackRole) |  |  |
| termination_rack_type | [RackType](#diode-v1-RackType) |  |  |
| termination_rear_port | [RearPort](#diode-v1-RearPort) |  |  |
| termination_region | [Region](#diode-v1-Region) |  |  |
| termination_role | [Role](#diode-v1-Role) |  |  |
| termination_route_target | [RouteTarget](#diode-v1-RouteTarget) |  |  |
| termination_service | [Service](#diode-v1-Service) |  |  |
| termination_site | [Site](#diode-v1-Site) |  |  |
| termination_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| termination_tag | [Tag](#diode-v1-Tag) |  |  |
| termination_tenant | [Tenant](#diode-v1-Tenant) |  |  |
| termination_tenant_group | [TenantGroup](#diode-v1-TenantGroup) |  |  |
| termination_tunnel | [Tunnel](#diode-v1-Tunnel) |  |  |
| termination_tunnel_group | [TunnelGroup](#diode-v1-TunnelGroup) |  |  |
| termination_tunnel_termination | [TunnelTermination](#diode-v1-TunnelTermination) |  |  |
| termination_vlan | [VLAN](#diode-v1-VLAN) |  |  |
| termination_vlan_group | [VLANGroup](#diode-v1-VLANGroup) |  |  |
| termination_vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| termination_vlan_translation_rule | [VLANTranslationRule](#diode-v1-VLANTranslationRule) |  |  |
| termination_vm_interface | [VMInterface](#diode-v1-VMInterface) |  |  |
| termination_vrf | [VRF](#diode-v1-VRF) |  |  |
| termination_virtual_chassis | [VirtualChassis](#diode-v1-VirtualChassis) |  |  |
| termination_virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| termination_virtual_circuit_termination | [VirtualCircuitTermination](#diode-v1-VirtualCircuitTermination) |  |  |
| termination_virtual_circuit_type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| termination_virtual_device_context | [VirtualDeviceContext](#diode-v1-VirtualDeviceContext) |  |  |
| termination_virtual_disk | [VirtualDisk](#diode-v1-VirtualDisk) |  |  |
| termination_virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| termination_wireless_lan | [WirelessLAN](#diode-v1-WirelessLAN) |  |  |
| termination_wireless_lan_group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) |  |  |
| termination_wireless_link | [WirelessLink](#diode-v1-WirelessLink) |  |  |
| termination_custom_field | [CustomField](#diode-v1-CustomField) |  |  |
| termination_custom_field_choice_set | [CustomFieldChoiceSet](#diode-v1-CustomFieldChoiceSet) |  |  |
| termination_journal_entry | [JournalEntry](#diode-v1-JournalEntry) |  |  |
| termination_module_type_profile | [ModuleTypeProfile](#diode-v1-ModuleTypeProfile) |  |  |
| termination_custom_link | [CustomLink](#diode-v1-CustomLink) |  |  |
| outside_ip | [IPAddress](#diode-v1-IPAddress) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [TunnelTermination.CustomFieldsEntry](#diode-v1-TunnelTermination-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-TunnelTermination-CustomFieldsEntry"></a>

### TunnelTermination.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VLAN"></a>

### VLAN



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| site | [Site](#diode-v1-Site) | optional |  |
| group | [VLANGroup](#diode-v1-VLANGroup) | optional |  |
| vid | [int64](#int64) |  |  |
| name | [string](#string) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| status | [string](#string) | optional |  |
| role | [Role](#diode-v1-Role) | optional |  |
| description | [string](#string) | optional |  |
| qinq_role | [string](#string) | optional |  |
| qinq_svlan | [VLAN](#diode-v1-VLAN) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VLAN.CustomFieldsEntry](#diode-v1-VLAN-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VLAN-CustomFieldsEntry"></a>

### VLAN.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VLANGroup"></a>

### VLANGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| scope_cluster | [Cluster](#diode-v1-Cluster) |  |  |
| scope_cluster_group | [ClusterGroup](#diode-v1-ClusterGroup) |  |  |
| scope_location | [Location](#diode-v1-Location) |  |  |
| scope_rack | [Rack](#diode-v1-Rack) |  |  |
| scope_region | [Region](#diode-v1-Region) |  |  |
| scope_site | [Site](#diode-v1-Site) |  |  |
| scope_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| vid_ranges | [int64](#int64) | repeated |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VLANGroup.CustomFieldsEntry](#diode-v1-VLANGroup-CustomFieldsEntry) | repeated |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VLANGroup-CustomFieldsEntry"></a>

### VLANGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VLANTranslationPolicy"></a>

### VLANTranslationPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VLANTranslationRule"></a>

### VLANTranslationRule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) |  |  |
| local_vid | [int64](#int64) |  |  |
| remote_vid | [int64](#int64) |  |  |
| description | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VMInterface"></a>

### VMInterface



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| name | [string](#string) |  |  |
| enabled | [bool](#bool) | optional |  |
| parent | [VMInterface](#diode-v1-VMInterface) | optional |  |
| bridge | [VMInterface](#diode-v1-VMInterface) | optional |  |
| mtu | [int64](#int64) | optional |  |
| primary_mac_address | [MACAddress](#diode-v1-MACAddress) | optional |  |
| description | [string](#string) | optional |  |
| mode | [string](#string) | optional |  |
| untagged_vlan | [VLAN](#diode-v1-VLAN) | optional |  |
| qinq_svlan | [VLAN](#diode-v1-VLAN) | optional |  |
| vlan_translation_policy | [VLANTranslationPolicy](#diode-v1-VLANTranslationPolicy) | optional |  |
| vrf | [VRF](#diode-v1-VRF) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VMInterface.CustomFieldsEntry](#diode-v1-VMInterface-CustomFieldsEntry) | repeated |  |
| tagged_vlans | [VLAN](#diode-v1-VLAN) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VMInterface-CustomFieldsEntry"></a>

### VMInterface.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VRF"></a>

### VRF



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| rd | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| enforce_unique | [bool](#bool) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VRF.CustomFieldsEntry](#diode-v1-VRF-CustomFieldsEntry) | repeated |  |
| import_targets | [RouteTarget](#diode-v1-RouteTarget) | repeated |  |
| export_targets | [RouteTarget](#diode-v1-RouteTarget) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VRF-CustomFieldsEntry"></a>

### VRF.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualChassis"></a>

### VirtualChassis



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| domain | [string](#string) | optional |  |
| master | [Device](#diode-v1-Device) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualChassis.CustomFieldsEntry](#diode-v1-VirtualChassis-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualChassis-CustomFieldsEntry"></a>

### VirtualChassis.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualCircuit"></a>

### VirtualCircuit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cid | [string](#string) |  |  |
| provider_network | [ProviderNetwork](#diode-v1-ProviderNetwork) |  |  |
| provider_account | [ProviderAccount](#diode-v1-ProviderAccount) | optional |  |
| type | [VirtualCircuitType](#diode-v1-VirtualCircuitType) |  |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualCircuit.CustomFieldsEntry](#diode-v1-VirtualCircuit-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualCircuit-CustomFieldsEntry"></a>

### VirtualCircuit.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualCircuitTermination"></a>

### VirtualCircuitTermination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtual_circuit | [VirtualCircuit](#diode-v1-VirtualCircuit) |  |  |
| role | [string](#string) | optional |  |
| interface | [Interface](#diode-v1-Interface) |  |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualCircuitTermination.CustomFieldsEntry](#diode-v1-VirtualCircuitTermination-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualCircuitTermination-CustomFieldsEntry"></a>

### VirtualCircuitTermination.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualCircuitType"></a>

### VirtualCircuitType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| color | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualCircuitType.CustomFieldsEntry](#diode-v1-VirtualCircuitType-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualCircuitType-CustomFieldsEntry"></a>

### VirtualCircuitType.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualDeviceContext"></a>

### VirtualDeviceContext



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| device | [Device](#diode-v1-Device) |  |  |
| identifier | [int64](#int64) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| primary_ip4 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| primary_ip6 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| status | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualDeviceContext.CustomFieldsEntry](#diode-v1-VirtualDeviceContext-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualDeviceContext-CustomFieldsEntry"></a>

### VirtualDeviceContext.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualDisk"></a>

### VirtualDisk



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| size | [int64](#int64) |  |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualDisk.CustomFieldsEntry](#diode-v1-VirtualDisk-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualDisk-CustomFieldsEntry"></a>

### VirtualDisk.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-VirtualMachine"></a>

### VirtualMachine



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| status | [string](#string) | optional |  |
| site | [Site](#diode-v1-Site) | optional |  |
| cluster | [Cluster](#diode-v1-Cluster) | optional |  |
| device | [Device](#diode-v1-Device) | optional |  |
| serial | [string](#string) | optional |  |
| role | [DeviceRole](#diode-v1-DeviceRole) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| platform | [Platform](#diode-v1-Platform) | optional |  |
| primary_ip4 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| primary_ip6 | [IPAddress](#diode-v1-IPAddress) | optional |  |
| vcpus | [double](#double) | optional |  |
| memory | [int64](#int64) | optional |  |
| disk | [int64](#int64) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [VirtualMachine.CustomFieldsEntry](#diode-v1-VirtualMachine-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-VirtualMachine-CustomFieldsEntry"></a>

### VirtualMachine.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-WirelessLAN"></a>

### WirelessLAN



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ssid | [string](#string) |  |  |
| description | [string](#string) | optional |  |
| group | [WirelessLANGroup](#diode-v1-WirelessLANGroup) | optional |  |
| status | [string](#string) | optional |  |
| vlan | [VLAN](#diode-v1-VLAN) | optional |  |
| scope_location | [Location](#diode-v1-Location) |  |  |
| scope_region | [Region](#diode-v1-Region) |  |  |
| scope_site | [Site](#diode-v1-Site) |  |  |
| scope_site_group | [SiteGroup](#diode-v1-SiteGroup) |  |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| auth_type | [string](#string) | optional |  |
| auth_cipher | [string](#string) | optional |  |
| auth_psk | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [WirelessLAN.CustomFieldsEntry](#diode-v1-WirelessLAN-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-WirelessLAN-CustomFieldsEntry"></a>

### WirelessLAN.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-WirelessLANGroup"></a>

### WirelessLANGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| parent | [WirelessLANGroup](#diode-v1-WirelessLANGroup) | optional |  |
| description | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [WirelessLANGroup.CustomFieldsEntry](#diode-v1-WirelessLANGroup-CustomFieldsEntry) | repeated |  |
| comments | [string](#string) | optional |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-WirelessLANGroup-CustomFieldsEntry"></a>

### WirelessLANGroup.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |






<a name="diode-v1-WirelessLink"></a>

### WirelessLink



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| interface_a | [Interface](#diode-v1-Interface) |  |  |
| interface_b | [Interface](#diode-v1-Interface) |  |  |
| ssid | [string](#string) | optional |  |
| status | [string](#string) | optional |  |
| tenant | [Tenant](#diode-v1-Tenant) | optional |  |
| auth_type | [string](#string) | optional |  |
| auth_cipher | [string](#string) | optional |  |
| auth_psk | [string](#string) | optional |  |
| distance | [double](#double) | optional |  |
| distance_unit | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| comments | [string](#string) | optional |  |
| tags | [Tag](#diode-v1-Tag) | repeated |  |
| custom_fields | [WirelessLink.CustomFieldsEntry](#diode-v1-WirelessLink-CustomFieldsEntry) | repeated |  |
| metadata | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="diode-v1-WirelessLink-CustomFieldsEntry"></a>

### WirelessLink.CustomFieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CustomFieldValue](#diode-v1-CustomFieldValue) |  |  |





 

 


<a name="diode_v1_ingester-proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| netbox_supported | bool | .google.protobuf.FieldOptions | 50001 |  |

 


<a name="diode-v1-IngesterService"></a>

### IngesterService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Ingest | [IngestRequest](#diode-v1-IngestRequest) | [IngestResponse](#diode-v1-IngestResponse) | Ingests data into the system |

 



<a name="diode_v1_reconciler-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/reconciler.proto



<a name="diode-v1-Change"></a>

### Change
A change


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| object_type | [string](#string) |  |  |
| object_primary_value | [string](#string) |  |  |
| change_type | [string](#string) |  |  |
| before | [bytes](#bytes) |  |  |
| after | [bytes](#bytes) |  |  |






<a name="diode-v1-ChangeSet"></a>

### ChangeSet
A change set


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | A change set ID |
| data | [bytes](#bytes) |  | Binary data representing the change set |
| branch_id | [string](#string) | optional | branch ID against which the change set was generated |
| deviation_name | [string](#string) | optional | deviation name |






<a name="diode-v1-Deviation"></a>

### Deviation
A deviation


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| ingestion_ts | [int64](#int64) |  |  |
| last_update_ts | [int64](#int64) |  |  |
| name | [string](#string) |  |  |
| source | [string](#string) |  |  |
| state | [State](#diode-v1-State) |  |  |
| object_type | [string](#string) |  |  |
| branch_id | [string](#string) | optional |  |
| ingested_entity | [Entity](#diode-v1-Entity) |  |  |
| error | [DeviationError](#diode-v1-DeviationError) |  |  |
| changes | [Change](#diode-v1-Change) | repeated |  |
| source_ts | [int64](#int64) |  |  |






<a name="diode-v1-DeviationError"></a>

### DeviationError
Deviation error


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| code | [string](#string) |  |  |
| details | [bytes](#bytes) |  |  |






<a name="diode-v1-IngestionLog"></a>

### IngestionLog
An ingestion log


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| data_type | [string](#string) |  | **Deprecated.**  |
| state | [State](#diode-v1-State) |  |  |
| request_id | [string](#string) |  |  |
| ingestion_ts | [int64](#int64) |  |  |
| producer_app_name | [string](#string) |  |  |
| producer_app_version | [string](#string) |  |  |
| sdk_name | [string](#string) |  |  |
| sdk_version | [string](#string) |  |  |
| entity | [Entity](#diode-v1-Entity) |  |  |
| error | [DeviationError](#diode-v1-DeviationError) |  |  |
| change_set | [ChangeSet](#diode-v1-ChangeSet) |  |  |
| object_type | [string](#string) |  |  |
| source_ts | [int64](#int64) |  |  |






<a name="diode-v1-IngestionMetrics"></a>

### IngestionMetrics
Ingestion metrics


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| total | [int32](#int32) |  |  |
| queued | [int32](#int32) |  |  |
| reconciled | [int32](#int32) |  |  |
| failed | [int32](#int32) |  |  |
| no_changes | [int32](#int32) |  |  |






<a name="diode-v1-RetrieveDeviationByIDRequest"></a>

### RetrieveDeviationByIDRequest
The request to retrieve deviation by ID


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Deviation ID |






<a name="diode-v1-RetrieveDeviationByIDResponse"></a>

### RetrieveDeviationByIDResponse
The response from the retrieve deviation by ID request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deviation | [Deviation](#diode-v1-Deviation) |  | Deviation |






<a name="diode-v1-RetrieveDeviationsRequest"></a>

### RetrieveDeviationsRequest
The request to retrieve deviations


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) | optional | Number of deviations per page, default is 100 |
| page_token | [string](#string) |  | Token to fetch the next page of results |
| ingestion_ts_start | [int64](#int64) |  | Optional start of ingestion timestamp range |
| ingestion_ts_end | [int64](#int64) |  | Optional end of ingestion timestamp range |
| state | [State](#diode-v1-State) | repeated | Optional filter by states |
| object_type | [string](#string) | repeated | Optional filter by object types |
| branch_id | [string](#string) | repeated | Optional filter by branch IDs |






<a name="diode-v1-RetrieveDeviationsResponse"></a>

### RetrieveDeviationsResponse
The response from the retrieve deviations request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deviations | [Deviation](#diode-v1-Deviation) | repeated | List of deviations |
| next_page_token | [string](#string) |  | Token for the next page of results, if any |






<a name="diode-v1-RetrieveIngestionLogsRequest"></a>

### RetrieveIngestionLogsRequest
The request to retrieve ingestion logs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) | optional | Number of logs per page, default is 100 |
| state | [State](#diode-v1-State) | optional | Optional filter by state field |
| data_type | [string](#string) |  | **Deprecated.** Optional filter by data type field |
| request_id | [string](#string) |  | Optional filter by request ID |
| ingestion_ts_start | [int64](#int64) |  | Optional start of ingestion timestamp range |
| ingestion_ts_end | [int64](#int64) |  | Optional end of ingestion timestamp range |
| page_token | [string](#string) |  | Token to fetch the next page of results |
| only_metrics | [bool](#bool) |  | Flag to return only the ingestion metrics |
| object_type | [string](#string) |  | Optional filter by object type |






<a name="diode-v1-RetrieveIngestionLogsResponse"></a>

### RetrieveIngestionLogsResponse
The response from the retrieve ingestion logs request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| logs | [IngestionLog](#diode-v1-IngestionLog) | repeated | List of ingestion logs |
| metrics | [IngestionMetrics](#diode-v1-IngestionMetrics) |  | ingestion metrics |
| next_page_token | [string](#string) |  | Token for the next page of results, if any |





 


<a name="diode-v1-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNSPECIFIED | 0 |  |
| QUEUED | 1 |  |
| OPEN | 2 |  |
| APPLIED | 3 |  |
| FAILED | 4 |  |
| NO_CHANGES | 5 |  |
| IGNORED | 6 |  |
| ERRORED | 7 |  |


 

 


<a name="diode-v1-ReconcilerService"></a>

### ReconcilerService
Reconciler service API

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RetrieveIngestionLogs | [RetrieveIngestionLogsRequest](#diode-v1-RetrieveIngestionLogsRequest) | [RetrieveIngestionLogsResponse](#diode-v1-RetrieveIngestionLogsResponse) | Retrieves ingestion logs |
| RetrieveDeviations | [RetrieveDeviationsRequest](#diode-v1-RetrieveDeviationsRequest) | [RetrieveDeviationsResponse](#diode-v1-RetrieveDeviationsResponse) | Retrieve deviations |
| RetrieveDeviationByID | [RetrieveDeviationByIDRequest](#diode-v1-RetrieveDeviationByIDRequest) | [RetrieveDeviationByIDResponse](#diode-v1-RetrieveDeviationByIDResponse) | Retrieve deviation by ID |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

