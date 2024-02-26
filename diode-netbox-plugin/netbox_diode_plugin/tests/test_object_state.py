#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ObjectStateView."""

from dcim.models import Site
from django.core.management import call_command
from rest_framework import status
from utilities.testing import APITestCase


class ObjectStateTestCase(APITestCase):
    """ObjectState test cases."""

    def setUp(self):
        """Set up test."""
        sites = (
            Site(
                id=1,
                name="Site 1",
                slug="site-1",
                facility="Alpha",
                description="First test site",
                physical_address="123 Fake St Lincoln NE 68588",
                shipping_address="123 Fake St Lincoln NE 68588",
                comments="Lorem ipsum etcetera",
            ),
            Site(
                id=2,
                name="Site 2",
                slug="site-2",
                facility="Bravo",
                description="Second test site",
                physical_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                shipping_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                comments="Lorem ipsum etcetera",
            ),
            Site(
                id=3,
                name="Site 3",
                slug="site-3",
                facility="Charlie",
                description="Third test site",
                physical_address="2321 Dovie Dale East Cristobal AK 71959",
                shipping_address="2321 Dovie Dale East Cristobal AK 71959",
                comments="Lorem ipsum etcetera",
            ),
        )
        Site.objects.bulk_create(sites)

        # call_command is because the searching using q parameter uses CachedValue to get the object ID
        call_command("reindex")

        self.url = "/api/plugins/diode/object-state/"

    def test_return_object_state_using_id(self):
        """Test searching using id parameter."""
        query_parameters = {"id": 1, "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("name"), "Site 1")

    def test_return_object_state_using_q(self):
        """Test searching using q parameter."""
        query_parameters = {"q": "Site 2", "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("name"), "Site 2")

    def test_object_not_found_return_empty(self):
        """Test empty searching."""
        query_parameters = {"q": "Site 10", "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json(), {})

    def test_missing_object_type_return_400(self):
        """Test API behavior with missing object type."""
        query_parameters = {"q": "Site 10", "object_type": ""}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_missing_q_and_id_parameters_return_400(self):
        """Test API behavior with missing q and ID parameters."""
        query_parameters = {"object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
