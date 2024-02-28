#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ApplyChangeSetView."""

from dcim.models import Site
from django.contrib.auth import get_user_model
from ipam.models import ASN, RIR
from users.models import Token
from utilities.testing import APITestCase

User = get_user_model()


class ApplyChangeSetTestCase(APITestCase):
    """ApplyChangeSet test cases."""

    def setUp(self):
        """Set up test."""
        self.user = User.objects.create_user(username="testcommonuser")
        self.add_permissions("netbox_diode_plugin.add_objectstate")
        self.user_token = Token.objects.create(user=self.user)

        self.user_header = {"HTTP_AUTHORIZATION": f"Token {self.user_token.key}"}

        rir = RIR.objects.create(name="RFC 6996", is_private=True)
        self.asns = [ASN(asn=65000 + i, rir=rir) for i in range(8)]
        ASN.objects.bulk_create(self.asns)

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
        )
        Site.objects.bulk_create(sites)

        self.url = "/api/plugins/diode/apply-change-set/"

    def test_apply_change_set_create_return_200(self):
        """Test apply change set to create."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-0>",
                    "change_type": "create",
                    "object_version": None,
                    "object_type": "dcim.site",
                    "object_id": 1,
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

    def test_apply_change_set_update_return_200(self):
        """Test apply change set to update."""
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

    def test_apply_change_set_update_with_error_return_400(self):
        """Test apply change set to update."""
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
        self.assertEqual(site_updated.name, "Site 2")
