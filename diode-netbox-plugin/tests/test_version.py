#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs - Tests."""

from netbox_diode_plugin.version import version_semver


def test_version():
    """Check the injected semver."""
    assert version_semver() == "0.0.0"
