#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Diode Reconciler SDK - Tests."""
import pytest

from netbox_diode_plugin.diode_reconciler_sdk.client import DiodeReconcilerClient


def test_new_client():
    """Test version of the SDK."""
    sdk = DiodeReconcilerClient("test", "0.0.0", "localhost:50051")
    assert sdk.version == "0.0.0"
    assert sdk.name == "test"
    assert sdk.target == "localhost:50051"
