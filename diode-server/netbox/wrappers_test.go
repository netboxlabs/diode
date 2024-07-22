package netbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataWrapper(t *testing.T) {
	tests := []struct {
		name      string
		dataType  string
		expected  ComparableData
		expectErr bool
	}{
		{"DcimDeviceObjectType", DcimDeviceObjectType, &DcimDeviceDataWrapper{}, false},
		{"DcimDeviceRoleObjectType", DcimDeviceRoleObjectType, &DcimDeviceRoleDataWrapper{}, false},
		{"DcimDeviceTypeObjectType", DcimDeviceTypeObjectType, &DcimDeviceTypeDataWrapper{}, false},
		{"DcimInterfaceObjectType", DcimInterfaceObjectType, &DcimInterfaceDataWrapper{}, false},
		{"DcimManufacturerObjectType", DcimManufacturerObjectType, &DcimManufacturerDataWrapper{}, false},
		{"DcimPlatformObjectType", DcimPlatformObjectType, &DcimPlatformDataWrapper{}, false},
		{"DcimSiteObjectType", DcimSiteObjectType, &DcimSiteDataWrapper{}, false},
		{"ExtrasTagObjectType", ExtrasTagObjectType, &TagDataWrapper{}, false},
		{"IpamIPAddressObjectType", IpamIPAddressObjectType, &IpamIPAddressDataWrapper{}, false},
		{"IpamPrefixObjectType", IpamPrefixObjectType, &IpamPrefixDataWrapper{}, false},
		{"UnsupportedType", "unsupported", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper, err := NewDataWrapper(tt.dataType)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, wrapper)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.expected, wrapper)
			}
		})
	}
}

func TestCopyData(t *testing.T) {
	type TestData struct {
		Field1 string
		Field2 int
	}

	src := &TestData{"test", 123}
	dst, err := copyData(src)
	require.NoError(t, err)
	assert.Equal(t, src, dst)
}

func TestDedupObjectsToReconcile(t *testing.T) {
	obj1 := &Tag{Name: "tag1"}
	obj2 := &Tag{Name: "tag2"}
	obj3 := &Tag{Name: "tag1"} // duplicate of obj1

	objects := []ComparableData{
		&TagDataWrapper{Tag: obj1},
		&TagDataWrapper{Tag: obj2},
		&TagDataWrapper{Tag: obj3},
	}

	deduped, err := dedupObjectsToReconcile(objects)
	require.NoError(t, err)
	assert.Len(t, deduped, 2)
}

func TestDcimDeviceDataWrapper(t *testing.T) {
	device := NewDcimDevice()
	wrapper := &DcimDeviceDataWrapper{Device: device}

	t.Run("Data", func(t *testing.T) {
		assert.Equal(t, device, wrapper.Data())
	})

	t.Run("IsValid", func(t *testing.T) {
		assert.True(t, wrapper.IsValid())
		wrapper.Device = nil
		assert.False(t, wrapper.IsValid())
	})

	t.Run("Normalise", func(t *testing.T) {
		wrapper.Device = device
		wrapper.Normalise()
		assert.True(t, wrapper.intended)
	})

	t.Run("DataType", func(t *testing.T) {
		assert.Equal(t, DcimDeviceObjectType, wrapper.DataType())
	})

	t.Run("ObjectStateQueryParams", func(t *testing.T) {
		wrapper.Device.Name = "test"
		params := wrapper.ObjectStateQueryParams()
		assert.Equal(t, "test", params["q"])
	})

	t.Run("ID", func(t *testing.T) {
		wrapper.Device.ID = 1
		assert.Equal(t, 1, wrapper.ID())
	})

	t.Run("HasChanged", func(t *testing.T) {
		assert.False(t, wrapper.HasChanged())
		wrapper.hasChanged = true
		assert.True(t, wrapper.HasChanged())
	})

	t.Run("IsPlaceholder", func(t *testing.T) {
		assert.False(t, wrapper.IsPlaceholder())
		wrapper.placeholder = true
		assert.True(t, wrapper.IsPlaceholder())
	})

	t.Run("SetDefaults", func(t *testing.T) {
		wrapper.SetDefaults()
		assert.NotNil(t, wrapper.Device.Status)
		assert.Equal(t, DcimDeviceStatusActive, *wrapper.Device.Status)
		wrapper.Device.Status = nil
		wrapper.SetDefaults()
		assert.NotNil(t, wrapper.Device.Status)
		assert.Equal(t, DcimDeviceStatusActive, *wrapper.Device.Status)
	})
}

func TestExtractFromObjectsMap(t *testing.T) {
	obj1 := &Tag{Name: "tag1"}
	obj2 := &Tag{Name: "tag2"}

	objectsMap := map[string]ComparableData{
		"key1": &TagDataWrapper{Tag: obj1},
		"key2": &TagDataWrapper{Tag: obj2},
	}

	assert.Equal(t, obj1, extractFromObjectsMap(objectsMap, "key1").Data())
	assert.Equal(t, obj2, extractFromObjectsMap(objectsMap, "key2").Data())
	assert.Nil(t, extractFromObjectsMap(objectsMap, "key3"))
}

func TestDcimDeviceDataWrapperNestedObjects(t *testing.T) {
	// Create nested objects
	manufacturer := &DcimManufacturer{Name: "manufacturer1"}
	site := &DcimSite{Name: "site1"}
	deviceType := &DcimDeviceType{Model: "model1", Manufacturer: manufacturer}
	deviceRole := &DcimDeviceRole{Name: "role1"}
	tag := &Tag{Name: "tag1"}
	platform := &DcimPlatform{Name: "platform1", Manufacturer: manufacturer}

	device := &DcimDevice{
		Name:       "device1",
		Site:       site,
		DeviceType: deviceType,
		Role:       deviceRole,
		Tags:       []*Tag{tag},
		Platform:   platform,
	}

	wrapper := &DcimDeviceDataWrapper{Device: device}

	// Call NestedObjects
	nestedObjects, err := wrapper.NestedObjects()
	require.NoError(t, err)

	// Check the length of nestedObjects
	assert.Len(t, nestedObjects, 8) // The device itself, site, deviceType, manufacturer, deviceRole, tag, and platform

	// Check the types and values of nested objects
	for _, obj := range nestedObjects {
		switch v := obj.(type) {
		case *DcimDeviceDataWrapper:
			assert.Equal(t, device, v.Device)
		case *DcimSiteDataWrapper:
			assert.Equal(t, site, v.Site)
		case *DcimDeviceTypeDataWrapper:
			assert.Equal(t, deviceType, v.DeviceType)
		case *DcimManufacturerDataWrapper:
			assert.Equal(t, manufacturer, v.Manufacturer)
		case *DcimDeviceRoleDataWrapper:
			assert.Equal(t, deviceRole, v.DeviceRole)
		case *TagDataWrapper:
			assert.Equal(t, tag, v.Tag)
		case *DcimPlatformDataWrapper:
			assert.Equal(t, platform, v.Platform)
		default:
			t.Fatalf("unexpected type: %T", v)
		}
	}
}
