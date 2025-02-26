package netbox

import (
	"fmt"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// TODO(ltucker): generate these

func GetObjectType(entity *diodepb.Entity) (string, error) {
	switch entity.GetEntity().(type) {
	case *diodepb.Entity_Device:
		return DcimDeviceObjectType, nil
	case *diodepb.Entity_DeviceRole:
		return DcimDeviceRoleObjectType, nil
	case *diodepb.Entity_DeviceType:
		return DcimDeviceTypeObjectType, nil
	case *diodepb.Entity_Interface:
		return DcimInterfaceObjectType, nil
	case *diodepb.Entity_Manufacturer:
		return DcimManufacturerObjectType, nil
	case *diodepb.Entity_Platform:
		return DcimPlatformObjectType, nil
	case *diodepb.Entity_Site:
		return DcimSiteObjectType, nil
	case *diodepb.Entity_IpAddress:
		return IpamIPAddressObjectType, nil
	case *diodepb.Entity_Prefix:
		return IpamPrefixObjectType, nil
	case *diodepb.Entity_ClusterGroup:
		return VirtualizationClusterGroupObjectType, nil
	case *diodepb.Entity_ClusterType:
		return VirtualizationClusterTypeObjectType, nil
	case *diodepb.Entity_Cluster:
		return VirtualizationClusterObjectType, nil
	case *diodepb.Entity_VirtualMachine:
		return VirtualizationVirtualMachineObjectType, nil
	case *diodepb.Entity_Vminterface:
		return VirtualizationVMInterfaceObjectType, nil
	case *diodepb.Entity_VirtualDisk:
		return VirtualizationVirtualDiskObjectType, nil
	default:
		return "", fmt.Errorf("unknown object type")
	}
}

func GetObjectTypeName(objectType string) (string, error) {
	switch objectType {
	case DcimDeviceObjectType:
		return DcimDeviceObjectTypeName, nil
	case DcimDeviceRoleObjectType:
		return DcimDeviceRoleObjectTypeName, nil
	case DcimDeviceTypeObjectType:
		return DcimDeviceTypeObjectTypeName, nil
	case DcimInterfaceObjectType:
		return DcimInterfaceObjectTypeName, nil
	case DcimManufacturerObjectType:
		return DcimManufacturerObjectTypeName, nil
	case DcimPlatformObjectType:
		return DcimPlatformObjectTypeName, nil
	case DcimSiteObjectType:
		return DcimSiteObjectTypeName, nil
	case IpamIPAddressObjectType:
		return IpamIPAddressObjectTypeName, nil
	case IpamPrefixObjectType:
		return IpamPrefixObjectTypeName, nil
	case VirtualizationClusterGroupObjectType:
		return VirtualizationClusterGroupObjectTypeName, nil
	case VirtualizationClusterTypeObjectType:
		return VirtualizationClusterTypeObjectTypeName, nil
	case VirtualizationClusterObjectType:
		return VirtualizationClusterObjectTypeName, nil
	case VirtualizationVirtualMachineObjectType:
		return VirtualizationVirtualMachineObjectTypeName, nil
	case VirtualizationVMInterfaceObjectType:
		return VirtualizationVMInterfaceObjectTypeName, nil
	case VirtualizationVirtualDiskObjectType:
		return VirtualizationVirtualDiskObjectTypeName, nil
	default:
		return "", fmt.Errorf("unknown object type")
	}
}

func GetPrimaryValue(entity *diodepb.Entity) (string, error) {
	switch e := entity.GetEntity().(type) {
	case *diodepb.Entity_Device:
		return e.Device.Name, nil
	case *diodepb.Entity_DeviceRole:
		return e.DeviceRole.Name, nil
	case *diodepb.Entity_DeviceType:
		return e.DeviceType.Model, nil
	case *diodepb.Entity_Interface:
		return e.Interface.Name, nil
	case *diodepb.Entity_Manufacturer:
		return e.Manufacturer.Name, nil
	case *diodepb.Entity_Platform:
		return e.Platform.Name, nil
	case *diodepb.Entity_Site:
		return e.Site.Name, nil
	case *diodepb.Entity_IpAddress:
		return e.IpAddress.Address, nil
	case *diodepb.Entity_Prefix:
		return e.Prefix.Prefix, nil
	case *diodepb.Entity_ClusterGroup:
		return e.ClusterGroup.Name, nil
	case *diodepb.Entity_ClusterType:
		return e.ClusterType.Name, nil
	case *diodepb.Entity_Cluster:
		return e.Cluster.Name, nil
	case *diodepb.Entity_VirtualMachine:
		return e.VirtualMachine.Name, nil
	case *diodepb.Entity_Vminterface:
		return e.Vminterface.Name, nil
	case *diodepb.Entity_VirtualDisk:
		return e.VirtualDisk.Name, nil
	default:
		return "", fmt.Errorf("unknown object type")
	}
}
