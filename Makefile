.PHONY: gen-diode-sdk-go
gen-diode-sdk-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.go.yaml

.PHONY: gen-diode-sdk-python
gen-diode-sdk-python:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.py.yaml --include-imports
	@find ../diode-sdk-python/netboxlabs/diode/sdk \( -name '*.py' -o -name '*.pyi' \) \
	-exec sed -i.bak -e 's/^from diode.v1/from netboxlabs.diode.sdk.diode.v1/' \
	-e 's/^from validate/from netboxlabs.diode.sdk.validate/' {} \; -exec rm -f {}.bak \;

.PHONY: gen-diode-server-go
gen-diode-server-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.server.go.yaml

.PHONY: gen-diode-netbox-plugin-reconciler-sdk-python
gen-diode-netbox-plugin-reconciler-sdk-python:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.netbox-plugin.reconciler.sdk.py.yaml --include-imports
	@find ../diode-netbox-plugin/netbox_diode_plugin/reconciler/sdk/diode/v1 \( -name '*.py' -o -name '*.pyi' \) -execdir mv {} ../../v1/ \;
	@rm -rf ../diode-netbox-plugin/netbox_diode_plugin/reconciler/sdk/diode
	@find ../diode-netbox-plugin/netbox_diode_plugin/reconciler/sdk/ \( -name '*.py' -o -name '*.pyi' \) \
	-exec sed -i.bak -e 's/^from diode.v1/from netbox_diode_plugin.reconciler.sdk.v1/' \
	-e 's/^from validate/from netbox_diode_plugin.reconciler.sdk.validate/' {} \; -exec rm -f {}.bak \;