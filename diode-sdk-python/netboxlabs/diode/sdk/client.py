#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs, Diode - SDK - Client."""

import logging
import os
import uuid
from typing import Iterable, Optional

import grpc

from netboxlabs.diode.sdk.diode.v1 import ingester_pb2, ingester_pb2_grpc
from netboxlabs.diode.sdk.exceptions import DiodeClientError, DiodeConfigError

_DIODE_API_KEY_ENVVAR_NAME = "DIODE_API_KEY"
_DIODE_SDK_LOG_LEVEL_ENVVAR_NAME = "DIODE_SDK_LOG_LEVEL"
_DEFAULT_STREAM = "latest"
_LOGGER = logging.getLogger(__name__)


class DiodeClient:
    """Diode Client."""

    _name = "diode-sdk-python"
    _version = "0.0.1"
    _app_name = None
    _app_version = None
    _channel = None
    _stub = None

    def __init__(
        self,
        target: str,
        app_name: str,
        app_version: str,
        api_key: Optional[str] = None,
    ):
        """Initiate a new client."""
        log_level = os.getenv(_DIODE_SDK_LOG_LEVEL_ENVVAR_NAME, "INFO").upper()
        logging.basicConfig(level=log_level)

        # TODO: validate target
        self._target = target

        self._app_name = app_name
        self._app_version = app_version

        if api_key is None:
            api_key = os.getenv(_DIODE_API_KEY_ENVVAR_NAME)
        if api_key is None:
            raise DiodeConfigError("API key is required")

        self._auth_metadata = (("diode-api-key", api_key),)
        # TODO: add support for secure channel (TLS verify flag and cert)
        self._channel = grpc.insecure_channel(target)
        self._stub = ingester_pb2_grpc.IngesterServiceStub(self._channel)
        # TODO: obtain meta data about the environment; Python version, CPU arch, OS

    @property
    def name(self) -> str:
        """Retrieve the name."""
        return self._name

    @property
    def version(self) -> str:
        """Retrieve the version."""
        return self._version

    @property
    def target(self) -> str:
        """Retrieve the target."""
        return self._target

    @property
    def app_name(self) -> str:
        """Retrieve the app name."""
        return self._app_name

    @property
    def app_version(self) -> str:
        """Retrieve the app version."""
        return self._app_version

    @property
    def channel(self) -> grpc.Channel:
        """Retrieve the channel."""
        return self._channel

    def __enter__(self):
        """Enters the runtime context related to the channel object."""
        return self

    def __exit__(self, exc_type, exc_value, exc_traceback):
        """Exits the runtime context related to the channel object."""
        self.close()

    def close(self):
        """Close the channel."""
        self._channel.close()

    def ingest(
        self,
        entities: Iterable[ingester_pb2.Entity],
        stream: Optional[str] = _DEFAULT_STREAM,
    ) -> ingester_pb2.IngestResponse:
        """Push a message."""
        try:
            request = ingester_pb2.IngestRequest(
                stream=stream,
                id=str(uuid.uuid4()),
                entities=entities,
                sdk_name=self.name,
                sdk_version=self.version,
                producer_app_name=self.app_name,
                producer_app_version=self.app_version,
            )

            return self._stub.Ingest(request, metadata=self._auth_metadata)
        except grpc.RpcError as err:
            raise DiodeClientError(err) from err
