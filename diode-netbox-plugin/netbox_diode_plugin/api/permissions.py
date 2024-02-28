#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Permissions."""

from rest_framework.permissions import SAFE_METHODS, BasePermission


class IsDiodeViewer(BasePermission):
    """Custom permission to allow users that has permission "netbox_diode_plugin.view_objectstate" to view the object type."""

    def has_permission(self, request, view):
        """Check if the request is in SAFE_METHODS and user has netbox_diode_plugin.view_objectstate permission."""
        return request.method in SAFE_METHODS and request.user.has_perm(
            "netbox_diode_plugin.view_objectstate"
        )


class IsDiodePost(BasePermission):
    """Custom permission to allow users that has permission "netbox_diode_plugin.add_objectstate" and POST requests."""

    def has_permission(self, request, view):
        """Check if the request is in POST and user has netbox_diode_plugin.add_objectstate permission."""
        return request.method in ["POST"] and request.user.has_perm(
            "netbox_diode_plugin.add_objectstate"
        )
