#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Permissions."""

from rest_framework.permissions import SAFE_METHODS, BasePermission


class IsDiodeViewer(BasePermission):
    """
    Custom permission to allow users that has permissions to view the object type.

    For example, if the request contains "object_type=dcim.site" and the user has this permission, he can see the object.
    """

    def has_permission(self, request, view):
        """Check if the request is in SAFE_METHODS = ('GET', 'HEAD', 'OPTIONS')."""
        if request.method in SAFE_METHODS:
            return True
        return False

    def has_object_permission(self, request, view, obj):
        """Check if the user has the permission to view the object type."""
        app_label, model_name = obj.split(".")
        if request.user.has_perm(f"{app_label}.view_{model_name}"):
            return True
        return False
