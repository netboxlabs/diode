#!/usr/bin/env python
# Copyright 2024 NetBox Labs Inc
"""Diode Netbox Plugin - Views."""

from django.shortcuts import redirect, render
from django.views.generic import View
from django_tables2 import SingleTableView


class DisplayStateView(View):
    def get(self, request):
        return render(request, "netbox_diode_plugin/display_state.html")

