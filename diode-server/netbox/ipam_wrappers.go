package netbox

import (
	"errors"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/mitchellh/hashstructure/v2"
)

// IpamIPAddressDataWrapper represents the IPAM IP address data wrapper
type IpamIPAddressDataWrapper struct {
	BaseDataWrapper[IpamIPAddress]
}

// IsValid returns true if the IPAddress is not nil
func (dw *IpamIPAddressDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Address == "" {
		dw.Field = nil
	}

	if dw.Field != nil {
		if err := dw.Field.Validate(); err != nil {
			return false
		}
	}

	return dw.Field != nil
}

// Normalise normalises the data
func (dw *IpamIPAddressDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *IpamIPAddressDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Address == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.placeholder = true
	}

	var assignedObject ComparableData
	if dw.Field.AssignedObject != nil {
		switch dw.Field.AssignedObject.(type) {
		case *IPAddressInterface:
			assignedObject = &DcimInterfaceDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimInterface]{Field: dw.Field.AssignedObject.(*IPAddressInterface).Interface, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}
		}
	}

	if assignedObject != nil {
		do, err := assignedObject.NestedObjects()
		if err != nil {
			return nil, err
		}

		objects = append(objects, do...)

		switch dw.Field.AssignedObject.(type) {
		case *IPAddressInterface:
			dw.Field.AssignedObject.(*IPAddressInterface).Interface = assignedObject.Data().(*DcimInterface)
		}
	}

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Address,
	}
	switch dw.Field.AssignedObject.(type) {
	case *IPAddressInterface:
		ao := dw.Field.AssignedObject.(*IPAddressInterface).Interface
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
	return dw.Field.ID
}

func (dw *IpamIPAddressDataWrapper) hash() string {
	var interfaceName, deviceName, siteName string
	if dw.Field.AssignedObject != nil {
		switch dw.Field.AssignedObject.(type) {
		case *IPAddressInterface:
			ao := dw.Field.AssignedObject.(*IPAddressInterface).Interface
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
	return slug.Make(fmt.Sprintf("%s-%s-%s-%s", dw.Field.Address, interfaceName, deviceName, siteName))
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

	if dw.Field.AssignedObject != nil {
		switch dw.Field.AssignedObject.(type) {
		case *IPAddressInterface:
			assignedObject := &DcimInterfaceDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimInterface]{Field: dw.Field.AssignedObject.(*IPAddressInterface).Interface, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}
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

		dw.Field.ID = intended.Field.ID
		dw.Field.Address = intended.Field.Address

		if actualAssignedObject != nil {
			assignedObjectsToReconcile, aoErr := actualAssignedObject.Patch(intendedAssignedObject, intendedNestedObjects)
			if aoErr != nil {
				return nil, aoErr
			}

			switch dw.Field.AssignedObject.(type) {
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
					if intended.Field.AssignedObject != nil {
						intendedAssignedInterfaceID = intended.Field.AssignedObject.(*IPAddressInterface).Interface.ID
						intendedAssignedInterfaceDeviceID = intended.Field.AssignedObject.(*IPAddressInterface).Interface.Device.ID
					}

					intended.Field.AssignedObject = &IPAddressInterface{
						Interface: &DcimInterface{
							ID: intendedAssignedInterfaceID,
							Device: &DcimDevice{
								ID: intendedAssignedInterfaceDeviceID,
							},
						},
					}
				}

				dw.Field.AssignedObject.(*IPAddressInterface).Interface = assignedInterface
			}

			dw.objectsToReconcile = append(dw.objectsToReconcile, assignedObjectsToReconcile...)
		}

		if dw.Field.AssignedObject == nil {
			dw.Field.AssignedObject = intended.Field.AssignedObject
		}

		if dw.Field.Status == nil {
			dw.Field.Status = intended.Field.Status
		}

		if dw.Field.Role == nil {
			dw.Field.Role = intended.Field.Role
		}

		if dw.Field.DNSName == nil {
			dw.Field.DNSName = intended.Field.DNSName
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.Comments == nil {
			dw.Field.Comments = intended.Field.Comments
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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

			switch dw.Field.AssignedObject.(type) {
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

				dw.Field.AssignedObject.(*IPAddressInterface).Interface = assignedInterface
			}

			objectsToReconcile = append(objectsToReconcile, assignedObjectsToReconcile...)
		}

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Status == nil {
		dw.Field.Status = &DefaultIPAddressStatus
	}
}

// IpamPrefixDataWrapper represents the IPAM Prefix data wrapper
type IpamPrefixDataWrapper struct {
	BaseDataWrapper[IpamPrefix]
}

// IsValid returns true if the IpamPrefix is not nil
func (dw *IpamPrefixDataWrapper) IsValid() bool {
	if dw.Field != nil && !dw.hasParent && dw.Field.Prefix == "" {
		dw.Field = nil
	}

	if dw.Field != nil {
		if err := dw.Field.Validate(); err != nil {
			return false
		}
	}

	return dw.Field != nil
}

// Normalise normalises the data
func (dw *IpamPrefixDataWrapper) Normalise() {
	if dw.IsValid() && dw.Field.Tags != nil && len(dw.Field.Tags) == 0 {
		dw.Field.Tags = nil
	}
	dw.intended = true
}

// NestedObjects returns all nested objects
func (dw *IpamPrefixDataWrapper) NestedObjects() ([]ComparableData, error) {
	if len(dw.nestedObjects) > 0 {
		return dw.nestedObjects, nil
	}

	if dw.Field != nil && dw.hasParent && dw.Field.Prefix == "" {
		dw.Field = nil
	}

	objects := make([]ComparableData, 0)

	if dw.Field == nil && dw.intended {
		return objects, nil
	}

	if dw.Field == nil && dw.hasParent {
		dw.placeholder = true
	}

	site := DcimSiteDataWrapper{BaseDataWrapper: BaseDataWrapper[DcimSite]{Field: dw.Field.Site, placeholder: dw.placeholder, hasParent: true, intended: dw.intended}}

	so, err := site.NestedObjects()
	if err != nil {
		return nil, err
	}

	objects = append(objects, so...)

	dw.Field.Site = site.Field

	if dw.Field.Tags != nil {
		for _, t := range dw.Field.Tags {
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
		"q": dw.Field.Prefix,
	}
}

// ID returns the ID of the data
func (dw *IpamPrefixDataWrapper) ID() int {
	return dw.Field.ID
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

	actualSite := extractFromObjectsMap(actualNestedObjectsMap, fmt.Sprintf("%p", dw.Field.Site))
	intendedSite := extractFromObjectsMap(intendedNestedObjects, fmt.Sprintf("%p", dw.Field.Site))

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

		dw.Field.ID = intended.Field.ID
		dw.Field.Prefix = intended.Field.Prefix

		if actualSite.IsPlaceholder() && intended.Field.Site != nil {
			intendedSite = extractFromObjectsMap(currentNestedObjectsMap, fmt.Sprintf("%p", intended.Field.Site))
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
			if intended.Field.Site != nil {
				intendedSiteID = intended.Field.Site.ID
			}

			intended.Field.Site = &DcimSite{
				ID: intendedSiteID,
			}
		}

		dw.Field.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		if dw.Field.Status == nil {
			dw.Field.Status = intended.Field.Status
		}

		if dw.Field.Description == nil {
			dw.Field.Description = intended.Field.Description
		}

		if dw.Field.Comments == nil {
			dw.Field.Comments = intended.Field.Comments
		}

		tagsToMerge := mergeTags(dw.Field.Tags, intended.Field.Tags, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
		dw.Field.Site = site

		dw.objectsToReconcile = append(dw.objectsToReconcile, siteObjectsToReconcile...)

		tagsToMerge := mergeTags(dw.Field.Tags, nil, intendedNestedObjects)

		if len(tagsToMerge) > 0 {
			dw.Field.Tags = tagsToMerge
		}

		for _, t := range dw.Field.Tags {
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
	if dw.Field.Status == nil {
		dw.Field.Status = &DefaultPrefixStatus
	}
}
