#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Utils."""

from django.conf import settings
from packaging import version

if version.parse(settings.VERSION).major >= 4:
    from core.models import ObjectType as NetBoxType
else:
    from django.contrib.contenttypes.models import ContentType as NetBoxType

supported_object_types = {
    "dcim": [
        "device",
        "devicerole",
        "devicetype",
        "interface",
        "manufacturer",
        "platform",
        "site",
    ],
    "ipam": ["ipaddress", "prefix"],
    "extras": ["tag"],
}


def get_supported_object_types(sender):
    """Get supported object types."""
    content_type = NetBoxType.objects.get_for_model(sender, for_concrete_model=False)
    app_label = content_type.app_label
    model_name = content_type.model

    return (
        f"{app_label}.{model_name}"
        if model_name in supported_object_types.get(app_label, [])
        else None
    )
