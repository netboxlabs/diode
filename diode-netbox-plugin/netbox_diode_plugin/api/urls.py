#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API URLs."""

from django.urls import include, path
from netbox.api.routers import NetBoxRouter

from .views import ObjectStateList

router = NetBoxRouter()

urlpatterns = [
    path("object-state/", ObjectStateList.as_view()),
    path("", include(router.urls)),
]
