package netbox

import (
	"errors"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/mitchellh/hashstructure/v2"
)

// IpamIPAddressDataWrapper represents the IPAM IP address data wrapper
type IpamIPAddressDataWrapper struct {
	IPAddress *IpamIPAddress

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*IpamIPAddressDataWrapper) comparableData() {}

// Data returns the IP address
func (dw *IpamIPAddressDataWrapper) Data() any {
	return dw.IPAddress
}

// IsValid returns true if the IPAddress is not nil
func (dw *IpamIPAddressDataWrapper) IsValid() bool {
	if dw.IPAddress != nil && !dw.hasParent && dw.IPAddress.Address == "" {
		dw.IPAddress = nil
	}

	if dw.IPAddress != nil {
		if err := dw.IPAddress.Validate(); err != nil {
			return false
		}
	}

	return dw.IPAddress != nil
}

// Normalise normalises the data
func (dw *IpamIPAddressDataWrapper) Normalise() {
	if dw.IsValid() && dw.IPAddress.Tags != nil && len(dw.IPAddress.Tags) == 0 {
		dw.IPAddress.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *IpamIPAddressDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.IPAddress != nil && dw.hasParent && dw.IPAddress.Address == "" {
		dw.IPAddress = nil
	}

	objects := make([]ComparableData, 0)

	if dw.IPAddress == nil && dw.intended {
		return objects, nil
	}

	if dw.IPAddress == nil && dw.hasParent {
		dw.placeholder = true
	}

	var assignedObject ComparableData
	if dw.IPAddress.AssignedObject != nil {
		switch dw.IPAddress.AssignedObject.(type) {
		case *IPAddressInterface:
			assignedObject = &DcimInterfaceDataWrapper{Interface: dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}
		}
	}

	if assignedObject != nil {
		do, err := assignedObject.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, do...)

		switch dw.IPAddress.AssignedObject.(type) {
		case *IPAddressInterface:
			dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface = assignedObject.Data().(*DcimInterface)
		}
	}

	if dw.IPAddress.Tags != nil {
		for _, t := range dw.IPAddress.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *IpamIPAddressDataWrapper) DataType() string {
	return IpamIPAddressObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *IpamIPAddressDataWrapper) ObjectStateQueryParams() map[string]string {
	params := map[string]string{
		"q": dw.IPAddress.Address,
	}
	switch dw.IPAddress.AssignedObject.(type) {
	case *IPAddressInterface:
		ao := dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface
		if ao != nil {
			params["interface__name"] = ao.Name
			if ao.Device != nil {
				params["interface__device__name"] = ao.Device.Name
				if ao.Device.Site != nil {
					params["interface__device__site__name"] = ao.Device.Site.Name
				}
			}
		}
	}
	return params
}

// ID returns the ID of the data
func (dw *IpamIPAddressDataWrapper) ID() int {
	return dw.IPAddress.ID
}

// HasChanged returns true if the data has changed
func (dw *IpamIPAddressDataWrapper) HasChanged() bool {
	return dw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (dw *IpamIPAddressDataWrapper) IsPlaceholder() bool {
	return dw.placeholder
}

func (dw *IpamIPAddressDataWrapper) hash() string {
	var interfaceName, deviceName, siteName string
	if dw.IPAddress.AssignedObject != nil {
		switch dw.IPAddress.AssignedObject.(type) {
		case *IPAddressInterface:
			ao := dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface
			if ao != nil {
				interfaceName = ao.Name
				if ao.Device != nil {
					deviceName = ao.Device.Name
					if ao.Device.Site != nil {
						siteName = ao.Device.Site.Name
					}
				}
			}
		}
	}
	return slug.Make(fmt.Sprintf("%s-%s-%s-%s", dw.IPAddress.Address, interfaceName, deviceName, siteName))
}

// Patch creates patches between the actual, intended and current data
func (dw *IpamIPAddressDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*IpamIPAddressDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	var actualAssignedObject ComparableData
	var intendedAssignedObject ComparableData

	if dw.IPAddress.AssignedObject != nil {
		switch dw.IPAddress.AssignedObject.(type) {
		case *IPAddressInterface:
			assignedObject := &DcimInterfaceDataWrapper{Interface: dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}
			actualAssignedObject = extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", assignedObject.Data()))
			intendedAssignedObject = extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", assignedObject.Data()))
		}
	}

	reconciliationRequired := true

	if intended != nil && dw.hash() == intended.hash() {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		dw.IPAddress.ID = intended.IPAddress.ID
		dw.IPAddress.Address = intended.IPAddress.Address

		if actualAssignedObject != nil {
			assignedObjectsToReconcile, aoErr := actualAssignedObject.Patch(intendedAssignedObject, intendedNestedObjects)
			if aoErr != nil {
				return nil, aoErr
			}

			switch dw.IPAddress.AssignedObject.(type) {
			case *IPAddressInterface:
				assignedInterface, err := copyData(actualAssignedObject.Data().(*DcimInterface))
				if err != nil {
					return nil, err
				}
				assignedInterface.Tags = nil

				if !actualAssignedObject.HasChanged() {
					assignedInterface = &DcimInterface{
						ID: actualAssignedObject.ID(),
						Device: &DcimDevice{
							ID: actualAssignedObject.Data().(*DcimInterface).Device.ID,
						},
					}

					intendedAssignedInterfaceID := intendedAssignedObject.ID()
					intendedAssignedInterfaceDeviceID := intendedAssignedObject.Data().(*DcimInterface).Device.ID
					if intended.IPAddress.AssignedObject != nil {
						intendedAssignedInterfaceID = intended.IPAddress.AssignedObject.(*IPAddressInterface).Interface.ID
						intendedAssignedInterfaceDeviceID = intended.IPAddress.AssignedObject.(*IPAddressInterface).Interface.Device.ID
					}

					intended.IPAddress.AssignedObject = &IPAddressInterface{
						Interface: &DcimInterface{
							ID: intendedAssignedInterfaceID,
							Device: &DcimDevice{
								ID: intendedAssignedInterfaceDeviceID,
							},
						},
					}
				}

				dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface = assignedInterface
			}

			dw.objectsToReconcile = append(dw.objectsToReconcile, assignedObjectsToReconcile...)
		}

		if dw.IPAddress.AssignedObject == nil {
			dw.IPAddress.AssignedObject = intended.IPAddress.AssignedObject
		}

		if dw.IPAddress.Status == nil {
			dw.IPAddress.Status = intended.IPAddress.Status
		}

		if dw.IPAddress.Role == nil {
			dw.IPAddress.Role = intended.IPAddress.Role
		}

		if dw.IPAddress.DNSName == nil {
			dw.IPAddress.DNSName = intended.IPAddress.DNSName
		}

		if dw.IPAddress.Description == nil {
			dw.IPAddress.Description = intended.IPAddress.Description
		}

		if dw.IPAddress.Comments == nil {
			dw.IPAddress.Comments = intended.IPAddress.Comments
		}

		tagsToMerge := mergeTags(dw.IPAddress.Tags, intended.IPAddress.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.IPAddress.Tags = tagsToMerge
		}

		for _, t := range dw.IPAddress.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		var objectsToReconcile []ComparableData
		if actualAssignedObject != nil {
			assignedObjectsToReconcile, aoErr := actualAssignedObject.Patch(intendedAssignedObject, intendedNestedObjects)
			if aoErr != nil {
				return nil, aoErr
			}

			switch dw.IPAddress.AssignedObject.(type) {
			case *IPAddressInterface:
				assignedInterface, err := copyData(actualAssignedObject.Data().(*DcimInterface))
				if err != nil {
					return nil, err
				}
				assignedInterface.Tags = nil

				if !actualAssignedObject.HasChanged() {
					assignedInterface = &DcimInterface{
						ID: actualAssignedObject.Data().(*DcimInterface).ID,
						Device: &DcimDevice{
							ID: actualAssignedObject.Data().(*DcimInterface).Device.ID,
						},
					}
				}

				dw.IPAddress.AssignedObject.(*IPAddressInterface).Interface = assignedInterface
			}

			objectsToReconcile = append(objectsToReconcile, assignedObjectsToReconcile...)
		}

		tagsToMerge := mergeTags(dw.IPAddress.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.IPAddress.Tags = tagsToMerge
		}

		for _, t := range dw.IPAddress.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		if objectsToReconcile != nil {
			dw.objectsToReconcile = append(dw.objectsToReconcile, objectsToReconcile...)
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the IP address
func (dw *IpamIPAddressDataWrapper) SetDefaults() {
	if dw.IPAddress.Status == nil {
		dw.IPAddress.Status = &DefaultIPAddressStatus
	}
}

// IpamPrefixDataWrapper represents the IPAM Prefix data wrapper
type IpamPrefixDataWrapper struct {
	Prefix *IpamPrefix

	placeholder        bool
	hasParent          bool
	intended           bool
	hasChanged         bool
	nestedObjects      []ComparableData
	objectsToReconcile []ComparableData
}

func (*IpamPrefixDataWrapper) comparableData() {}

// Data returns the Prefix
func (dw *IpamPrefixDataWrapper) Data() any {
	return dw.Prefix
}

// IsValid returns true if the IpamPrefix is not nil
func (dw *IpamPrefixDataWrapper) IsValid() bool {
	if dw.Prefix != nil && !dw.hasParent && dw.Prefix.Prefix == "" {
		dw.Prefix = nil
	}

	if dw.Prefix != nil {
		if err := dw.Prefix.Validate(); err != nil {
			return false
		}
	}

	return dw.Prefix != nil
}

// Normalise normalises the data
func (dw *IpamPrefixDataWrapper) Normalise() {
	if dw.IsValid() && dw.Prefix.Tags != nil && len(dw.Prefix.Tags) == 0 {
		dw.Prefix.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *IpamPrefixDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Prefix != nil && dw.hasParent && dw.Prefix.Prefix == "" {
		dw.Prefix = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Prefix == nil && dw.intended {
		return objects, nil
	}

	if dw.Prefix == nil && dw.hasParent {
		dw.placeholder = true
	}

	site := DcimSiteDataWrapper{Site: dw.Prefix.Site, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	dw.Prefix.Site = site.Site

	if dw.Prefix.Tags != nil {
		for _, t := range dw.Prefix.Tags {
			if t.Slug == "" {
				t.Slug = slug.Make(t.Name)
			}
			objects = append(objects, &TagDataWrapper{Tag: t, hasParent: true})
		}
	}

	dw.nestedObjects = objects

	objects = append(objects, dw)

	return objects, nil
}

// DataType returns the data type
func (dw *IpamPrefixDataWrapper) DataType() string {
	return IpamPrefixObjectType
}

// ObjectStateQueryParams returns the query parameters needed to retrieve its object state
func (dw *IpamPrefixDataWrapper) ObjectStateQueryParams() map[string]string {
	return map[string]string{
		"q": dw.Prefix.Prefix,
	}
}

// ID returns the ID of the data
func (dw *IpamPrefixDataWrapper) ID() int {
	return dw.Prefix.ID
}

// HasChanged returns true if the data has changed
func (dw *IpamPrefixDataWrapper) HasChanged() bool {
	return dw.hasChanged
}

// IsPlaceholder returns true if the data is a placeholder
func (dw *IpamPrefixDataWrapper) IsPlaceholder() bool {
	return dw.placeholder
}

// Patch creates patches between the actual, intended and current data
func (dw *IpamPrefixDataWrapper) Patch(cmp ComparableData, intendedNestedObjects map[string]ComparableData) ([]ComparableData, error) {
	intended, ok := cmp.(*IpamPrefixDataWrapper)
	if !ok && intended != nil {
		return nil, errors.New("invalid data type")
	}

	actualNestedObjectsMap := make(map[string]ComparableData)
	for _, obj := range dw.nestedObjects {
		actualNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
	}

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Prefix.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Prefix.Site))

	reconciliationRequired := true

	if intended != nil {
		currentNestedObjectsMap := make(map[string]ComparableData)
		currentNestedObjects, err := intended.NestedObjects()
		if err != nil {
			return nil, err
		}
		for _, obj := range currentNestedObjects {
			currentNestedObjectsMap[fmt.Sprintf("%p", obj.Data())] = obj
		}

		dw.Prefix.ID = intended.Prefix.ID
		dw.Prefix.Prefix = intended.Prefix.Prefix

		if actualSite.IsPlaceholder() && intended.Prefix.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Prefix.Site))
		}

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}

			intendedSiteID := intendedSite.ID()
			if intended.Prefix.Site != nil {
				intendedSiteID = intended.Prefix.Site.ID
			}

			intended.Prefix.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		dw.Prefix.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		if dw.Prefix.Status == nil {
			dw.Prefix.Status = intended.Prefix.Status
		}

		if dw.Prefix.Description == nil {
			dw.Prefix.Description = intended.Prefix.Description
		}

		if dw.Prefix.Comments == nil {
			dw.Prefix.Comments = intended.Prefix.Comments
		}

		tagsToMerge := mergeTags(dw.Prefix.Tags, intended.Prefix.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Prefix.Tags = tagsToMerge
		}

		for _, t := range dw.Prefix.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}

		actualHash, _ := hashstructure.Hash(dw.Data(), hashstructure.FormatV2, nil)
		intendedHash, _ := hashstructure.Hash(intended.Data(), hashstructure.FormatV2, nil)

		reconciliationRequired = actualHash != intendedHash
	} else {
		dw.SetDefaults()

		siteObjectsToReconcile, siteErr := actualSite.Patch(intendedSite, intendedNestedObjects)
		if siteErr != nil {
			return nil, siteErr
		}

		site, err := copyData(actualSite.Data().(*DcimSite))
		if err != nil {
			return nil, err
		}
		site.Tags = nil

		if !actualSite.HasChanged() {
			site = &DcimSite{
				ID: actualSite.ID(),
			}
		}
		dw.Prefix.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Prefix.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Prefix.Tags = tagsToMerge
		}

		for _, t := range dw.Prefix.Tags {
			if t.ID == 0 {
				dw.objectsToReconcile = append(dw.objectsToReconcile, &TagDataWrapper{Tag: t, hasParent: true})
			}
		}
	}

	if reconciliationRequired {
		dw.hasChanged = true
		dw.objectsToReconcile = append(dw.objectsToReconcile, dw)
	}

	return dw.objectsToReconcile, nil
}

// SetDefaults sets the default values for the IPAM Prefix
func (dw *IpamPrefixDataWrapper) SetDefaults() {
	if dw.Prefix.Status == nil {
		dw.Prefix.Status = &DefaultPrefixStatus
	}
}
