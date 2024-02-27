#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - API URLs."""

from django.urls import include, path
from netbox.api.routers import NetBoxRouter

from .views import ApplyChangeSetView, ObjectStateView

router = NetBoxRouter()
# router.register("apply-change-set", ApplyChangeSetView, basename="apply-change-set")

urlpatterns = [
    path("object-state/", ObjectStateView.as_view()),
    path("apply-change-set/", ApplyChangeSetView.as_view()),
    path("", include(router.urls)),
]
