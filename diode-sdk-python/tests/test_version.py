#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs - Tests."""

from netboxlabs.diode.sdk.version import version_semver


def test_version():
    """Check the injected semver."""
    assert version_semver() == "0.0.0"
