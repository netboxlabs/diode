package netbox

import (
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

const (
	// ExtrasTagObjectType represents the tag object type
	ExtrasTagObjectType = "extras.tag"
)

// ComparableData is an interface for NetBox comparable data
type ComparableData interface {
	comparableData()

	// FromProtoEntity sets the data from a proto entity
	FromProtoEntity(protoData *diodepb.Entity) error

	// Data returns the data
	Data() any

	// IsValid checks if the data is not nil
	IsValid() bool

	// Normalise normalises the data
	Normalise()

	// NestedObjects returns all nested objects
	NestedObjects() ([]ComparableData, error)

	// DataType returns the data type
	DataType() string

	// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
	ObjectStateQueryParams() map[string]string

	// ID returns the ID of the data
	ID() int

	// IsPlaceholder returns true if the data is a placeholder
	IsPlaceholder() bool

	// SetDefaults sets the default values for the data
	SetDefaults()

	// Patch creates patches between the actual, intended and current data
	Patch(ComparableData, map[string]ComparableData) ([]ComparableData, error)

	// HasChanged returns true if the data has changed
	HasChanged() bool
}

// BaseDataWrapper is the base struct for all data wrappers
type BaseDataWrapper struct {
	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

// IsPlaceholder returns true if the data is a placeholder
func (bw *BaseDataWrapper) IsPlaceholder() bool {
	return bw.placeholder
}

// HasChanged returns true if the data has changed
func (bw *BaseDataWrapper) HasChanged() bool {
	return bw.hasChanged
}

func copyData[T any](srcData *T) (*T, error) {
	var dstData T
	if err := copier.Copy(&dstData, srcData); err != nil {
		return nil, err
	}
	return &dstData, nil
}

// Tag represents a tag
type Tag struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Slug  string `json:"slug,omitempty"`
	Color string `json:"color,omitempty"`
}

// TagDataWrapper represents a tag data wrapper
type TagDataWrapper struct {
	Tag *Tag

	placeholder bool
	hasParent   bool
}

func (*TagDataWrapper) comparableData() {}

// FromProtoEntity sets the data from a proto entity
func (dw *TagDataWrapper) FromProtoEntity(*diodepb.Entity) error {
	return nil
}

// Data returns the Tag
func (dw *TagDataWrapper) Data() any {
	return dw.Tag
}

// IsValid returns true if the Tag is not nil
func (dw *TagDataWrapper) IsValid() bool {
	return dw.Tag != nil
}

// Normalise normalises the data
func (dw *TagDataWrapper) Normalise() {}

// NestedObjects returns all nested objects
func (dw *TagDataWrapper) NestedObjects() ([]ComparableData, error) {
	return nil, nil
}

// DataType returns the data type
func (dw *TagDataWrapper) DataType() string {
	return ExtrasTagObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *TagDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.Tag.Name,
	}
}

// ID returns the ID of the data
func (dw *TagDataWrapper) ID() int {
	return dw.Tag.ID
}

// HasChanged returns true if the data has changed
func (dw *TagDataWrapper) HasChanged() bool {
	return false
}

// IsPlaceholder returns true if the data is a placeholder
func (dw *TagDataWrapper) IsPlaceholder() bool {
	return dw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (dw *TagDataWrapper) Patch(cmp ComparableData, _ map[string]ComparableData) ([]ComparableData, error) {
	d2, ok := cmp.(*TagDataWrapper)
	if !ok && d2 != nil {
		return nil, errors.New("invalid data type")
	}

	return nil, nil
}

// SetDefaults sets the default values for the platform
func (dw *TagDataWrapper) SetDefaults() {}

// FromProtoTags converts a slice of diode tags to a slice of NetBox tags
func FromProtoTags(tagsPb []*diodepb.Tag) []*Tag {
	if tagsPb == nil {
		return nil
	}

	var tags []*Tag
	for _, tagPb := range tagsPb {
		tags = append(tags, &Tag{
			Name:  tagPb.Name,
			Slug:  tagPb.Slug,
			Color: tagPb.Color,
		})
	}

	return tags
}

// NewDataWrapper creates a new data wrapper for the given data type
func NewDataWrapper(dataType string) (ComparableData, error) {
	switch dataType {
	case DcimDeviceObjectType:
		return &DcimDeviceDataWrapper{}, nil
	case DcimDeviceRoleObjectType:
		return &DcimDeviceRoleDataWrapper{}, nil
	case DcimDeviceTypeObjectType:
		return &DcimDeviceTypeDataWrapper{}, nil
	case DcimInterfaceObjectType:
		return &DcimInterfaceDataWrapper{}, nil
	case DcimManufacturerObjectType:
		return &DcimManufacturerDataWrapper{}, nil
	case DcimPlatformObjectType:
		return &DcimPlatformDataWrapper{}, nil
	case DcimSiteObjectType:
		return &DcimSiteDataWrapper{}, nil
	case ExtrasTagObjectType:
		return &TagDataWrapper{}, nil
	case IpamIPAddressObjectType:
		return &IpamIPAddressDataWrapper{}, nil
	case IpamPrefixObjectType:
		return &IpamPrefixDataWrapper{}, nil
	case VirtualizationClusterGroupObjectType:
		return &VirtualizationClusterGroupDataWrapper{}, nil
	case VirtualizationClusterTypeObjectType:
		return &VirtualizationClusterTypeDataWrapper{}, nil
	case VirtualizationClusterObjectType:
		return &VirtualizationClusterDataWrapper{}, nil
	case VirtualizationVirtualMachineObjectType:
		return &VirtualizationVirtualMachineDataWrapper{}, nil
	case VirtualizationVMInterfaceObjectType:
		return &VirtualizationVMInterfaceDataWrapper{}, nil
	case VirtualizationVirtualDiskObjectType:
		return &VirtualizationVirtualDiskDataWrapper{}, nil
	default:
		return nil, fmt.Errorf("unsupported data type %s", dataType)
	}
}

func extractFromObjectsMap(objectsMap map[string]ComparableData, key string) ComparableData {
	if obj, ok := objectsMap[key]; ok {
		return obj
	}
	return nil
}

func dedupObjectsToReconcile(objects []ComparableData) ([]ComparableData, error) {
	hashes := make(map[uint64]struct{})
	dedupedObjectsToReconcile := make([]ComparableData, 0)
	for _, o := range objects {
		hash, err := hashstructure.Hash(o.Data(), hashstructure.FormatV2, nil)
		if err != nil {
			return nil, err
		}
		if _, ok := hashes[hash]; ok {
			continue
		}
		hashes[hash] = struct{}{}
		dedupedObjectsToReconcile = append(dedupedObjectsToReconcile, o)
	}

	return dedupedObjectsToReconcile, nil
}

// int32PtrToIntPtr converts int32 pointer to int pointer
func int32PtrToIntPtr(v *int32) *int {
	var i *int
	if v != nil {
		i = new(int)
		*i = int(*v)
	}
	return i
}
