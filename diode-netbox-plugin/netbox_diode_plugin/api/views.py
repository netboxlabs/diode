#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Views."""

from django.contrib.contenttypes.models import ContentType
from django.db.models import Q
from extras.models import CachedValue
from netbox.search import LookupTypes
from rest_framework import views
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from netbox_diode_plugin.api.serializers import ObjectStateSerializer


class ObjectStateView(views.APIView):
    """ObjectState view."""

    authentication_classes = []  # disables authentication
    permission_classes = []

    def get(self, request, *args, **kwargs):
        """
        Return a JSON with object_type, object_change_id, and object.

        Search for objects according to object type.
        If the obj_type parameter is not in the parameters, raise a ValidationError.
        When object ID is provided in the request, search using it in the model specified by object type.
        If ID is not provided, use the q parameter for searching.
        Lookup is iexact
        """
        obj_type = self.request.query_params.get("obj_type", None)

        if not obj_type:
            raise ValidationError("obj_type parameter is required")

        app_label, model_name = obj_type.split(".")
        object_type = ContentType.objects.get_by_natural_key(app_label, model_name)
        object_type_model = object_type.model_class()

        object_id = self.request.query_params.get("id", None)

        if object_id:
            queryset = object_type_model.objects.filter(id=object_id)
        else:
            lookup = LookupTypes.EXACT
            search_value = self.request.query_params.get("q", None)
            if not search_value:
                raise ValidationError("id or q parameter is required")

            query_filter = Q(**{f"value__{lookup}": search_value})
            query_filter &= Q(object_type__in=[object_type])

            object_id_in_cached_value = CachedValue.objects.filter(
                query_filter
            ).values_list("object_id", flat=True)

            queryset = object_type_model.objects.filter(
                id__in=object_id_in_cached_value
            )

        serializer = ObjectStateSerializer(
            queryset,
            many=True,
            context={
                "request": request,
                "object_type": f"{obj_type}",
            },
        )

        if len(serializer.data) > 0:
            return Response(serializer.data[0])
        return Response(serializer.data)
