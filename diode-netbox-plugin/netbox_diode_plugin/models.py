# !/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Models."""

from django.db import models


class Diode(models.Model):
    """Dummy model used to generate permissions for Diode NetBox Plugin. Does not exist in the database."""

    class Meta:
        """Meta class."""

        managed = False

        default_permissions = ()

        permissions = (
            ("view_diode", "Can view Diode"),
            ("add_diode", "Can apply change sets from Diode"),
        )
