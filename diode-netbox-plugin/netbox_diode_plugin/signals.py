#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Signals."""


import logging

from django.contrib.contenttypes.models import ContentType

from django.db.models.signals import post_save
from django.dispatch import receiver
from django.forms import model_to_dict

from extras.models import ObjectChange
from dcim.models import Device
from netbox_diode_plugin.diode_reconciler_sdk.client import DiodeReconcilerClient

logger = logging.getLogger("netbox.netbox_diode_plugin")


@receiver(post_save)
def handle_notifify_reconciliation(instance, created, sender, update_fields, **kwargs):
    """Handle notify reconciliation."""
    logger.debug("Handling notify reconciliation.")

    # print(created)
    # print(instance.id)
    # print(sender)
    # print(update_fields)

    content_type = ContentType.objects.get_for_model(sender, for_concrete_model=False)
    app_label = content_type.app_label
    model_name = content_type.model

    # print(f"{app_label=}")
    # print(f"{model_name=}")

    if app_label == "dcim":
        object_changed = (
            ObjectChange.objects.filter(changed_object_id=instance.id)
            .order_by("id")
            .last()
        )
        object_id = instance.id
        object_type = f"{app_label}.{model_name}"
        object_changed_id = object_changed.id
        object = model_to_dict(instance)

        sdk = DiodeReconcilerClient("localhost:50051", "foobar")
        sdk.add_object_state(
            object_id=object_id,
            object_type=object_type,
            object_change_id=object_changed_id,
            object=object,
        )
