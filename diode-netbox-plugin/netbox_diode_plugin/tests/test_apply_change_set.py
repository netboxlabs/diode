#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Tests for ApplyChangeSetView."""

from utilities.testing import APITestCase
from ipam.models import ASN, RIR


class ApplyChangeSetTestCase(APITestCase):
    """ApplyChangeSet test cases."""

    def setUp(self):

        rir = RIR.objects.create(name="RFC 6996", is_private=True)
        self.asns = [ASN(asn=65000 + i, rir=rir) for i in range(8)]
        ASN.objects.bulk_create(self.asns)

        self.url = "/api/plugins/diode/apply-change-set/"

    def test_apply_change_set(self):
        """Test apply change set."""
        payload = {
            "change_set_id": "<UUID-0>",
            "change_set": [
                {
                    "change_id": "<UUID-1>",
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

        response = self.client.post(self.url, payload, format="json")
        print(response.content)
        self.assertEqual(response.status_code, 200)
        pass
