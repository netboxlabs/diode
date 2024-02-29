#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ApplyChangeSetView."""

from dcim.models import (
    Device,
    DeviceRole,
    DeviceType,
    Manufacturer,
    Rack,
    Site,
)
from django.contrib.auth import get_user_model
from ipam.models import ASN, RIR
from users.models import Token
from utilities.testing import APITestCase
from virtualization.models import Cluster, ClusterType

User = get_user_model()


class BaseApplyChangeSet(APITestCase):
    """Base ApplyChangeSet test case."""

    def setUp(self):
        """Set up test."""
        self.user = User.objects.create_user(username="testcommonuser")
        self.add_permissions("netbox_diode_plugin.add_objectstate")
        self.user_token = Token.objects.create(user=self.user)

        self.user_header = {"HTTP_AUTHORIZATION": f"Token {self.user_token.key}"}

        rir = RIR.objects.create(name="RFC 6996", is_private=True)
        self.asns = [ASN(asn=65000 + i, rir=rir) for i in range(8)]
        ASN.objects.bulk_create(self.asns)

        self.sites = (
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

        devices = (
            Device(
                id=1,
                device_type=self.device_types[0],
                role=self.roles[0],
                name="Device 1",
                site=self.sites[0],
                rack=self.racks[0],
                cluster=self.clusters[0],
                local_context_data={"A": 1},
            ),
            Device(
                id=2,
                device_type=self.device_types[0],
                role=self.roles[0],
                name="Device 2",
                site=self.sites[0],
                rack=self.racks[0],
                cluster=self.clusters[0],
                local_context_data={"B": 2},
            ),
        )
        Device.objects.bulk_create(devices)

        self.url = "/api/plugins/diode/apply-change-set/"


class ApplyChangeSetTestCase(BaseApplyChangeSet):
    """ApplyChangeSet test cases."""

    def test_change_type_create_return_200(self):
        """Test create change_type with successful."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
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
            ],
        }

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get("result"), "success")

    def test_change_type_update_return_200(self):
        """Test update change_type with successful."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 2,
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

        site_updated = Site.objects.get(id=2)

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get("result"), "success")
        self.assertEqual(site_updated.name, "Site A")

    def test_change_type_create_with_error_return_400(self):
        """Test create change_type with wrong payload."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
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

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_created = Site.objects.filter(name="Site A")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(response.json().get("change_id"), "<UUID-0>")
        self.assertFalse(site_created.exists())

    def test_change_type_update_with_error_return_400(self):
        """Test update change_type with wrong payload."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 2,
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

        site_updated = Site.objects.get(id=2)

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(response.json().get("change_id"), "<UUID-0>")
        self.assertEqual(site_updated.name, "Site 2")

    def test_change_type_create_with_multiples_objects_return_200(self):
        """Test create change type with two objects."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
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
                    "change_id": "<UUID-1>",
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": None,
                    "data": {
                        "device_type": self.device_types[1].pk,
                        "role": self.roles[1].pk,
                        "name": "Test Device 4",
                        "site": self.sites[1].pk,
                        "rack": self.racks[1].pk,
                        "cluster": self.clusters[1].pk,
                    },
                },
            ],
        }

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get("result"), "success")

    def test_change_type_update_with_multiples_objects_return_200(self):
        """Test update change type with two objects."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 2,
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
                    "change_id": "<UUID-1>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": 1,
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

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_updated = Site.objects.get(id=2)
        devide_updated = Device.objects.get(id=1)

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get("result"), "success")
        self.assertEqual(site_updated.name, "Site A")
        self.assertEqual(devide_updated.name, "Test Device 3")

    def test_change_type_create_and_update_with_error_in_one_object_return_400(self):
        """Test create and update change type with one object with error."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
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
                    "change_id": "<UUID-1>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.device",
                    "object_id": 1,
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

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_created = Site.objects.filter(name="Site Z")
        device_created = Device.objects.filter(name="Test Device 4")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(response.json().get("change_id"), "<UUID-1>")
        self.assertFalse(site_created.exists())
        self.assertFalse(device_created.exists())

    def test_multiples_change_type_create_with_error_in_one_object_return_400(self):
        """Test create change_type with error in one object."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
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
                    "change_id": "<UUID-1>",
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
            ],
        }

        response = self.client.post(
            self.url, payload, format="json", **self.user_header
        )

        site_created = Site.objects.filter(name="Site Z")
        device_created = Device.objects.filter(name="Test Device 4")

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(response.json().get("change_id"), "<UUID-1>")
        self.assertFalse(site_created.exists())
        self.assertFalse(device_created.exists())

    def test_change_type_update_with_object_id_not_exist_return_400(self):
        """Test update object with nonexistent object_id."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
                    "change_type": "update",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 3,
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

        site_updated = Site.objects.get(id=2)
        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.json().get("result"), "failed")
        self.assertEqual(response.json().get("change_id"), "<UUID-0>")
        self.assertEqual(site_updated.name, "Site 2")
