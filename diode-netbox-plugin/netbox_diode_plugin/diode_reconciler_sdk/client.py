#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Diode Reconciler SDK."""

import grpc

from netbox_diode_plugin.diode_reconciler_sdk.reconciler.v1 import reconciler_pb2, reconciler_pb2_grpc


class DiodeReconcilerClient:
    """Diode Reconciler client."""

    _name = None
    _version = None
    _target = None
    _channel = None
    _stub = None

    def __init__(self, name: str, version: str, target: str, api_key: str) -> None:
        """Initiate a new client configuration."""
        self._name = name
        self._version = version
        # TODO(mfiedorowicz): configure secure channel with auth metatada callback
        self._auth_metadata = (("diode-api-key", api_key),)
        self._channel = grpc.insecure_channel(target)
        self._stub = reconciler_pb2_grpc.ReconcilerServiceStub(self._channel)

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
    def channel(self) -> grpc.Channel:
        """Retrieve the channel."""
        return self._channel

    def add_object_state(self, object_id: int, object_type: str,
                         object_change_id: int, object: dict) -> reconciler_pb2.AddObjectStateResponse:
        """Add an object state."""
        request = reconciler_pb2.AddObjectStateRequest(object_id=object_id, object_type=object_type,
                                                       object_change_id=object_change_id, object=object,
                                                       sdk_name=self.name, sdk_version=self.version)

        return self._stub.AddObjectState(request, metadata=self._auth_metadata)
