#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Views."""

from django.contrib.contenttypes.models import ContentType
from django.db.models import Q
from extras.models import CachedValue
from netbox.search import LookupTypes
from rest_framework import views, generics, status
from rest_framework.exceptions import ValidationError
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from utilities.api import get_serializer_for_model

from netbox_diode_plugin.api.permissions import IsDiodeViewer
from netbox_diode_plugin.api.serializers import ObjectStateSerializer


class ObjectStateView(views.APIView):
    """ObjectState view."""

    permission_classes = [IsAuthenticated, IsDiodeViewer]

    def get(self, request, *args, **kwargs):
        """
        Return a JSON with object_type, object_change_id, and object.

        Search for objects according to object type.
        If the obj_type parameter is not in the parameters, raise a ValidationError.
        When object ID is provided in the request, search using it in the model specified by object type.
        If ID is not provided, use the q parameter for searching.
        Lookup is iexact
        """
        object_type = self.request.query_params.get("object_type", None)

        if not object_type:
            raise ValidationError("object_type parameter is required")

        app_label, model_name = object_type.split(".")
        object_content_type = ContentType.objects.get_by_natural_key(
            app_label, model_name
        )
        object_type_model = object_content_type.model_class()

        object_id = self.request.query_params.get("id", None)

        if object_id:
            queryset = object_type_model.objects.filter(id=object_id)
        else:
            lookup = LookupTypes.EXACT
            search_value = self.request.query_params.get("q", None)
            if not search_value:
                raise ValidationError("id or q parameter is required")

            query_filter = Q(**{f"value__{lookup}": search_value})
            query_filter &= Q(object_type__in=[object_content_type])

            object_id_in_cached_value = CachedValue.objects.filter(
                query_filter
            ).values_list("object_id", flat=True)

            queryset = object_type_model.objects.filter(
                id__in=object_id_in_cached_value
            )

        self.check_object_permissions(request, queryset)

        serializer = ObjectStateSerializer(
            queryset,
            many=True,
            context={
                "request": request,
                "object_type": f"{object_type}",
            },
        )

        if len(serializer.data) > 0:
            return Response(serializer.data[0])
        return Response({})


class ApplyChangeSetView(views.APIView):
    """ApplyChangeSet view."""

    authentication_classes = []
    permission_classes = []

    # def get_serializer(self, *args, **kwargs):
    #     """
    #     Return the serializer instance that should be used for validating and
    #     deserializing input, and for serializing output.
    #     """
    #     print(self.request.data)
    #     serializer_class = self.get_serializer_class()
    #     kwargs.setdefault("context", self.get_serializer_context())
    #     return serializer_class(*args, **kwargs)

    def post(self, request, *args, **kwargs):
        """
        Create a new change set and apply it to the current state.

        The request body should contain a list of changes to be applied.
        """
        change_set = self.request.data.get("change_set", None)
        if not change_set:
            raise ValidationError("change_set parameter is required")

        for change in change_set:
            change_type = change.get("change_type", None)

            object_type = change.get("object_type", None)
            if not object_type:
                raise ValidationError("object_type parameter is required")
            app_label, model_name = object_type.split(".")
            object_content_type = ContentType.objects.get_by_natural_key(
                app_label, model_name
            )
            object_type_model = object_content_type.model_class()

            if change_type == "create":
                object_data = change.get("data", None)
                if not object_data:
                    raise ValidationError("data parameter is required")

                serializer = get_serializer_for_model(object_type_model)(
                    data=object_data, context={"request": request}
                )

                serializer.is_valid(raise_exception=True)

                serializer.save()

                return Response(
                    serializer.data,
                    status=status.HTTP_201_CREATED,
                )

        return Response({"status": "success"})
