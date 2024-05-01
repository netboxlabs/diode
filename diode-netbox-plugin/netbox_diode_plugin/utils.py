#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Utils."""

from django.contrib.contenttypes.models import ContentType

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
    content_type = ContentType.objects.get_for_model(sender, for_concrete_model=False)
    app_label = content_type.app_label
    model_name = content_type.model

    return (
        f"{app_label}.{model_name}"
        if model_name in supported_object_types.get(app_label, [])
        else None
    )
