#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""NetBox Labs, Diode - Client configuration."""


class ClientConfiguration:
    """Diode client configuration."""

    _name = None
    _version = None

    def __init__(self, name: str, version: str):
        """Initiate a new client configuration."""
        self._name = name
        self._version = version
        # TODO: ensure version is a valid semver
        # TODO: obtain meta data about the environment; Python version, CPU arch, OS

    @property
    def name(self) -> str:
        """Retrieve the name."""
        return self._name

    @property
    def version(self) -> str:
        """Retrieve the version."""
        return self._version
