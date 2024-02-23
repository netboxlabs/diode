#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Serializer."""

from extras.models import ObjectChange
from rest_framework import serializers
from utilities.api import get_serializer_for_model


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
        serializer = get_serializer_for_model(instance)

        object_data = instance.__class__.objects.filter(id=instance.id)

        context = {"request": self.context.get("request")}
        return serializer(object_data, context=context, many=True).data[0]
