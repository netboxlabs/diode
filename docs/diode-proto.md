# Protocol Documentation

<a name="top"></a>

## Table of Contents

- [diode/v1/manufacturer.proto](#diode_v1_manufacturer-proto)
    - [Manufacturer](#diode-v1-Manufacturer)

- [diode/v1/device_type.proto](#diode_v1_device_type-proto)
    - [DeviceType](#diode-v1-DeviceType)

- [diode/v1/platform.proto](#diode_v1_platform-proto)
    - [Platform](#diode-v1-Platform)

- [diode/v1/role.proto](#diode_v1_role-proto)
    - [Role](#diode-v1-Role)

- [diode/v1/site.proto](#diode_v1_site-proto)
    - [Site](#diode-v1-Site)

- [diode/v1/device.proto](#diode_v1_device-proto)
    - [Device](#diode-v1-Device)

- [diode/v1/interface.proto](#diode_v1_interface-proto)
    - [Interface](#diode-v1-Interface)

- [diode/v1/distributor.proto](#diode_v1_distributor-proto)
    - [IngestEntity](#diode-v1-IngestEntity)
    - [PushRequest](#diode-v1-PushRequest)
    - [PushResponse](#diode-v1-PushResponse)

    - [DistributorService](#diode-v1-DistributorService)

- [Scalar Value Types](#scalar-value-types)

<a name="diode_v1_manufacturer-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/manufacturer.proto

<a name="diode-v1-Manufacturer"></a>

### Manufacturer

A manufacturer

| Field | Type              | Label | Description |
|-------|-------------------|-------|-------------|
| name  | [string](#string) |       |             |
| slug  | [string](#string) |       |             |

<a name="diode_v1_device_type-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/device_type.proto

<a name="diode-v1-DeviceType"></a>

### DeviceType

A device type

| Field        | Type                                   | Label | Description |
|--------------|----------------------------------------|-------|-------------|
| manufacturer | [Manufacturer](#diode-v1-Manufacturer) |       |             |
| model        | [string](#string)                      |       |             |
| slug         | [string](#string)                      |       |             |

<a name="diode_v1_platform-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/platform.proto

<a name="diode-v1-Platform"></a>

### Platform

A platform

| Field | Type              | Label | Description |
|-------|-------------------|-------|-------------|
| id    | [uint64](#uint64) |       |             |
| name  | [string](#string) |       |             |

<a name="diode_v1_role-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/role.proto

<a name="diode-v1-Role"></a>

### Role

A role

| Field   | Type              | Label | Description |
|---------|-------------------|-------|-------------|
| name    | [string](#string) |       |             |
| slug    | [string](#string) |       |             |
| vm_role | [bool](#bool)     |       |             |

<a name="diode_v1_site-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/site.proto

<a name="diode-v1-Site"></a>

### Site

A site

| Field | Type              | Label | Description |
|-------|-------------------|-------|-------------|
| name  | [string](#string) |       |             |
| slug  | [string](#string) |       |             |

<a name="diode_v1_device-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/device.proto

<a name="diode-v1-Device"></a>

### Device

A device

| Field       | Type                               | Label | Description |
|-------------|------------------------------------|-------|-------------|
| name        | [string](#string)                  |       |             |
| device_fqdn | [string](#string)                  |       |             |
| device_type | [DeviceType](#diode-v1-DeviceType) |       |             |
| role        | [Role](#diode-v1-Role)             |       |             |
| platform    | [Platform](#diode-v1-Platform)     |       |             |
| serial      | [string](#string)                  |       |             |
| site        | [Site](#diode-v1-Site)             |       |             |
| vc_position | [int32](#int32)                    |       |             |

<a name="diode_v1_interface-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/interface.proto

<a name="diode-v1-Interface"></a>

### Interface

An interface

| Field       | Type                                        | Label | Description |
|-------------|---------------------------------------------|-------|-------------|
| device      | [google.protobuf.Any](#google-protobuf-Any) |       |             |
| name        | [string](#string)                           |       |             |
| type        | [string](#string)                           |       |             |
| enabled     | [bool](#bool)                               |       |             |
| mtu         | [int32](#int32)                             |       |             |
| mac_address | [string](#string)                           |       |             |
| mgmt_only   | [bool](#bool)                               |       |             |

<a name="diode_v1_distributor-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## diode/v1/distributor.proto

<a name="diode-v1-IngestEntity"></a>

### IngestEntity

An ingest entity wrapper

| Field        | Type                                                    | Label | Description                                   |
|--------------|---------------------------------------------------------|-------|-----------------------------------------------|
| site         | [Site](#diode-v1-Site)                                  |       |                                               |
| platform     | [Platform](#diode-v1-Platform)                          |       |                                               |
| manufacturer | [Manufacturer](#diode-v1-Manufacturer)                  |       |                                               |
| device       | [Device](#diode-v1-Device)                              |       |                                               |
| device_role  | [Role](#diode-v1-Role)                                  |       |                                               |
| device_type  | [DeviceType](#diode-v1-DeviceType)                      |       |                                               |
| interface    | [Interface](#diode-v1-Interface)                        |       |                                               |
| timestamp    | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |       | The timestamp of the data discovery at source |

<a name="diode-v1-PushRequest"></a>

### PushRequest

The request to push data

| Field                | Type                                   | Label    | Description |
|----------------------|----------------------------------------|----------|-------------|
| stream               | [string](#string)                      |          |             |
| data                 | [IngestEntity](#diode-v1-IngestEntity) | repeated |             |
| id                   | [string](#string)                      |          |             |
| producer_app_name    | [string](#string)                      |          |             |
| producer_app_version | [string](#string)                      |          |             |
| sdk_name             | [string](#string)                      |          |             |
| sdk_version          | [string](#string)                      |          |             |

<a name="diode-v1-PushResponse"></a>

### PushResponse

The response from the push request

| Field  | Type              | Label    | Description |
|--------|-------------------|----------|-------------|
| errors | [string](#string) | repeated |             |

<a name="diode-v1-DistributorService"></a>

### DistributorService

Distributor API

| Method Name | Request Type                         | Response Type                          | Description                  |
|-------------|--------------------------------------|----------------------------------------|------------------------------|
| Push        | [PushRequest](#diode-v1-PushRequest) | [PushResponse](#diode-v1-PushResponse) | Ingests data into the system |

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

