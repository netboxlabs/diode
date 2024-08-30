.PHONY: gen-diode-sdk-go
gen-diode-sdk-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.go.yaml

.PHONY: gen-diode-sdk-python
gen-diode-sdk-python:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.py.yaml --include-imports
	@find ../diode-sdk-python/netboxlabs/diode/sdk \( -name '*.py' -o -name '*.pyi' \) \
	-exec sed -i '' 's/^from diode.v1/from netboxlabs.diode.sdk.diode.v1/; s/^from validate/from netboxlabs.diode.sdk.validate/' {} \;

.PHONY: gen-diode-server-go
gen-diode-server-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.server.go.yaml
