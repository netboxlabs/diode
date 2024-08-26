.PHONY: gen-diode-sdk-go gen-diode-sdk-python

gen-diode-sdk-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.go.yaml

gen-diode-go-internal:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.go-internal.yaml

gen-diode-sdk-python:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.py.yaml --include-imports
	@find ../diode-sdk-python/netboxlabs/diode/sdk \( -name '*.py' -o -name '*.pyi' \) \
	-exec sed -i.bak -e 's/^from diode.v1/from netboxlabs.diode.sdk.diode.v1/' \
	-e 's/^from validate/from netboxlabs.diode.sdk.validate/' {} \; -exec rm -f {}.bak \;