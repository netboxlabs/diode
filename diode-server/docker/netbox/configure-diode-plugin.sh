#!/bin/bash

echo "🎛️ Configuring diode-netbox-plugin"

echo "PLUGINS = [\"netbox_diode_plugin\"]" > /etc/netbox/config/plugins.py

./manage.py configurediodeplugin || { echo "❌ enabling diode-netbox-plugin failed"; exit 1; }

echo "✅ diode-netbox-plugin configured successfully!"
