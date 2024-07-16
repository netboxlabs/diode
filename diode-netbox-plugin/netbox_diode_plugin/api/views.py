#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API Views."""
from typing import Any, Dict, Optional

from django.conf import settings
from packaging import version

if version.parse(settings.VERSION).major >= 4:
    from core.models import ObjectType as NetBoxType
else:
    from django.contrib.contenttypes.models import ContentType as NetBoxType
from django.core.exceptions import FieldError
from django.db import transaction
from django.db.models import Q
from extras.models import CachedValue
from netbox.search import LookupTypes
from rest_framework import status, views
from rest_framework.exceptions import ValidationError
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from utilities.api import get_serializer_for_model

from netbox_diode_plugin.api.permissions import IsDiodeReader, IsDiodeWriter
from netbox_diode_plugin.api.serializers import (
    ApplyChangeSetRequestSerializer,
    ObjectStateSerializer,
)


class ObjectStateView(views.APIView):
    """ObjectState view."""

    permission_classes = [IsAuthenticated, IsDiodeReader]

    def _get_lookups(self, object_type_model: str) -> tuple:
        """
        This method returns a tuple of related object lookups based on the provided object type model.

        Args:
        ----
            object_type_model (str): The name of the object type model.

        Returns:
        -------
            tuple: A tuple of related object lookups. The tuple is empty if the object type model does not match any
            of the specified models.

        """
        if object_type_model == "ipaddress":
            return "interface", "interface__device", "interface__device__site"
        if object_type_model == "interface":
            return "device", "device__site"
        if object_type_model == "device":
            return ("site",)
        return ()

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
        object_content_type = NetBoxType.objects.get_by_natural_key(
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

            lookups = self._get_lookups(object_type_model)

            if lookups:
                queryset = queryset.prefetch_related(*lookups)

            additional_attributes_query_filter = (
                self._additional_attributes_query_filter()
            )

            if additional_attributes_query_filter:
                try:
                    queryset = queryset.filter(**additional_attributes_query_filter)
                except (FieldError, ValueError):
                    return Response(
                        {"errors": ["invalid additional attributes provided"]},
                        status=status.HTTP_400_BAD_REQUEST,
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

    def _additional_attributes_query_filter(self):
        additional_attributes = {}
        for attr in self.request.query_params:
            if attr not in ["object_type", "id", "q"]:
                additional_attributes[attr] = self.request.query_params.get(attr)

        return dict(additional_attributes.items())


class ApplyChangeSetView(views.APIView):
    """ApplyChangeSet view."""

    permission_classes = [IsAuthenticated, IsDiodeWriter]

    @staticmethod
    def _get_object_type_model(object_type: str):
        """Get the object type model from object_type."""
        app_label, model_name = object_type.split(".")
        object_content_type = NetBoxType.objects.get_by_natural_key(
            app_label, model_name
        )
        return object_content_type.model_class()

    def _get_assigned_object_type(self, model_name: str):
        """Get the object type model from applied IPAddress assigned object."""
        assignable_object_types = {
            "interface": "dcim.interface",
        }
        return assignable_object_types.get(model_name.lower(), None)

    def _get_serializer(
        self,
        change_type: str,
        object_id: int,
        object_type: str,
        object_data: dict,
        change_set_id: str,
    ):
        """Get the serializer for the object type."""
        object_type_model = self._get_object_type_model(object_type)
        if change_type == "create":
            serializer = get_serializer_for_model(object_type_model)(
                data=object_data, context={"request": self.request}
            )
        elif change_type == "update":
            lookups = ()
            args = {}

            primary_ip_to_set: Optional[dict] = None

            if object_id:
                args["id"] = object_id
            elif object_type == "dcim.device" and any(
                object_data.get(attr) for attr in ("primary_ip4", "primary_ip6")
            ):
                ip_address = self._retrieve_primary_ip_address(
                    "primary_ip4", object_data
                )

                if ip_address is None:
                    ip_address = self._retrieve_primary_ip_address(
                        "primary_ip6", object_data
                    )

                if ip_address is None:
                    raise ValidationError("primary IP not found")

                if ip_address:
                    primary_ip_to_set = {
                        "id": ip_address.id,
                        "family": ip_address.family,
                    }

                lookups = ("site",)
                args["name"] = object_data.get("name")
                args["site__name"] = object_data.get("site").get("name")
            else:
                raise ValidationError("object_id parameter is required")

            try:
                instance = object_type_model.objects.prefetch_related(*lookups).get(
                    **args
                )
                if object_type == "dcim.device" and primary_ip_to_set:
                    object_data = {
                        "id": instance.id,
                        "device_type": instance.device_type.id,
                        "role": instance.role.id,
                        "site": instance.site.id,
                        f'primary_ip{primary_ip_to_set.get("family")}': primary_ip_to_set.get(
                            "id"
                        ),
                    }
            except object_type_model.DoesNotExist:
                raise ValidationError(f"object with id {object_id} does not exist")

            serializer = get_serializer_for_model(object_type_model)(
                instance, data=object_data, context={"request": self.request}
            )
        else:
            raise ValidationError("Invalid change_type")
        return serializer

    def _retrieve_primary_ip_address(self, primary_ip_attr: str, object_data: dict):
        ip_address = object_data.get(primary_ip_attr)
        if ip_address is None:
            return None

        ipaddress_assigned_object = object_data.get(primary_ip_attr, {}).get(
            "assigned_object", None
        )
        if ipaddress_assigned_object is None:
            return None

        interface = ipaddress_assigned_object.get("interface")
        if interface is None:
            return None

        interface_device = interface.get("device")
        if interface_device is None:
            return None

        ip_address_object = self._get_object_type_model("ipam.ipaddress").objects.get(
            address=ip_address.get("address"),
            interface__name=interface.get("name"),
            interface__device__name=interface_device.get("name"),
            interface__device__site__name=interface_device.get("site").get("name"),
        )
        return ip_address_object

    @staticmethod
    def _get_error_response(change_set_id, error):
        return Response(
            {
                "change_set_id": change_set_id,
                "result": "failed",
                "errors": error,
            },
            status=status.HTTP_400_BAD_REQUEST,
        )

    def _ipaddress_assigned_object(self, change_set: list) -> list:
        ipaddress_assigned_object = [
            change.get("data").get("assigned_object", None)
            for change in change_set
            if change.get("object_type") == "ipam.ipaddress"
            and change.get("data", {}).get("assigned_object", None)
        ]

        return ipaddress_assigned_object

    def _retrieve_assigned_object_interface_device_lookup_args(
        self, device: dict
    ) -> dict:
        """
        This method retrieves the lookup arguments for the interface device of an assigned object.

        Args:
        ----
            device (dict): A dictionary containing the details of the device. It should contain either 'id' or 'name'
                of the device and 'site' which is another dictionary containing either 'id' or 'name' of the site.

        Returns:
        -------
            dict: A dictionary containing the lookup arguments for the interface device.

        Raises:
        ------
            ValidationError: If neither 'id' nor 'name' is provided for the device or the site.

        """
        args = {}
        if device.get("id"):
            args["device__id"] = device.get("id")
        elif device.get("name"):
            args["device__name"] = device.get("name")
        else:
            raise ValidationError(
                "Interface device needs to have either id or name provided"
            )

        site = device.get("site", {})
        if site:
            if site.get("id"):
                args["device__site__id"] = site.get("id")
            elif site.get("name"):
                args["device__site__name"] = site.get("name")
            else:
                raise ValidationError(
                    "Interface device site needs to have either id or name provided"
                )
        return args

    def _handle_ipaddress_assigned_object(
        self, object_data: dict, ipaddress_assigned_object: list
    ) -> Optional[Dict[str, Any]]:
        """Handle IPAM IP address assigned object."""
        if any(ipaddress_assigned_object):
            assigned_object_keys = list(ipaddress_assigned_object[0].keys())
            model_name = assigned_object_keys[0]
            assigned_object_type = self._get_assigned_object_type(model_name)
            assigned_object_model = self._get_object_type_model(assigned_object_type)
            assigned_object_properties_dict = dict(
                ipaddress_assigned_object[0][model_name].items()
            )

            if len(assigned_object_properties_dict) == 0:
                return {"assigned_object": f"properties not provided for {model_name}"}

            try:
                lookups = (
                    ("device", "device__site") if model_name == "interface" else ()
                )
                args = {}

                if model_name == "interface":
                    if assigned_object_properties_dict.get("id"):
                        args["id"] = assigned_object_properties_dict.get("id")
                    elif assigned_object_properties_dict.get("name"):
                        try:
                            device = assigned_object_properties_dict.get("device", {})
                            args = self._retrieve_assigned_object_interface_device_lookup_args(
                                device
                            )
                            args["name"] = assigned_object_properties_dict.get("name")
                        except ValidationError as e:
                            return {"assigned_object": str(e)}
                    else:
                        return {
                            "assigned_object": f"provided properties '{assigned_object_properties_dict}' not sufficient to retrieve {model_name}"
                        }

                assigned_object_instance = (
                    assigned_object_model.objects.prefetch_related(*lookups).get(**args)
                )
            except assigned_object_model.DoesNotExist:
                return {
                    "assigned_object": f"Assigned object with name {ipaddress_assigned_object[0][model_name]} does not exist"
                }

            object_data.pop("assigned_object")
            object_data["assigned_object_type"] = assigned_object_type
            object_data["assigned_object_id"] = assigned_object_instance.id
        return None

    def post(self, request, *args, **kwargs):
        """
        Create a new change set and apply it to the current state.

        The request body should contain a list of changes to be applied.
        """
        serializer_errors = []

        request_serializer = ApplyChangeSetRequestSerializer(data=request.data)

        change_set_id = self.request.data.get("change_set_id", None)

        if not request_serializer.is_valid():
            for field_error_name in request_serializer.errors:
                self._extract_serializer_errors(
                    field_error_name, request_serializer, serializer_errors
                )

            return self._get_error_response(change_set_id, serializer_errors)

        change_set = request_serializer.data.get("change_set", None)

        ipaddress_assigned_object = self._ipaddress_assigned_object(change_set)

        try:
            with transaction.atomic():
                for change in change_set:
                    change_id = change.get("change_id", None)
                    change_type = change.get("change_type", None)
                    object_type = change.get("object_type", None)
                    object_data = change.get("data", None)
                    object_id = change.get("object_id", None)

                    errors = None
                    if (
                        any(ipaddress_assigned_object)
                        and object_type == "ipam.ipaddress"
                    ):
                        errors = self._handle_ipaddress_assigned_object(
                            object_data, ipaddress_assigned_object
                        )

                    if errors is not None:
                        serializer_errors.append({"change_id": change_id, **errors})
                        continue

                    serializer = self._get_serializer(
                        change_type, object_id, object_type, object_data, change_set_id
                    )

                    if serializer.is_valid():
                        serializer.save()
                    else:
                        errors_dict = {
                            field_name: f"{field_name}: {str(field_errors[0])}"
                            for field_name, field_errors in serializer.errors.items()
                        }

                        serializer_errors.append(
                            {"change_id": change_id, **errors_dict}
                        )
                if len(serializer_errors) > 0:
                    raise ApplyChangeSetException
        except ApplyChangeSetException:
            return self._get_error_response(change_set_id, serializer_errors)

        data = {"change_set_id": change_set_id, "result": "success"}
        return Response(data, status=status.HTTP_200_OK)

    def _extract_serializer_errors(
        self, field_error_name, request_serializer, serializer_errors
    ):
        if isinstance(request_serializer.errors[field_error_name], dict):
            for error_index, error_values in request_serializer.errors[
                field_error_name
            ].items():
                errors_dict = {
                    "change_id": request_serializer.data.get("change_set")[
                        error_index
                    ].get("change_id")
                }

                for field_name, field_errors in error_values.items():
                    errors_dict[field_name] = f"{str(field_errors[0])}"

                serializer_errors.append(errors_dict)
        else:
            errors = {
                field_error_name: f"{str(field_errors)}"
                for field_errors in request_serializer.errors[field_error_name]
            }

            serializer_errors.append(errors)


class ApplyChangeSetException(Exception):
    """ApplyChangeSetException used to cause atomic transaction rollback."""

    pass
