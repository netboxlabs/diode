#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ApplyChangeSetView."""

import uuid

from dcim.models import (
    Device,
    DeviceRole,
    DeviceType,
    Interface,
    Manufacturer,
    Rack,
    Site,
)
from django.contrib.auth import get_user_model
from ipam.models import ASN, RIR, IPAddress
from netaddr import IPNetwork
from rest_framework import status
from users.models import Token
from utilities.testing import APITestCase
from virtualization.models import Cluster, ClusterType

User = get_user_model()


class BaseApplyChangeSet(APITestCase):
    """Base ApplyChangeSet test case."""

    def setUp(self):
        """Set up test."""
        # Necessary to use with signals.
        self.user_netbox_to_diode = User.objects.create_user(username="NETBOX_TO_DIODE")
        Token.objects.create(user=self.user_netbox_to_diode)

        self.user = User.objects.create_user(username="testcommonuser")
        self.add_permissions("netbox_diode_plugin.add_diode")
        self.user_token = Token.objects.create(user=self.user)

        self.user_header = {"HTTP_AUTHORIZATION": f"Token {self.user_token.key}"}

        rir = RIR.objects.create(name="RFC 6996", is_private=True)
        self.asns = [ASN(asn=65000 + i, rir=rir) for i in range(8)]
        ASN.objects.bulk_create(self.asns)

        self.sites = (
            Site(
                id=10,
                name="Site 1",
                slug="site-1",
                facility="Alpha",
                description="First test site",
                physical_address="123 Fake St Lincoln NE 68588",
                shipping_address="123 Fake St Lincoln NE 68588",
                comments="Lorem ipsum etcetera",
            ),
            Site(
                id=20,
                name="Site 2",
                slug="site-2",
                facility="Bravo",
                description="Second test site",
                physical_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                shipping_address="725 Cyrus Valleys Suite 761 Douglasfort NE 57761",
                comments="Lorem ipsum etcetera",
            ),
        )
        Site.objects.bulk_create(self.sites)

        self.racks = (
            Rack(name="Rack 1", site=self.sites[0]),
            Rack(name="Rack 2", site=self.sites[1]),
        )
        Rack.objects.bulk_create(self.racks)

        manufacturer = Manufacturer.objects.create(
            name="Manufacturer 1", slug="manufacturer-1"
        )

        self.device_types = (
            DeviceType(
                manufacturer=manufacturer, model="Device Type 1", slug="device-type-1"
            ),
            DeviceType(
                manufacturer=manufacturer,
                model="Device Type 2",
                slug="device-type-2",
                u_height=2,
            ),
        )
        DeviceType.objects.bulk_create(self.device_types)

        self.roles = (
            DeviceRole(name="Device Role 1", slug="device-role-1", color="ff0000"),
            DeviceRole(name="Device Role 2", slug="device-role-2", color="00ff00"),
        )
        DeviceRole.objects.bulk_create(self.roles)

        cluster_type = ClusterType.objects.create(
            name="Cluster Type 1", slug="cluster-type-1"
        )

        self.clusters = (
            Cluster(name="Cluster 1", type=cluster_type),
            Cluster(name="Cluster 2", type=cluster_type),
        )
        Cluster.objects.bulk_create(self.clusters)

        self.devices = (
            Device(
                id=10,
                device_type=self.device_types[0],
                role=self.roles[0],
                name="Device 1",
                site=self.sites[0],
                rack=self.racks[0],
                cluster=self.clusters[0],
                local_context_data={"A": 1},
            ),
            Device(
                id=20,
                device_type=self.device_types[0],
                role=self.roles[0],
                name="Device 2",
                site=self.sites[0],
                rack=self.racks[0],
                cluster=self.clusters[0],
                local_context_data={"B": 2},
            ),
        )
        Device.objects.bulk_create(self.devices)

        self.interfaces = (
            Interface(name="Interface 1", device=self.devices[0], type="1000baset"),
            Interface(name="Interface 2", device=self.devices[0], type="1000baset"),
            Interface(name="Interface 3", device=self.devices[0], type="1000baset"),
            Interface(name="Interface 4", device=self.devices[0], type="1000baset"),
            Interface(name="Interface 5", device=self.devices[0], type="1000baset"),
        )
        Interface.objects.bulk_create(self.interfaces)

        self.ip_addresses = (
            IPAddress(
                address=IPNetwork("10.0.0.1/24"), assigned_object=self.interfaces[0]
            ),
            IPAddress(
                address=IPNetwork("192.0.2.1/24"), assigned_object=self.interfaces[1]
            ),
        )
        IPAddress.objects.bulk_create(self.ip_addresses)

        self.url = "/netbox/api/plugins/diode/apply-change-set/"

    def send_request(self, payload, status_code=status.HTTP_200_OK):
        """Post the payload to the url and return the response."""
        response = self.client.post(
            self.url, data=payload, format="json", **self.user_header
        )
        self.assertEqual(response.status_code, status_code)
        return response


class ApplyChangeSetTestCase(BaseApplyChangeSet):
    """ApplyChangeSet test cases."""

    @staticmethod
    def get_change_id(payload, index):
        """Get change_id from payload."""
        return payload.get("change_set")[index].get("change_id")

    def test_change_type_create_return_200(self):
        """Test create change_type with successful."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.interface",
                    "object_id": None,
                    "data": {
                        "name": "Interface 1",
                        "device": self.devices[1].pk,
                        "type": "other",
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.163.2.1/24",
                        "assigned_object": {
                            "interface": {"id": self.interfaces[2].pk},
                        },
                    },
                },
            ],
        }

        response = self.send_request(payload)

        self.assertEqual(response.json().get("result"), "success")

    def test_change_type_update_return_200(self):
        """Test update change_type with successful."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 20,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
            ],
        }

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_updated = Site.objects.get(id=20)

        self.assertEqual(response.json().get("result"), "success")
        self.assertEqual(site_updated.name, "Site A")

    def test_change_type_create_with_error_return_400(self):
        """Test create change_type with wrong payload."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": 1,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        site_created = Site.objects.filter(name="Site A")

        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(
            response.json().get("errors")[0].get("change_id"),
            self.get_change_id(payload, 0),
        )
        self.assertIn(
            'Expected a list of items but got type "int".',
            response.json().get("errors")[0].get("asns"),
        )
        self.assertFalse(site_created.exists())

    def test_change_type_update_with_error_return_400(self):
        """Test update change_type with wrong payload."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 20,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": 1,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        site_updated = Site.objects.get(id=20)

        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(
            response.json().get("errors")[0].get("change_id"),
            self.get_change_id(payload, 0),
        )
        self.assertIn(
            'Expected a list of items but got type "int".',
            response.json().get("errors")[0].get("asns"),
        )
        self.assertEqual(site_updated.name, "Site 2")

    def test_change_type_create_with_multiples_objects_return_200(self):
        """Test create change type with two objects."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site Z",
                        "slug": "site-z",
                        "facility": "Omega",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": None,
                    "data": {
                        "device_type": self.device_types[1].pk,
                        "role": self.roles[1].pk,
                        "name": "Test Device 500",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
            ],
        }

        response = self.send_request(payload)

        self.assertEqual(response.json().get("result"), "success")

    def test_change_type_update_with_multiples_objects_return_200(self):
        """Test update change type with two objects."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 20,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": 10,
                    "data": {
                        "device_type": self.device_types[1].pk,
                        "role": self.roles[1].pk,
                        "name": "Test Device 3",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
            ],
        }

        response = self.send_request(payload)

        site_updated = Site.objects.get(id=20)
        device_updated = Device.objects.get(id=10)

        self.assertEqual(response.json().get("result"), "success")
        self.assertEqual(site_updated.name, "Site A")
        self.assertEqual(device_updated.name, "Test Device 3")

    def test_change_type_create_and_update_with_error_in_one_object_return_400(self):
        """Test create and update change type with one object with error."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site Z",
                        "slug": "site-z",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": 10,
                    "data": {
                        "device_type": 3,
                        "role": self.roles[1].pk,
                        "name": "Test Device 4",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        site_created = Site.objects.filter(name="Site Z")
        device_created = Device.objects.filter(name="Test Device 4")

        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(
            response.json().get("errors")[0].get("change_id"),
            self.get_change_id(payload, 1),
        )
        self.assertIn(
            "Related object not found using the provided numeric ID",
            response.json().get("errors")[0].get("device_type"),
        )
        self.assertFalse(site_created.exists())
        self.assertFalse(device_created.exists())

    def test_multiples_create_type_error_in_two_objects_return_400(self):
        """Test create with error in two objects."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site Z",
                        "slug": "site-z",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": [self.asns[0].pk, self.asns[1].pk],
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": None,
                    "data": {
                        "device_type": 3,
                        "role": self.roles[1].pk,
                        "name": "Test Device 4",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": None,
                    "data": {
                        "device_type": 100,
                        "role": 10,
                        "name": "Test Device 40",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        site_created = Site.objects.filter(name="Site Z")
        device_created = Device.objects.filter(name="Test Device 4")

        self.assertEqual(response.json().get("result"), "failed")

        self.assertEqual(
            response.json().get("errors")[0].get("change_id"),
            self.get_change_id(payload, 1),
        )
        self.assertIn(
            "Related object not found using the provided numeric ID",
            response.json().get("errors")[0].get("device_type"),
        )

        self.assertEqual(
            response.json().get("errors")[1].get("change_id"),
            self.get_change_id(payload, 2),
        )
        self.assertIn(
            "Related object not found using the provided numeric ID",
            response.json().get("errors")[1].get("device_type"),
        )

        self.assertFalse(site_created.exists())
        self.assertFalse(device_created.exists())

    def test_change_type_update_with_object_id_not_exist_return_400(self):
        """Test update object with nonexistent object_id."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 30,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": 1,
                    },
                },
            ],
        }

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_updated = Site.objects.get(id=20)

        self.assertEqual(response.json()[0], "object with id 30 does not exist")
        self.assertEqual(site_updated.name, "Site 2")

    def test_change_set_id_field_not_provided_return_400(self):
        """Test update object with change_set_id incorrect."""
        payload = {
            "change_set_id": None,
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 20,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": 1,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertIsNone(response.json().get("errors")[0].get("change_id"))
        self.assertEqual(
            response.json().get("errors")[0].get("change_set_id"),
            "This field may not be null.",
        )

    def test_change_set_id_change_id_and_change_type_field_not_provided_return_400(
        self,
    ):
        """Test update object with change_set_id, change_id, and change_type incorrect."""
        payload = {
            "change_set_id": "",
            "change_set": [
                {
                    "change_id": "",
                    "change_type": "",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 20,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                        "asns": 1,
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertEqual(
            response.json().get("errors")[0].get("change_set_id"),
            "Must be a valid UUID.",
        )
        self.assertEqual(
            response.json().get("errors")[1].get("change_id"),
            "Must be a valid UUID.",
        )
        self.assertEqual(
            response.json().get("errors")[1].get("change_type"),
            "This field may not be blank.",
        )

    def test_change_set_id_field_and_change_set_not_provided_return_400(self):
        """Test update object with change_set_id and change_set incorrect."""
        payload = {
            "change_set_id": "",
            "change_set": [],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertEqual(
            response.json().get("errors")[0].get("change_set_id"),
            "Must be a valid UUID.",
        )
        self.assertEqual(
            response.json().get("errors")[1].get("change_set"),
            "This list may not be empty.",
        )

    def test_change_type_and_object_type_provided_return_400(
        self,
    ):
        """Test change_type and object_type incorrect."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": None,
                    "object_version": None,
                    "object_type": "",
                    "object_id": None,
                    "data": {
                        "name": "Site A",
                        "slug": "site-a",
                        "facility": "Alpha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                    },
                },
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": None,
                    "data": {
                        "name": "Site Z",
                        "slug": "site-z",
                        "facility": "Betha",
                        "description": "",
                        "physical_address": "123 Fake St Lincoln NE 68588",
                        "shipping_address": "123 Fake St Lincoln NE 68588",
                        "comments": "Lorem ipsum etcetera",
                    },
                },
            ],
        }

        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        # First item of change_set
        self.assertEqual(
            response.json().get("errors")[0].get("change_id"),
            self.get_change_id(payload, 0),
        )
        self.assertEqual(
            response.json().get("errors")[0].get("change_type"),
            "This field may not be null.",
        )
        self.assertEqual(
            response.json().get("errors")[0].get("object_type"),
            "This field may not be blank.",
        )

        # Second item of change_set
        self.assertEqual(
            response.json().get("errors")[1].get("change_id"),
            self.get_change_id(payload, 1),
        )
        self.assertEqual(
            response.json().get("errors")[1].get("change_type"),
            "This field may not be blank.",
        )

    def test_create_ip_address_return_200(self):
        """Test create ip_address with successful."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.161.3.1/24",
                        "assigned_object": {
                            "interface": {
                                "name": self.interfaces[3].name,
                                "device": {
                                    "name": self.devices[0].name,
                                    "site": {"name": self.sites[0].name},
                                },
                            },
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload)

        self.assertEqual(response.json().get("result"), "success")

    def test_create_ip_address_return_400(self):
        """Test create ip_address with missing interface name."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.161.3.1/24",
                        "assigned_object": {
                            "interface": {
                                # Forcing to miss the name of the interface
                                "device": {
                                    "name": self.devices[0].name,
                                    "site": {"name": self.sites[0].name},
                                },
                            },
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertIn(
            "not sufficient to retrieve interface",
            response.json().get("errors")[0].get("assigned_object"),
        )

    def test_create_ip_address_not_exist_interface_return_400(self):
        """Test create ip_address with not valid interface."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.161.3.1/24",
                        "assigned_object": {
                            "interface": {
                                "name": "not_exist",
                                "device": {
                                    "name": self.devices[0].name,
                                    "site": {"name": self.sites[0].name},
                                },
                            },
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertIn(
            "does not exist",
            response.json().get("errors")[0].get("assigned_object"),
        )

    def test_create_ip_address_missing_device_interface_return_400(self):
        """Test create ip_address with missing device interface name."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.161.3.1/24",
                        "assigned_object": {
                            "interface": {
                                "name": "not_exist",
                                "device": {
                                    "site": {"name": self.sites[0].name},
                                },
                            },
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertIn(
            "Interface device needs to have either id or name provided",
            response.json().get("errors")[0].get("assigned_object"),
        )

    def test_create_ip_address_missing_interface_device_site_return_400(self):
        """Test create ip_address with missing interface device site name."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "ipam.ipaddress",
                    "object_id": None,
                    "data": {
                        "address": "192.161.3.1/24",
                        "assigned_object": {
                            "interface": {
                                "name": "not_exist",
                                "device": {
                                    "name": self.devices[0].name,
                                    "site": {"facility": "Betha"},
                                },
                            },
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertIn(
            "Interface device site needs to have either id or name provided",
            response.json().get("errors")[0].get("assigned_object"),
        )

    def test_primary_ip_address_not_found_return_400(self):
        """Test update primary ip address with site name."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "data": {
                        "name": self.devices[0].name,
                        "site": {"name": self.sites[0].name},
                        "primary_ip6": {
                            "address": "2001:DB8:0000:0000:244:17FF:FEB6:D37D/64",
                        },
                    },
                },
            ],
        }
        response = self.send_request(payload, status_code=status.HTTP_400_BAD_REQUEST)

        self.assertEqual(response.json()[0], "primary IP not found")

    def test_add_primary_ip_address_to_device(self):
        """Add primary ip address to device."""
        payload = {
            "change_set_id": str(uuid.uuid4()),
            "change_set": [
                {
                    "change_id": str(uuid.uuid4()),
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "data": {
                        "name": self.devices[0].name,
                        "site": {"name": self.sites[0].name},
                        "primary_ip4": {
                            "address": str(self.ip_addresses[0].address),
                            "assigned_object": {
                                "interface": {
                                    "name": self.interfaces[0].name,
                                    "device": {
                                        "name": self.devices[0].name,
                                        "site": {"name": self.sites[0].name},
                                    },
                                },
                            },
                        },
                    },
                },
            ],
        }

        response = self.send_request(payload)

        device_updated = Device.objects.get(id=10)

        self.assertEqual(response.json().get("result"), "success")
        self.assertEqual(device_updated.name, self.devices[0].name)
        self.assertEqual(device_updated.primary_ip4, self.ip_addresses[0])
