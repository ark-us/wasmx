<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [wasmx/websrv/params.proto](#wasmx/websrv/params.proto)
    - [Params](#wasmx.websrv.Params)
  
- [wasmx/websrv/genesis.proto](#wasmx/websrv/genesis.proto)
    - [GenesisState](#wasmx.websrv.GenesisState)
  
- [wasmx/websrv/query.proto](#wasmx/websrv/query.proto)
    - [HttpRequestGet](#wasmx.websrv.HttpRequestGet)
    - [HttpRequestGetResponse](#wasmx.websrv.HttpRequestGetResponse)
    - [QueryContractByRouteRequest](#wasmx.websrv.QueryContractByRouteRequest)
    - [QueryContractByRouteResponse](#wasmx.websrv.QueryContractByRouteResponse)
    - [QueryHttpGetRequest](#wasmx.websrv.QueryHttpGetRequest)
    - [QueryHttpGetResponse](#wasmx.websrv.QueryHttpGetResponse)
    - [QueryParamsRequest](#wasmx.websrv.QueryParamsRequest)
    - [QueryParamsResponse](#wasmx.websrv.QueryParamsResponse)
    - [QueryRouteByContractRequest](#wasmx.websrv.QueryRouteByContractRequest)
    - [QueryRouteByContractResponse](#wasmx.websrv.QueryRouteByContractResponse)
    - [RequestParam](#wasmx.websrv.RequestParam)
    - [RequestUrl](#wasmx.websrv.RequestUrl)
  
    - [Query](#wasmx.websrv.Query)
  
- [wasmx/websrv/tx.proto](#wasmx/websrv/tx.proto)
    - [MsgRegisterRoute](#wasmx.websrv.MsgRegisterRoute)
    - [MsgRegisterRouteResponse](#wasmx.websrv.MsgRegisterRouteResponse)
  
    - [Msg](#wasmx.websrv.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="wasmx/websrv/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/params.proto



<a name="wasmx.websrv.Params"></a>

### Params
Params defines the parameters for the module.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/websrv/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/genesis.proto



<a name="wasmx.websrv.GenesisState"></a>

### GenesisState
GenesisState defines the websrv module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.websrv.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/websrv/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/query.proto



<a name="wasmx.websrv.HttpRequestGet"></a>

### HttpRequestGet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `url` | [RequestUrl](#wasmx.websrv.RequestUrl) |  |  |






<a name="wasmx.websrv.HttpRequestGetResponse"></a>

### HttpRequestGetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `content` | [bytes](#bytes) |  | The http get response |
| `content_type` | [string](#string) |  | Content-Type |






<a name="wasmx.websrv.QueryContractByRouteRequest"></a>

### QueryContractByRouteRequest
QueryContractByRouteRequest is the request type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |






<a name="wasmx.websrv.QueryContractByRouteResponse"></a>

### QueryContractByRouteResponse
QueryContractByRouteResponse is the response type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="wasmx.websrv.QueryHttpGetRequest"></a>

### QueryHttpGetRequest
QueryHttpGetRequest is the request type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `http_request` | [bytes](#bytes) |  |  |






<a name="wasmx.websrv.QueryHttpGetResponse"></a>

### QueryHttpGetResponse
QueryHttpGetResponse is the response type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [HttpRequestGetResponse](#wasmx.websrv.HttpRequestGetResponse) |  |  |






<a name="wasmx.websrv.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="wasmx.websrv.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.websrv.Params) |  | params holds all the parameters of this module. |






<a name="wasmx.websrv.QueryRouteByContractRequest"></a>

### QueryRouteByContractRequest
QueryRouteByContractRequest is the request type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="wasmx.websrv.QueryRouteByContractResponse"></a>

### QueryRouteByContractResponse
QueryRouteByContractResponse is the response type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |






<a name="wasmx.websrv.RequestParam"></a>

### RequestParam



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |






<a name="wasmx.websrv.RequestUrl"></a>

### RequestUrl



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `Path` | [string](#string) |  |  |
| `params` | [RequestParam](#wasmx.websrv.RequestParam) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `HttpGet` | [QueryHttpGetRequest](#wasmx.websrv.QueryHttpGetRequest) | [QueryHttpGetResponse](#wasmx.websrv.QueryHttpGetResponse) | HttpGet makes a get request to the webserver | GET|/wasmx/websrv/v1/get/{http_request}|
| `ContractByRoute` | [QueryContractByRouteRequest](#wasmx.websrv.QueryContractByRouteRequest) | [QueryContractByRouteResponse](#wasmx.websrv.QueryContractByRouteResponse) | ContractByRoute gets the contract controlling a given route | GET|/wasmx/websrv/v1/route/{path}|
| `RouteByContract` | [QueryRouteByContractRequest](#wasmx.websrv.QueryRouteByContractRequest) | [QueryRouteByContractResponse](#wasmx.websrv.QueryRouteByContractResponse) | RouteByContract gets the route controlled by a given contract | GET|/wasmx/websrv/v1/contract/{contract_address}|
| `Params` | [QueryParamsRequest](#wasmx.websrv.QueryParamsRequest) | [QueryParamsResponse](#wasmx.websrv.QueryParamsResponse) | Parameters queries the parameters of the module. | GET|/wasmx/websrv/v1/params|

 <!-- end services -->



<a name="wasmx/websrv/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/tx.proto



<a name="wasmx.websrv.MsgRegisterRoute"></a>

### MsgRegisterRoute
MsgRegisterRoute submit Wasm code to the system


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `path` | [string](#string) |  | Route path |
| `contract_address` | [string](#string) |  | Contract address in bech32 format |






<a name="wasmx.websrv.MsgRegisterRouteResponse"></a>

### MsgRegisterRouteResponse
MsgRegisterRouteResponse returns store result data.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterRoute` | [MsgRegisterRoute](#wasmx.websrv.MsgRegisterRoute) | [MsgRegisterRouteResponse](#wasmx.websrv.MsgRegisterRouteResponse) | RegisterRoute to register a route | |

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

