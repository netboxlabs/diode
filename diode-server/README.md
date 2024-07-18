# Diode servers

Diode server is splited into two services:

### Ingester Service

- Responsible for receiving and validating ingestion data.
- Utilizes `IngesterService.Ingest` RPC method.
- Supports multiple API keys for data source authorization.
- Validates incoming data and pushes it into Redis streams.

### Reconciler Service

- Processes data from Redis streams and converts it for storage.
- Manages data sources and their API keys.
- Implements a reconciliation engine to detect and store deltas between ingested data and the current NetBox object state.
