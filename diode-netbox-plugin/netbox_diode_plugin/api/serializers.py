#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Serializer."""
import copy
import logging
from collections import OrderedDict

from dcim.api.serializers import (
    DeviceRoleSerializer,
    DeviceSerializer,
    DeviceTypeSerializer,
    InterfaceSerializer,
    ManufacturerSerializer,
    PlatformSerializer,
    SiteSerializer,
)
from django.core.exceptions import FieldDoesNotExist
from extras.models import ObjectChange
from ipam.api.serializers import IPAddressSerializer, PrefixSerializer
from rest_framework import serializers
from rest_framework.utils.serializer_helpers import ReturnDict
from utilities.api import get_serializer_for_model

logger = logging.getLogger("netbox.netbox_diode_plugin.api.serializers")


def dynamic_import(name):
    """Dynamically import a class from an absolute path string."""
    components = name.split(".")
    mod = __import__(components[0])
    for comp in components[1:]:
        mod = getattr(mod, comp)
    return mod


def get_diode_serializer(instance):
    """Get the Diode serializer based on instance model."""
    serializer = get_serializer_for_model(instance)

    serializer_name = f"netbox_diode_plugin.api.serializers.Diode{serializer.__name__}"

    try:
        serializer = dynamic_import(serializer_name)
    except AttributeError:
        logger.warning(f"Could not find serializer for {serializer_name}")
        pass

    return serializer


class ObjectStateSerializer(serializers.Serializer):
    """Object State Serializer."""

    object_type = serializers.SerializerMethodField(read_only=True)
    object_change_id = serializers.SerializerMethodField(read_only=True)
    object = serializers.SerializerMethodField(read_only=True)

    def get_object_type(self, instance):
        """
        Get the object type from context sent from view.

        Return a string with the format "app.model".
        """
        return self.context.get("object_type")

    def get_object_change_id(self, instance):
        """
        Get the object changed based on instance ID.

        Return the ID of last change.
        """
        object_changed = (
            ObjectChange.objects.filter(changed_object_id=instance.id)
            .order_by("-id")
            .values_list("id", flat=True)
        )
        return object_changed[0] if len(object_changed) > 0 else None

    def get_object(self, instance):
        """
        Get the serializer based on instance model.

        Get the data from the model according to its ID.
        Return the object according to serializer defined in the Netbox.
        """
        serializer = get_diode_serializer(instance)

        object_data = instance.__class__.objects.filter(id=instance.id)

        context = {"request": self.context.get("request")}

        data = serializer(object_data, context=context, many=True).data[0]

        return data


class ChangeSerialiazer(serializers.Serializer):
    """ChangeSet Serializer."""

    change_id = serializers.UUIDField(required=True)
    change_type = serializers.CharField(required=True)
    object_version = serializers.IntegerField(required=False, allow_null=True)
    object_type = serializers.CharField(required=True)
    object_id = serializers.IntegerField(required=False, allow_null=True)
    data = serializers.DictField(required=True)


class ApplyChangeSetRequestSerializer(serializers.Serializer):
    """ApplyChangeSet request Serializer."""

    change_set_id = serializers.UUIDField(required=True)
    change_set = serializers.ListField(
        child=ChangeSerialiazer(), required=True, allow_empty=False
    )


class DiodeIPAddressSerializer(IPAddressSerializer):
    """Diode IP Address Serializer."""

    class Meta:
        """Meta class."""

        model = IPAddressSerializer.Meta.model
        fields = IPAddressSerializer.Meta.fields

    def get_assigned_object(self, obj):
        """Get the assigned object based on the instance model."""
        if obj.assigned_object is None:
            return None

        serializer = get_diode_serializer(obj.assigned_object)

        context = {"request": self.context["request"]}
        assigned_object = serializer(obj.assigned_object, context=context).data

        if assigned_object.get("device"):
            device_serializer = get_diode_serializer(obj.assigned_object.device)
            device = device_serializer(obj.assigned_object.device, context=context).data
            assigned_object["device"] = device

        if serializer.__name__.endswith("InterfaceSerializer"):
            assigned_object = {"interface": assigned_object}

        return assigned_object


class DiodeSiteSerializer(SiteSerializer):
    """Diode Site Serializer."""

    status = serializers.CharField()

    class Meta:
        """Meta class."""

        model = SiteSerializer.Meta.model
        fields = SiteSerializer.Meta.fields


class DiodeDeviceRoleSerializer(DeviceRoleSerializer):
    """Diode Device Role Serializer."""

    class Meta:
        """Meta class."""

        model = DeviceRoleSerializer.Meta.model
        fields = DeviceRoleSerializer.Meta.fields


class DiodeManufacturerSerializer(ManufacturerSerializer):
    """Diode Manufacturer Serializer."""

    class Meta:
        """Meta class."""

        model = ManufacturerSerializer.Meta.model
        fields = ManufacturerSerializer.Meta.fields


class DiodePlatformSerializer(PlatformSerializer):
    """Diode Platform Serializer."""

    manufacturer = DiodeManufacturerSerializer(required=False, allow_null=True)

    class Meta:
        """Meta class."""

        model = PlatformSerializer.Meta.model
        fields = PlatformSerializer.Meta.fields


class DiodeDeviceTypeSerializer(DeviceTypeSerializer):
    """Diode Device Type Serializer."""

    default_platform = DiodePlatformSerializer(required=False, allow_null=True)
    manufacturer = DiodeManufacturerSerializer(required=False, allow_null=True)

    class Meta:
        """Meta class."""

        model = DeviceTypeSerializer.Meta.model
        fields = DeviceTypeSerializer.Meta.fields


class DiodeDeviceSerializer(DeviceSerializer):
    """Diode Device Serializer."""

    site = DiodeSiteSerializer()
    device_type = DiodeDeviceTypeSerializer()
    role = DiodeDeviceRoleSerializer()
    device_role = DiodeDeviceRoleSerializer(
        read_only=True, help_text="Deprecated in v3.6 in favor of `role`."
    )
    platform = DiodePlatformSerializer(required=False, allow_null=True)
    primary_ip = DiodeIPAddressSerializer(read_only=True)
    primary_ip4 = DiodeIPAddressSerializer(required=False, allow_null=True)
    primary_ip6 = DiodeIPAddressSerializer(required=False, allow_null=True)
    oob_ip = DiodeIPAddressSerializer(required=False, allow_null=True)
    status = serializers.CharField()

    class Meta:
        """Meta class."""

        model = DeviceSerializer.Meta.model
        fields = DeviceSerializer.Meta.fields


class DiodeNestedInterfaceSerializer(InterfaceSerializer):
    """Diode Nested Interface Serializer."""

    class Meta:
        """Meta class."""

        model = InterfaceSerializer.Meta.model
        fields = InterfaceSerializer.Meta.fields


class DiodeInterfaceSerializer(InterfaceSerializer):
    """Diode Interface Serializer."""

    device = DiodeDeviceSerializer()
    parent = DiodeNestedInterfaceSerializer()
    type = serializers.CharField()
    mode = serializers.CharField()

    class Meta:
        """Meta class."""

        model = InterfaceSerializer.Meta.model
        fields = InterfaceSerializer.Meta.fields


class DiodePrefixSerializer(PrefixSerializer):
    """Diode Prefix Serializer."""

    site = DiodeSiteSerializer()
    status = serializers.CharField()

    class Meta:
        """Meta class."""

        model = PrefixSerializer.Meta.model
        fields = PrefixSerializer.Meta.fields
