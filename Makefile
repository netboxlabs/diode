.PHONY: gen-diode-sdk-go
gen-diode-sdk-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.go.yaml
	@gofmt -w ../diode-sdk-go/diode

.PHONY: gen-diode-sdk-python
gen-diode-sdk-python:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.sdk.py.yaml --include-imports
	@find ../diode-sdk-python/netboxlabs/diode/sdk \( -name '*.py' -o -name '*.pyi' \) \
	-exec sed -i.bak -e 's/^from diode.v1/from netboxlabs.diode.sdk.diode.v1/' \
	-e 's/^from validate/from netboxlabs.diode.sdk.validate/' {} \; -exec rm -f {}.bak \;

.PHONY: gen-diode-server-go
gen-diode-server-go:
	@cd diode-proto/ && buf format -w && buf generate --template buf.gen.server.go.yaml
	@gofmt -w diode-server/gen

.PHONY: detect-breaking-changes
detect-breaking-changes:
	@cd diode-proto/ && buf breaking --against '../.git#subdir=diode-proto'

.PHONY: ruff-check
ruff-check:
	@echo "🔍 Running ruff check on tests directory..."
	@test -d tests/.venv || python3 -m venv tests/.venv
	@. tests/.venv/bin/activate && \
		pip install -q -r tests/requirements.txt && \
		ruff check tests/ --exclude "*_pb2*.py"
	@echo "✅ Ruff check completed"

.PHONY: ruff-format-check
ruff-format-check:
	@echo "🔍 Running ruff format check on tests directory..."
	@test -d tests/.venv || python3 -m venv tests/.venv
	@. tests/.venv/bin/activate && \
		pip install -q -r tests/requirements.txt && \
		ruff format --check tests/ --exclude "*_pb2*.py"
	@echo "✅ Ruff format check completed"

.PHONY: ruff-fix
ruff-fix:
	@echo "🔧 Running ruff fix on tests directory..."
	@test -d tests/.venv || python3 -m venv tests/.venv
	@. tests/.venv/bin/activate && \
		pip install -q -r tests/requirements.txt && \
		ruff check --fix tests/ --exclude "*_pb2*.py" && \
		ruff format tests/ --exclude "*_pb2*.py"
	@echo "✅ Ruff fix completed"