# !/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Models."""

from netbox.models import NetBoxModel


class ObjectState(NetBoxModel):
    """
    Dummy model used to generate permissions for Diode NetBox Plugin. Does not exist in the database.
    """

    class Meta:
        managed = False

        default_permissions = ()

        permissions = (
            ("view_objectstate", "Can view ObjectState"),
        )
