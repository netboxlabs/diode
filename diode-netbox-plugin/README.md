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

## Running Tests

```bash
make docker-compose-netbox-plugin-test
```