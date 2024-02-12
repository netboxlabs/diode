#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs - Tests."""

from netboxlabs.diode.sdk import ClientConfiguration

def test_init():
    """Ensure we can initiate a client configuration."""
    config = ClientConfiguration(name="my-producer", version="0.0.1")
    assert config.name == "my-producer"
    assert config.version == "0.0.1"
