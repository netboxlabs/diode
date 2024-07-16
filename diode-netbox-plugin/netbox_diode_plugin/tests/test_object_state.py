#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ObjectStateView."""

from dcim.models import Device, DeviceRole, DeviceType, Manufacturer, Rack, Site
from django.contrib.auth import get_user_model
from django.core.management import call_command
from rest_framework import status
from users.models import Token
from utilities.testing import APITestCase
from virtualization.models import Cluster, ClusterType

User = get_user_model()


class ObjectStateTestCase(APITestCase):
    """ObjectState test cases."""

    @classmethod
    def setUpClass(cls):
        """Set up class."""
        super().setUpClass()

        cls.sites = (
            Site(
                name="Site 1",
                slug="site-1",
                facility="Alpha",
                description="First test site",
                physical_address="123 Fake St Lincoln NE 68588",
                shipping_address="123 Fake St Lincoln NE 68588",
                comments="Lorem ipsum etcetera",
            ),
            Site(
                name="Site 2",
                slug="site-2",
                facility="Bravo",
                description="Second test site",
                physical_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                shipping_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                comments="Lorem ipsum etcetera",
            ),
            Site(
                name="Site 3",
                slug="site-3",
                facility="Charlie",
                description="Third test site",
                physical_address="2321 Dovie Dale East Cristobal AK 71959",
                shipping_address="2321 Dovie Dale East Cristobal AK 71959",
                comments="Lorem ipsum etcetera",
            ),
        )
        Site.objects.bulk_create(cls.sites)

        cls.manufacturer = (
            Manufacturer(name="Cisco", slug="cisco"),
            Manufacturer(name="Manufacturer 2", slug="manufacturer-2"),
        )

        Manufacturer.objects.bulk_create(cls.manufacturer)

        cls.device_types = (
            DeviceType(
                manufacturer=cls.manufacturer[0],
                model="ISR4321",
                slug="isr4321",
            ),
            DeviceType(
                manufacturer=cls.manufacturer[1],
                model="ISR4321",
                slug="isr4321",
            ),
            DeviceType(
                manufacturer=cls.manufacturer[1],
                model="Device Type 2",
                slug="device-type-2",
                u_height=2,
            ),
        )
        DeviceType.objects.bulk_create(cls.device_types)

        cls.roles = (
            DeviceRole(name="Device Role 1", slug="device-role-1", color="ff0000"),
            DeviceRole(name="Device Role 2", slug="device-role-2", color="00ff00"),
        )
        DeviceRole.objects.bulk_create(cls.roles)

        cls.racks = (
            Rack(name="Rack 1", site=cls.sites[0]),
            Rack(name="Rack 2", site=cls.sites[1]),
        )
        Rack.objects.bulk_create(cls.racks)

        cluster_type = ClusterType.objects.create(
            name="Cluster Type 1", slug="cluster-type-1"
        )

        cls.clusters = (
            Cluster(name="Cluster 1", type=cluster_type),
            Cluster(name="Cluster 2", type=cluster_type),
        )
        Cluster.objects.bulk_create(cls.clusters)

        cls.devices = (
            Device(
                id=10,
                device_type=cls.device_types[0],
                role=cls.roles[0],
                name="Device 1",
                site=cls.sites[0],
                rack=cls.racks[0],
                cluster=cls.clusters[0],
                local_context_data={"A": 1},
            ),
            Device(
                id=20,
                device_type=cls.device_types[0],
                role=cls.roles[0],
                name="Device 2",
                site=cls.sites[0],
                rack=cls.racks[0],
                cluster=cls.clusters[0],
                local_context_data={"B": 2},
            ),
        )
        Device.objects.bulk_create(cls.devices)

        # call_command is because the searching using q parameter uses CachedValue to get the object ID
        call_command("reindex")

    def setUp(self):
        """Set up test."""
        self.root_user = User.objects.create_user(
            username="root_user", is_staff=True, is_superuser=True
        )
        self.root_token = Token.objects.create(user=self.root_user)

        self.user = User.objects.create_user(username="testcommonuser")
        self.add_permissions("netbox_diode_plugin.view_diode")
        self.user_token = Token.objects.create(user=self.user)

        # another_user does not have permission.
        self.another_user = User.objects.create_user(username="another_user")
        self.another_user_token = Token.objects.create(user=self.another_user)

        self.root_header = {"HTTP_AUTHORIZATION": f"Token {self.root_token.key}"}
        self.user_header = {"HTTP_AUTHORIZATION": f"Token {self.user_token.key}"}
        self.another_user_header = {
            "HTTP_AUTHORIZATION": f"Token {self.another_user_token.key}"
        }

        self.url = "/netbox/api/plugins/diode/object-state/"

    def test_return_object_state_using_id(self):
        """Test searching using id parameter - Root User."""
        site_id = Site.objects.get(name=self.sites[0]).id
        query_parameters = {"id": site_id, "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("name"), "Site 1")

    def test_return_object_state_using_q(self):
        """Test searching using q parameter - Root User."""
        query_parameters = {"q": "Site 2", "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("name"), "Site 2")

    def test_object_not_found_return_empty(self):
        """Test empty searching - Root User."""
        query_parameters = {"q": "Site 10", "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json(), {})

    def test_missing_object_type_return_400(self):
        """Test API behavior with missing object type - Root User."""
        query_parameters = {"q": "Site 10", "object_type": ""}

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_missing_q_and_id_parameters_return_400(self):
        """Test API behavior with missing q and ID parameters - Root User."""
        query_parameters = {"object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_request_user_not_authenticated_return_403(self):
        """Test API behavior with user unauthenticated."""
        query_parameters = {"id": 1, "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters)

        self.assertEqual(response.status_code, status.HTTP_403_FORBIDDEN)

    def test_common_user_with_permissions_get_object_state_using_id(self):
        """Test searching using id parameter for Common User with permission."""
        site_id = Site.objects.get(name=self.sites[0]).id
        query_parameters = {"id": site_id, "object_type": "dcim.site"}

        response = self.client.get(self.url, query_parameters, **self.user_header)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("name"), "Site 1")

    def test_common_user_without_permissions_get_object_state_using_id_return_403(self):
        """
        Test searching using id parameter for Common User without permission.

        User has no permissions.
        """
        query_parameters = {"id": 1, "object_type": "dcim.device"}

        response = self.client.get(
            self.url, query_parameters, **self.another_user_header
        )

        self.assertEqual(response.status_code, status.HTTP_403_FORBIDDEN)

    def test_return_object_state_using_q_objects_with_different_manufacturer_return_cisco_manufacturer(
        self,
    ):
        """Test searching using q parameter - DevicesTypes with different manufacturer."""
        query_parameters = {
            "q": "ISR4321",
            "object_type": "dcim.devicetype",
            "manufacturer__name": "Cisco",
        }

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.json().get("object").get("model"), "ISR4321")
        self.assertEqual(
            response.json().get("object").get("manufacturer").get("name"), "Cisco"
        )

    def test_invalid_object_state_using_q_objects_and_wrong_additional_attributes_return_400(
        self,
    ):
        """Test searching using q parameter - invalid additional attributes."""
        query_parameters = {
            "q": "ISR4321",
            "object_type": "dcim.devicetype",
            "attr_name": "manufacturer.name",
            "attr_value": "Cisco",
        }

        response = self.client.get(self.url, query_parameters, **self.root_header)

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
