#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Views."""

from django.contrib.contenttypes.models import ContentType
from django.db import transaction
from django.db.models import Q
from extras.models import CachedValue
from netbox.search import LookupTypes
from rest_framework import status, views
from rest_framework.exceptions import ValidationError
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from utilities.api import get_serializer_for_model

from netbox_diode_plugin.api.permissions import IsDiodePost, IsDiodeViewer
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

    permission_classes = [IsAuthenticated, IsDiodePost]

    @staticmethod
    def _get_object_type_model(object_type):
        """Get the object type model from object_type."""
        app_label, model_name = object_type.split(".")
        object_content_type = ContentType.objects.get_by_natural_key(
            app_label, model_name
        )
        object_type_model = object_content_type.model_class()
        return object_type_model

    @staticmethod
    @transaction.atomic()
    def _perform_save(serializer):
        """Save an object and rollback if any problem."""
        serializer.save()

    @transaction.atomic(savepoint=True)
    def _validate_serializer(
        self, change_type, object_id, object_type_model, object_data
    ):
        """Validate the serializer."""
        if change_type == "create":
            serializer = get_serializer_for_model(object_type_model)(
                data=object_data, context={"request": self.request}
            )
        elif change_type == "update":
            if not object_id:
                raise ValidationError("object_id parameter is required")
            try:
                instance = object_type_model.objects.get(id=object_id)
            except object_type_model.DoesNotExist:
                return "failed", f"Object with id {object_id} does not exist"

            serializer = get_serializer_for_model(object_type_model)(
                instance, data=object_data, context={"request": self.request}
            )
        else:
            return "failed", "Invalid change_type"

        if not serializer.is_valid():
            return "failed", serializer.errors

        self.check_object_permissions(self.request, serializer)

        return "success", serializer

    def post(self, request, *args, **kwargs):
        """
        Create a new change set and apply it to the current state.

        The request body should contain a list of changes to be applied.
        """
        change_set_id = self.request.data.get("change_set_id", None)
        change_set = self.request.data.get("change_set", None)
        if not change_set:
            raise ValidationError("change_set parameter is required")

        serializer_list = []

        for change in change_set:

            change_id = change.get("change_id", None)
            change_type = change.get("change_type", None)

            object_type = change.get("object_type", None)
            if not object_type:
                raise ValidationError("object_type parameter is required")

            object_type_model = self._get_object_type_model(object_type)

            object_data = change.get("data", None)
            if not object_data:
                raise ValidationError("data parameter is required")

            object_id = change.get("object_id", None)

            result, serializer_result = self._validate_serializer(
                change_type, object_id, object_type_model, object_data
            )

            if result == "failed":
                return Response(
                    {
                        "change_set_id": change_set_id,
                        "change_id": change_id,
                        "result": "failed",
                        "error": serializer_result,
                    },
                    status=status.HTTP_400_BAD_REQUEST,
                )

            serializer_list.append(serializer_result)

        [self._perform_save(serializer) for serializer in serializer_list]

        data = {"change_set_id": change_set_id, "result": "success"}
        return Response(data, status=status.HTTP_200_OK)
