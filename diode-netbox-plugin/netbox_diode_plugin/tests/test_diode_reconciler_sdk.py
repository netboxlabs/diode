#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Diode Reconciler SDK - Tests."""
import pytest

from netbox_diode_plugin.diode_reconciler_sdk.client import DiodeReconcilerClient


def test_new_client():
    """Test version of the SDK."""
    sdk = DiodeReconcilerClient("localhost:50051", "foobar")
    assert sdk.name == "diode-reconciler-sdk-python"
    assert sdk.version == "v0.0.0-dev-unknown"
    assert sdk.target == "localhost:50051"
    assert sdk.channel is not None
