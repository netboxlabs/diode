# Protocol Documentation

<a name="top"></a>

## Table of Contents

- [diode/v1/ingester.proto](#diode_v1_ingester-proto)
    - [Cluster](#diode-v1-Cluster)
    - [ClusterGroup](#diode-v1-ClusterGroup)
    - [ClusterType](#diode-v1-ClusterType)
    - [Device](#diode-v1-Device)
    - [DeviceType](#diode-v1-DeviceType)
    - [Entity](#diode-v1-Entity)
    - [IPAddress](#diode-v1-IPAddress)
    - [IngestRequest](#diode-v1-IngestRequest)
    - [IngestResponse](#diode-v1-IngestResponse)
    - [Interface](#diode-v1-Interface)
    - [Manufacturer](#diode-v1-Manufacturer)
    - [Platform](#diode-v1-Platform)
    - [Prefix](#diode-v1-Prefix)
    - [Role](#diode-v1-Role)
    - [Site](#diode-v1-Site)
    - [Tag](#diode-v1-Tag)
    - [VMInterface](#diode-v1-VMInterface)
    - [VirtualDisk](#diode-v1-VirtualDisk)
    - [VirtualMachine](#diode-v1-VirtualMachine)

    - [IngesterService](#diode-v1-IngesterService)

- [diode/v1/reconciler.proto](#diode_v1_reconciler-proto)
    - [Change](#diode-v1-Change)
    - [ChangeSet](#diode-v1-ChangeSet)
    - [Deviation](#diode-v1-Deviation)
    - [DeviationError](#diode-v1-DeviationError)
    - [IngestionError](#diode-v1-IngestionError)
    - [IngestionError.Details](#diode-v1-IngestionError-Details)
    - [IngestionError.Details.Error](#diode-v1-IngestionError-Details-Error)
    - [IngestionLog](#diode-v1-IngestionLog)
    - [IngestionMetrics](#diode-v1-IngestionMetrics)
    - [RetrieveDeviationByIDRequest](#diode-v1-RetrieveDeviationByIDRequest)
    - [RetrieveDeviationByIDResponse](#diode-v1-RetrieveDeviationByIDResponse)
    - [RetrieveDeviationsRequest](#diode-v1-RetrieveDeviationsRequest)
    - [RetrieveDeviationsResponse](#diode-v1-RetrieveDeviationsResponse)
    - [RetrieveIngestionLogsRequest](#diode-v1-RetrieveIngestionLogsRequest)
    - [RetrieveIngestionLogsResponse](#diode-v1-RetrieveIngestionLogsResponse)

    - [State](#diode-v1-State)

    - [ReconcilerService](#diode-v1-ReconcilerService)

- [Scalar Value Types](#scalar-value-types)

<a name="diode_v1_ingester-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/ingester.proto

<a name="diode-v1-Cluster"></a>

### Cluster

A Cluster

| Field       | Type                                   | Label    | Description |
|-------------|----------------------------------------|----------|-------------|
| name        | [string](#string)                      |          |             |
| type        | [ClusterType](#diode-v1-ClusterType)   |          |             |
| group       | [ClusterGroup](#diode-v1-ClusterGroup) |          |             |
| site        | [Site](#diode-v1-Site)                 |          |             |
| status      | [string](#string)                      |          |             |
| description | [string](#string)                      | optional |             |
| tags        | [Tag](#diode-v1-Tag)                   | repeated |             |

<a name="diode-v1-ClusterGroup"></a>

### ClusterGroup

A Cluster Group

| Field       | Type                 | Label    | Description |
|-------------|----------------------|----------|-------------|
| name        | [string](#string)    |          |             |
| slug        | [string](#string)    |          |             |
| description | [string](#string)    | optional |             |
| tags        | [Tag](#diode-v1-Tag) | repeated |             |

<a name="diode-v1-ClusterType"></a>

### ClusterType

A Cluster Type

| Field       | Type                 | Label    | Description |
|-------------|----------------------|----------|-------------|
| name        | [string](#string)    |          |             |
| slug        | [string](#string)    |          |             |
| description | [string](#string)    | optional |             |
| tags        | [Tag](#diode-v1-Tag) | repeated |             |

<a name="diode-v1-Device"></a>

### Device

A device

| Field       | Type                               | Label    | Description |
|-------------|------------------------------------|----------|-------------|
| name        | [string](#string)                  |          |             |
| device_fqdn | [string](#string)                  | optional |             |
| device_type | [DeviceType](#diode-v1-DeviceType) |          |             |
| role        | [Role](#diode-v1-Role)             |          |             |
| platform    | [Platform](#diode-v1-Platform)     |          |             |
| serial      | [string](#string)                  | optional |             |
| site        | [Site](#diode-v1-Site)             |          |             |
| asset_tag   | [string](#string)                  | optional |             |
| status      | [string](#string)                  |          |             |
| description | [string](#string)                  | optional |             |
| comments    | [string](#string)                  | optional |             |
| tags        | [Tag](#diode-v1-Tag)               | repeated |             |
| primary_ip4 | [IPAddress](#diode-v1-IPAddress)   |          |             |
| primary_ip6 | [IPAddress](#diode-v1-IPAddress)   |          |             |

<a name="diode-v1-DeviceType"></a>

### DeviceType

A device type

| Field        | Type                                   | Label    | Description |
|--------------|----------------------------------------|----------|-------------|
| model        | [string](#string)                      |          |             |
| slug         | [string](#string)                      |          |             |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |          |             |
| description  | [string](#string)                      | optional |             |
| comments     | [string](#string)                      | optional |             |
| part_number  | [string](#string)                      | optional |             |
| tags         | [Tag](#diode-v1-Tag)                   | repeated |             |

<a name="diode-v1-Entity"></a>

### Entity

An ingest entity wrapper

| Field           | Type                                                    | Label | Description                                   |
|-----------------|---------------------------------------------------------|-------|-----------------------------------------------|
| site            | [Site](#diode-v1-Site)                                  |       |                                               |
| platform        | [Platform](#diode-v1-Platform)                          |       |                                               |
| manufacturer    | [Manufacturer](#diode-v1-Manufacturer)                  |       |                                               |
| device          | [Device](#diode-v1-Device)                              |       |                                               |
| device_role     | [Role](#diode-v1-Role)                                  |       |                                               |
| device_type     | [DeviceType](#diode-v1-DeviceType)                      |       |                                               |
| interface       | [Interface](#diode-v1-Interface)                        |       |                                               |
| ip_address      | [IPAddress](#diode-v1-IPAddress)                        |       |                                               |
| prefix          | [Prefix](#diode-v1-Prefix)                              |       |                                               |
| cluster_group   | [ClusterGroup](#diode-v1-ClusterGroup)                  |       |                                               |
| cluster_type    | [ClusterType](#diode-v1-ClusterType)                    |       |                                               |
| cluster         | [Cluster](#diode-v1-Cluster)                            |       |                                               |
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine)              |       |                                               |
| vminterface     | [VMInterface](#diode-v1-VMInterface)                    |       |                                               |
| virtual_disk    | [VirtualDisk](#diode-v1-VirtualDisk)                    |       |                                               |
| timestamp       | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |       | The timestamp of the data discovery at source |

<a name="diode-v1-IPAddress"></a>

### IPAddress

An IP address.

| Field       | Type                             | Label    | Description |
|-------------|----------------------------------|----------|-------------|
| address     | [string](#string)                |          |             |
| interface   | [Interface](#diode-v1-Interface) |          |             |
| status      | [string](#string)                |          |             |
| role        | [string](#string)                |          |             |
| dns_name    | [string](#string)                | optional |             |
| description | [string](#string)                | optional |             |
| comments    | [string](#string)                | optional |             |
| tags        | [Tag](#diode-v1-Tag)             | repeated |             |

<a name="diode-v1-IngestRequest"></a>

### IngestRequest

The request to ingest the data

| Field                | Type                       | Label    | Description |
|----------------------|----------------------------|----------|-------------|
| stream               | [string](#string)          |          |             |
| entities             | [Entity](#diode-v1-Entity) | repeated |             |
| id                   | [string](#string)          |          |             |
| producer_app_name    | [string](#string)          |          |             |
| producer_app_version | [string](#string)          |          |             |
| sdk_name             | [string](#string)          |          |             |
| sdk_version          | [string](#string)          |          |             |

<a name="diode-v1-IngestResponse"></a>

### IngestResponse

The response from the ingest request

| Field  | Type              | Label    | Description |
|--------|-------------------|----------|-------------|
| errors | [string](#string) | repeated |             |

<a name="diode-v1-Interface"></a>

### Interface

An interface

| Field          | Type                       | Label    | Description |
|----------------|----------------------------|----------|-------------|
| device         | [Device](#diode-v1-Device) |          |             |
| name           | [string](#string)          |          |             |
| label          | [string](#string)          | optional |             |
| type           | [string](#string)          |          |             |
| enabled        | [bool](#bool)              | optional |             |
| mtu            | [int32](#int32)            | optional |             |
| mac_address    | [string](#string)          | optional |             |
| speed          | [int32](#int32)            | optional |             |
| wwn            | [string](#string)          | optional |             |
| mgmt_only      | [bool](#bool)              | optional |             |
| description    | [string](#string)          | optional |             |
| mark_connected | [bool](#bool)              | optional |             |
| mode           | [string](#string)          |          |             |
| tags           | [Tag](#diode-v1-Tag)       | repeated |             |

<a name="diode-v1-Manufacturer"></a>

### Manufacturer

A manufacturer

| Field       | Type                 | Label    | Description |
|-------------|----------------------|----------|-------------|
| name        | [string](#string)    |          |             |
| slug        | [string](#string)    |          |             |
| description | [string](#string)    | optional |             |
| tags        | [Tag](#diode-v1-Tag) | repeated |             |

<a name="diode-v1-Platform"></a>

### Platform

A platform

| Field        | Type                                   | Label    | Description |
|--------------|----------------------------------------|----------|-------------|
| name         | [string](#string)                      |          |             |
| slug         | [string](#string)                      |          |             |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |          |             |
| description  | [string](#string)                      | optional |             |
| tags         | [Tag](#diode-v1-Tag)                   | repeated |             |

<a name="diode-v1-Prefix"></a>

### Prefix

An IPAM prefix.

| Field         | Type                   | Label    | Description |
|---------------|------------------------|----------|-------------|
| prefix        | [string](#string)      |          |             |
| site          | [Site](#diode-v1-Site) |          |             |
| status        | [string](#string)      |          |             |
| is_pool       | [bool](#bool)          | optional |             |
| mark_utilized | [bool](#bool)          | optional |             |
| description   | [string](#string)      | optional |             |
| comments      | [string](#string)      | optional |             |
| tags          | [Tag](#diode-v1-Tag)   | repeated |             |

<a name="diode-v1-Role"></a>

### Role

A role

| Field       | Type                 | Label    | Description |
|-------------|----------------------|----------|-------------|
| name        | [string](#string)    |          |             |
| slug        | [string](#string)    |          |             |
| color       | [string](#string)    |          |             |
| description | [string](#string)    | optional |             |
| tags        | [Tag](#diode-v1-Tag) | repeated |             |

<a name="diode-v1-Site"></a>

### Site

A site

| Field       | Type                 | Label    | Description |
|-------------|----------------------|----------|-------------|
| name        | [string](#string)    |          |             |
| slug        | [string](#string)    |          |             |
| status      | [string](#string)    |          |             |
| facility    | [string](#string)    | optional |             |
| time_zone   | [string](#string)    | optional |             |
| description | [string](#string)    | optional |             |
| comments    | [string](#string)    | optional |             |
| tags        | [Tag](#diode-v1-Tag) | repeated |             |

<a name="diode-v1-Tag"></a>

### Tag

A tag

| Field | Type              | Label | Description |
|-------|-------------------|-------|-------------|
| name  | [string](#string) |       |             |
| slug  | [string](#string) |       |             |
| color | [string](#string) |       |             |

<a name="diode-v1-VMInterface"></a>

### VMInterface

A Virtual Machine Interface

| Field           | Type                                       | Label    | Description |
|-----------------|--------------------------------------------|----------|-------------|
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |          |             |
| name            | [string](#string)                          |          |             |
| enabled         | [bool](#bool)                              | optional |             |
| mtu             | [int32](#int32)                            | optional |             |
| mac_address     | [string](#string)                          | optional |             |
| description     | [string](#string)                          | optional |             |
| tags            | [Tag](#diode-v1-Tag)                       | repeated |             |

<a name="diode-v1-VirtualDisk"></a>

### VirtualDisk

A Virtual Disk

| Field           | Type                                       | Label    | Description |
|-----------------|--------------------------------------------|----------|-------------|
| virtual_machine | [VirtualMachine](#diode-v1-VirtualMachine) |          |             |
| name            | [string](#string)                          |          |             |
| size            | [int32](#int32)                            |          |             |
| description     | [string](#string)                          | optional |             |
| tags            | [Tag](#diode-v1-Tag)                       | repeated |             |

<a name="diode-v1-VirtualMachine"></a>

### VirtualMachine

A Virtual Machine

| Field       | Type                             | Label    | Description |
|-------------|----------------------------------|----------|-------------|
| name        | [string](#string)                |          |             |
| status      | [string](#string)                |          |             |
| site        | [Site](#diode-v1-Site)           |          |             |
| cluster     | [Cluster](#diode-v1-Cluster)     |          |             |
| role        | [Role](#diode-v1-Role)           |          |             |
| device      | [Device](#diode-v1-Device)       |          |             |
| platform    | [Platform](#diode-v1-Platform)   |          |             |
| primary_ip4 | [IPAddress](#diode-v1-IPAddress) |          |             |
| primary_ip6 | [IPAddress](#diode-v1-IPAddress) |          |             |
| vcpus       | [int32](#int32)                  | optional |             |
| memory      | [int32](#int32)                  | optional |             |
| disk        | [int32](#int32)                  | optional |             |
| description | [string](#string)                | optional |             |
| comments    | [string](#string)                | optional |             |
| tags        | [Tag](#diode-v1-Tag)             | repeated |             |

<a name="diode-v1-IngesterService"></a>

### IngesterService

Ingestion API

| Method Name | Request Type                             | Response Type                              | Description                  |
|-------------|------------------------------------------|--------------------------------------------|------------------------------|
| Ingest      | [IngestRequest](#diode-v1-IngestRequest) | [IngestResponse](#diode-v1-IngestResponse) | Ingests data into the system |

<a name="diode_v1_reconciler-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/reconciler.proto

<a name="diode-v1-Change"></a>

### Change

A change

| Field                | Type              | Label | Description |
|----------------------|-------------------|-------|-------------|
| id                   | [string](#string) |       |             |
| object_type          | [string](#string) |       |             |
| object_primary_value | [string](#string) |       |             |
| change_type          | [string](#string) |       |             |
| before               | [bytes](#bytes)   |       |             |
| after                | [bytes](#bytes)   |       |             |

<a name="diode-v1-ChangeSet"></a>

### ChangeSet

A change set

| Field          | Type              | Label    | Description                                          |
|----------------|-------------------|----------|------------------------------------------------------|
| id             | [string](#string) |          | A change set ID                                      |
| data           | [bytes](#bytes)   |          | Binary data representing the change set              |
| branch_id      | [string](#string) | optional | branch ID against which the change set was generated |
| deviation_name | [string](#string) | optional | deviation name                                       |

<a name="diode-v1-Deviation"></a>

### Deviation

A deviation

| Field           | Type                                       | Label    | Description |
|-----------------|--------------------------------------------|----------|-------------|
| id              | [string](#string)                          |          |             |
| ingestion_ts    | [int64](#int64)                            |          |             |
| last_update_ts  | [int64](#int64)                            |          |             |
| name            | [string](#string)                          |          |             |
| source          | [string](#string)                          |          |             |
| state           | [State](#diode-v1-State)                   |          |             |
| object_type     | [string](#string)                          |          |             |
| branch_id       | [string](#string)                          | optional |             |
| ingested_entity | [Entity](#diode-v1-Entity)                 |          |             |
| error           | [DeviationError](#diode-v1-DeviationError) |          |             |
| changes         | [Change](#diode-v1-Change)                 | repeated |             |

<a name="diode-v1-DeviationError"></a>

### DeviationError

Deviation error

| Field   | Type              | Label | Description |
|---------|-------------------|-------|-------------|
| message | [string](#string) |       |             |
| code    | [int32](#int32)   |       |             |

<a name="diode-v1-IngestionError"></a>

### IngestionError

IngestionError represents an error occurring while processing an ingestion entity

| Field   | Type                                                       | Label | Description |
|---------|------------------------------------------------------------|-------|-------------|
| message | [string](#string)                                          |       |             |
| code    | [int32](#int32)                                            |       |             |
| details | [IngestionError.Details](#diode-v1-IngestionError-Details) |       |             |

<a name="diode-v1-IngestionError-Details"></a>

### IngestionError.Details

| Field         | Type                                                                   | Label    | Description |
|---------------|------------------------------------------------------------------------|----------|-------------|
| change_set_id | [string](#string)                                                      |          |             |
| result        | [string](#string)                                                      |          |             |
| errors        | [IngestionError.Details.Error](#diode-v1-IngestionError-Details-Error) | repeated |             |

<a name="diode-v1-IngestionError-Details-Error"></a>

### IngestionError.Details.Error

| Field     | Type              | Label | Description                 |
|-----------|-------------------|-------|-----------------------------|
| error     | [string](#string) |       | key value pair of the error |
| change_id | [string](#string) |       |                             |

<a name="diode-v1-IngestionLog"></a>

### IngestionLog

An ingestion log

| Field                | Type                                       | Label | Description     |
|----------------------|--------------------------------------------|-------|-----------------|
| id                   | [string](#string)                          |       |                 |
| data_type            | [string](#string)                          |       | **Deprecated.** |
| state                | [State](#diode-v1-State)                   |       |                 |
| request_id           | [string](#string)                          |       |                 |
| ingestion_ts         | [int64](#int64)                            |       |                 |
| producer_app_name    | [string](#string)                          |       |                 |
| producer_app_version | [string](#string)                          |       |                 |
| sdk_name             | [string](#string)                          |       |                 |
| sdk_version          | [string](#string)                          |       |                 |
| entity               | [Entity](#diode-v1-Entity)                 |       |                 |
| error                | [IngestionError](#diode-v1-IngestionError) |       |                 |
| change_set           | [ChangeSet](#diode-v1-ChangeSet)           |       |                 |
| object_type          | [string](#string)                          |       |                 |

<a name="diode-v1-IngestionMetrics"></a>

### IngestionMetrics

Ingestion metrics

| Field      | Type            | Label | Description |
|------------|-----------------|-------|-------------|
| total      | [int32](#int32) |       |             |
| queued     | [int32](#int32) |       |             |
| reconciled | [int32](#int32) |       |             |
| failed     | [int32](#int32) |       |             |
| no_changes | [int32](#int32) |       |             |

<a name="diode-v1-RetrieveDeviationByIDRequest"></a>

### RetrieveDeviationByIDRequest

The request to retrieve deviation by ID

| Field | Type              | Label | Description  |
|-------|-------------------|-------|--------------|
| id    | [string](#string) |       | Deviation ID |

<a name="diode-v1-RetrieveDeviationByIDResponse"></a>

### RetrieveDeviationByIDResponse

The response from the retrieve deviation by ID request

| Field     | Type                             | Label | Description |
|-----------|----------------------------------|-------|-------------|
| deviation | [Deviation](#diode-v1-Deviation) |       | Deviation   |

<a name="diode-v1-RetrieveDeviationsRequest"></a>

### RetrieveDeviationsRequest

The request to retrieve deviations

| Field              | Type                     | Label    | Description                                   |
|--------------------|--------------------------|----------|-----------------------------------------------|
| page_size          | [int32](#int32)          | optional | Number of deviations per page, default is 100 |
| page_token         | [string](#string)        |          | Token to fetch the next page of results       |
| ingestion_ts_start | [int64](#int64)          |          | Optional start of ingestion timestamp range   |
| ingestion_ts_end   | [int64](#int64)          |          | Optional end of ingestion timestamp range     |
| state              | [State](#diode-v1-State) | repeated | Optional filter by states                     |
| object_type        | [string](#string)        | repeated | Optional filter by object types               |
| branch_id          | [string](#string)        | repeated | Optional filter by branch IDs                 |

<a name="diode-v1-RetrieveDeviationsResponse"></a>

### RetrieveDeviationsResponse

The response from the retrieve deviations request

| Field           | Type                             | Label    | Description                                |
|-----------------|----------------------------------|----------|--------------------------------------------|
| deviations      | [Deviation](#diode-v1-Deviation) | repeated | List of deviations                         |
| next_page_token | [string](#string)                |          | Token for the next page of results, if any |

<a name="diode-v1-RetrieveIngestionLogsRequest"></a>

### RetrieveIngestionLogsRequest

The request to retrieve ingestion logs

| Field              | Type                     | Label    | Description                                        |
|--------------------|--------------------------|----------|----------------------------------------------------|
| page_size          | [int32](#int32)          | optional | Number of logs per page, default is 100            |
| state              | [State](#diode-v1-State) | optional | Optional filter by state field                     |
| data_type          | [string](#string)        |          | **Deprecated.** Optional filter by data type field |
| request_id         | [string](#string)        |          | Optional filter by request ID                      |
| ingestion_ts_start | [int64](#int64)          |          | Optional start of ingestion timestamp range        |
| ingestion_ts_end   | [int64](#int64)          |          | Optional end of ingestion timestamp range          |
| page_token         | [string](#string)        |          | Token to fetch the next page of results            |
| only_metrics       | [bool](#bool)            |          | Flag to return only the ingestion metrics          |
| object_type        | [string](#string)        |          | Optional filter by object type                     |

<a name="diode-v1-RetrieveIngestionLogsResponse"></a>

### RetrieveIngestionLogsResponse

The response from the retrieve ingestion logs request

| Field           | Type                                           | Label    | Description                                |
|-----------------|------------------------------------------------|----------|--------------------------------------------|
| logs            | [IngestionLog](#diode-v1-IngestionLog)         | repeated | List of ingestion logs                     |
| metrics         | [IngestionMetrics](#diode-v1-IngestionMetrics) |          | ingestion metrics                          |
| next_page_token | [string](#string)                              |          | Token for the next page of results, if any |

<a name="diode-v1-State"></a>

### State

| Name              | Number | Description |
|-------------------|--------|-------------|
| STATE_UNSPECIFIED | 0      |             |
| QUEUED            | 1      |             |
| OPEN              | 2      |             |
| APPLIED           | 3      |             |
| FAILED            | 4      |             |
| NO_CHANGES        | 5      |             |
| IGNORED           | 6      |             |
| ERRORED           | 7      |             |

<a name="diode-v1-ReconcilerService"></a>

### ReconcilerService

Reconciler service API

| Method Name           | Request Type                                                           | Response Type                                                            | Description              |
|-----------------------|------------------------------------------------------------------------|--------------------------------------------------------------------------|--------------------------|
| RetrieveIngestionLogs | [RetrieveIngestionLogsRequest](#diode-v1-RetrieveIngestionLogsRequest) | [RetrieveIngestionLogsResponse](#diode-v1-RetrieveIngestionLogsResponse) | Retrieves ingestion logs |
| RetrieveDeviations    | [RetrieveDeviationsRequest](#diode-v1-RetrieveDeviationsRequest)       | [RetrieveDeviationsResponse](#diode-v1-RetrieveDeviationsResponse)       | Retrieve deviations      |
| RetrieveDeviationByID | [RetrieveDeviationByIDRequest](#diode-v1-RetrieveDeviationByIDRequest) | [RetrieveDeviationByIDResponse](#diode-v1-RetrieveDeviationByIDResponse) | Retrieve deviation by ID |

## Scalar Value Types

| .proto Type                    | Notes                                                                                                                                           | C++    | Java       | Python      | Go      | C#         | PHP            | Ruby                           |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------|--------|------------|-------------|---------|------------|----------------|--------------------------------|
| <a name="double" /> double     |                                                                                                                                                 | double | double     | float       | float64 | double     | float          | Float                          |
| <a name="float" /> float       |                                                                                                                                                 | float  | float      | float       | float32 | float      | float          | Float                          |
| <a name="int32" /> int32       | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="int64" /> int64       | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="uint32" /> uint32     | Uses variable-length encoding.                                                                                                                  | uint32 | int        | int/long    | uint32  | uint       | integer        | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64     | Uses variable-length encoding.                                                                                                                  | uint64 | long       | int/long    | uint64  | ulong      | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32     | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s.                            | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64     | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s.                            | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="fixed32" /> fixed32   | Always four bytes. More efficient than uint32 if values are often greater than 2^28.                                                            | uint32 | int        | int         | uint32  | uint       | integer        | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64   | Always eight bytes. More efficient than uint64 if values are often greater than 2^56.                                                           | uint64 | long       | int/long    | uint64  | ulong      | integer/string | Bignum                         |
| <a name="sfixed32" /> sfixed32 | Always four bytes.                                                                                                                              | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes.                                                                                                                             | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="bool" /> bool         |                                                                                                                                                 | bool   | boolean    | boolean     | bool    | bool       | boolean        | TrueClass/FalseClass           |
| <a name="string" /> string     | A string must always contain UTF-8 encoded or 7-bit ASCII text.                                                                                 | string | String     | str/unicode | string  | string     | string         | String (UTF-8)                 |
| <a name="bytes" /> bytes       | May contain any arbitrary sequence of bytes.                                                                                                    | string | ByteString | str         | []byte  | ByteString | string         | String (ASCII-8BIT)            |

