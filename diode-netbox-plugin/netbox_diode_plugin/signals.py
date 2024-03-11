#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Signals."""

import logging

from django.contrib.auth import get_user_model
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.forms import model_to_dict
from extras.models import ObjectChange
from users.models import Token

from netbox_diode_plugin.diode_reconciler_sdk.client import DiodeReconcilerClient
from netbox_diode_plugin.utils import get_supported_object_types

logger = logging.getLogger("netbox.netbox_diode_plugin")

User = get_user_model()


def get_netbox_to_diode_token():
    """Get token for NETBOX_TO_DIODE."""
    user = get_user_model().objects.get(username="NETBOX_TO_DIODE")
    return Token.objects.get(user=user)


@receiver(post_save)
def handle_notify_diode(instance, created, sender, update_fields, **kwargs):
    """Handle notify reconciliation."""
    logger.debug("Handling notify reconciliation.")

    supported_object_type = get_supported_object_types(sender)  # noqa

    if supported_object_type:

        model_name = supported_object_type.split(".")[1]  # noqa

        object_changed = (
            ObjectChange.objects.filter(changed_object_id=instance.id)
            .order_by("id")
            .last()
        )
        object_id = instance.id  # noqa
        object_type = supported_object_type  # noqa
        object_changed_id = object_changed if object_changed else None  # noqa
        object = {model_name: model_to_dict(instance)}  # noqa

        try:
            sdk = DiodeReconcilerClient(  # noqa
                "diode-reconciler:8081", get_netbox_to_diode_token()
            )

            # Comment out because the DiodeReconcilerClient need some adjustments.

            # sdk.add_object_state(
            #     object_id=object_id,
            #     object_type=object_type,
            #     object_change_id=object_changed_id,
            #     object=object,
            # )
        except Exception as e:
            logger.error(e)

            return False

        return True

    return None
