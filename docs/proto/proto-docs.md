<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [wasmx/websrv/params.proto](#wasmx/websrv/params.proto)
    - [Params](#wasmx.websrv.Params)
  
- [wasmx/websrv/genesis.proto](#wasmx/websrv/genesis.proto)
    - [GenesisState](#wasmx.websrv.GenesisState)
  
- [wasmx/websrv/query.proto](#wasmx/websrv/query.proto)
    - [HeaderItem](#wasmx.websrv.HeaderItem)
    - [HttpRequest](#wasmx.websrv.HttpRequest)
    - [HttpResponse](#wasmx.websrv.HttpResponse)
    - [QueryContractByRouteRequest](#wasmx.websrv.QueryContractByRouteRequest)
    - [QueryContractByRouteResponse](#wasmx.websrv.QueryContractByRouteResponse)
    - [QueryHttpRequestGet](#wasmx.websrv.QueryHttpRequestGet)
    - [QueryHttpResponseGet](#wasmx.websrv.QueryHttpResponseGet)
    - [QueryParamsRequest](#wasmx.websrv.QueryParamsRequest)
    - [QueryParamsResponse](#wasmx.websrv.QueryParamsResponse)
    - [QueryRouteByContractRequest](#wasmx.websrv.QueryRouteByContractRequest)
    - [QueryRouteByContractResponse](#wasmx.websrv.QueryRouteByContractResponse)
    - [RequestQueryParam](#wasmx.websrv.RequestQueryParam)
  
    - [HeaderOption](#wasmx.websrv.HeaderOption)
  
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



<a name="wasmx.websrv.HeaderItem"></a>

### HeaderItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `header_type` | [HeaderOption](#wasmx.websrv.HeaderOption) |  |  |
| `Value` | [string](#string) |  |  |






<a name="wasmx.websrv.HttpRequest"></a>

### HttpRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `header` | [HeaderItem](#wasmx.websrv.HeaderItem) | repeated |  |
| `query_params` | [RequestQueryParam](#wasmx.websrv.RequestQueryParam) | repeated |  |






<a name="wasmx.websrv.HttpResponse"></a>

### HttpResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `header` | [HeaderItem](#wasmx.websrv.HeaderItem) | repeated |  |
| `content` | [bytes](#bytes) |  |  |






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






<a name="wasmx.websrv.QueryHttpRequestGet"></a>

### QueryHttpRequestGet
QueryHttpGetRequest is the request type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `http_request` | [bytes](#bytes) |  |  |






<a name="wasmx.websrv.QueryHttpResponseGet"></a>

### QueryHttpResponseGet
QueryHttpResponseGet is the response type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [HttpResponse](#wasmx.websrv.HttpResponse) |  |  |






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






<a name="wasmx.websrv.RequestQueryParam"></a>

### RequestQueryParam



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |





 <!-- end messages -->


<a name="wasmx.websrv.HeaderOption"></a>

### HeaderOption
HeaderOption enumerates the valid http headers supported.

| Name | Number | Description |
| ---- | ------ | ----------- |
| content_type | 0 | content_type Content-Type indicates the media type of the resource; "text/html; charset=UTF-8" |
| content_encoding | 1 | content_encoding Content-Encoding is used to specify the compression algorithm |
| content_language | 2 | content_language Content-Language describes the human language(s) intended for the audience; "en" |
| location | 3 | location Location indicates the URL to redirect a page to |
| status | 4 | status Status indicates the status code response; "200" |
| www_authenticate | 5 | www_authenticate WWW-Authenticate defines the authentication method that should be used to access a resource |
| authorization | 6 | authorization Authorization contains the credentials to authenticate a user-agent with a server |
| content_length | 7 | content_length Content-Length indicates the size of the resource, in decimal number of bytes; "0" |
| content_location | 8 | content_location Content-Location "/" |
| gateway_interface | 9 | gateway_interface Gateway-Interface; ""CGI/1.1"" |
| connection | 10 | connection Connection controls whether the network connection stays open after the current transaction finishes |
| keep_alive | 11 | keep_alive Keep-Alive controls how long a persistent connection should stay open |
| cookie | 12 | cookie Cookie contains stored HTTP cookies previously sent by the server with the Set-Cookie header |
| set_cookie | 13 | set_cookie Set-Cookie send cookies from the server to the user-agent |
| access_control_allow_origin | 14 | access_control_allow_origin Access-Control-Allow-Origin indicates whether the response can be shared |
| server | 15 | server Server |
| auth_type | 16 | auth_type Auth-Type |
| accept | 17 | accept Accept |
| request_method | 18 | request_method Request-Method; "GET" |
| http_host | 19 | http_host Http-Host; "example.com" |
| path_info | 20 | path_info Path-Info; "/foo/bar" |
| query_string | 21 | query_string Query-String; "var1=value1&var2=with%20percent%20encoding" |
| remote_addr | 22 | remote_addr Remote-Addr |
| server_port | 23 | server_port ServerPort; "80" |
| accept_push_policy | 24 | status Accept-Push-Policy |
| accept_signature | 25 | accept_signature Accept-Signature indicates the intention to take advantage of any available signatures and to indicate what kinds of signatures it supports |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `HttpGet` | [QueryHttpRequestGet](#wasmx.websrv.QueryHttpRequestGet) | [QueryHttpResponseGet](#wasmx.websrv.QueryHttpResponseGet) | HttpGet makes a get request to the webserver | GET|/wasmx/websrv/v1/get/{http_request}|
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

