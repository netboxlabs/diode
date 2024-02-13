# Diode NetBox Plugin

## Installation

```bash
pip install netboxlabs-diode-netbox-plugin
```

In your `configuration.py` file, add `netbox_diode_plugin` to the `PLUGINS` list.

```python
PLUGINS = [
    "netbox_diode_plugin",
]
```

Create a plugin configuration entry:

```python
PLUGINS_CONFIG = {
    "netbox_diode_plugin": {
        
    }
}
```

