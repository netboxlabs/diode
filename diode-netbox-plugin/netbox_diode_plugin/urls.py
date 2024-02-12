#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - URLs."""

from django.urls import path
from . import views

urlpatterns = (
    path("display-state/", views.DisplayStateView.as_view(), name="display_state"),

)
