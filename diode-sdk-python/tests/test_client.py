#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs - Tests."""

import grpc
import pytest

from netboxlabs.diode.sdk import DiodeClient
from netboxlabs.diode.sdk.exceptions import DiodeClientError, DiodeConfigError


def test_init():
    """Ensure we can initiate a client configuration."""
    config = DiodeClient(
        target="localhost:8081",
        app_name="my-producer",
        app_version="0.0.1",
        api_key="abcde",
    )
    assert config.target == "localhost:8081"
    assert config.name == "diode-sdk-python"
    assert config.version == "0.0.1"
    assert config.app_name == "my-producer"
    assert config.app_version == "0.0.1"


def test_config_error():
    """Ensure we can raise a config error."""
    with pytest.raises(DiodeConfigError) as err:
        DiodeClient(
            target="localhost:8081", app_name="my-producer", app_version="0.0.1"
        )
    assert (
        str(err.value) == "api_key param or DIODE_API_KEY environment variable required"
    )


def test_client_error():
    """Ensure we can raise a client error."""
    with pytest.raises(DiodeClientError) as err:
        client = DiodeClient(
            target="invalid:8081",
            app_name="my-producer",
            app_version="0.0.1",
            api_key="abcde",
        )
        client.ingest(entities=[])
    assert err.value.status_code == grpc.StatusCode.UNAVAILABLE
