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

## Running Tests

a) Start the container in diode/diode-server:
```bash
make docker-compose-up
```

b) Enter in the Diode-Netbox container:

```bash
docker exec -it diode-netbox-1 /bin/bash
```

c) Execute the tests:
```bash
./manage.py test --keepdb netbox_diode_plugin.tests.test_object_state.ObjectStateListTestCase
```

