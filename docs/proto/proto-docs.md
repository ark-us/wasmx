<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [mythos/network/v1/params.proto](#mythos/network/v1/params.proto)
    - [Params](#mythos.network.v1.Params)
  
- [mythos/network/v1/genesis.proto](#mythos/network/v1/genesis.proto)
    - [GenesisState](#mythos.network.v1.GenesisState)
  
- [mythos/network/v1/query.proto](#mythos/network/v1/query.proto)
    - [QueryParamsRequest](#mythos.network.v1.QueryParamsRequest)
    - [QueryParamsResponse](#mythos.network.v1.QueryParamsResponse)
  
    - [Query](#mythos.network.v1.Query)
  
- [mythos/network/v1/server.proto](#mythos/network/v1/server.proto)
    - [MsgPing](#mythos.network.v1.MsgPing)
    - [MsgPingResponse](#mythos.network.v1.MsgPingResponse)
  
    - [Msg](#mythos.network.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="mythos/network/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/params.proto



<a name="mythos.network.v1.Params"></a>

### Params
Params defines the parameters for the module.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mythos/network/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/genesis.proto



<a name="mythos.network.v1.GenesisState"></a>

### GenesisState
GenesisState defines the network module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#mythos.network.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mythos/network/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/query.proto



<a name="mythos.network.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="mythos.network.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#mythos.network.v1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.network.v1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#mythos.network.v1.QueryParamsRequest) | [QueryParamsResponse](#mythos.network.v1.QueryParamsResponse) | Params gets the module params | GET|/network/v1/codes/params|

 <!-- end services -->



<a name="mythos/network/v1/server.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/server.proto



<a name="mythos.network.v1.MsgPing"></a>

### MsgPing
MsgPing


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `message` | [string](#string) |  |  |






<a name="mythos.network.v1.MsgPingResponse"></a>

### MsgPingResponse
MsgPingResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `message` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.network.v1.Msg"></a>

### Msg
Msg defines the grpc server

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Ping` | [MsgPing](#mythos.network.v1.MsgPing) | [MsgPingResponse](#mythos.network.v1.MsgPingResponse) | Ping | |

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

