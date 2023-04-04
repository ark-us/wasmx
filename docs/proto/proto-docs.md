<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [wasmx/websrv/params.proto](#wasmx/websrv/params.proto)
    - [Params](#wasmx.websrv.Params)
  
- [wasmx/websrv/genesis.proto](#wasmx/websrv/genesis.proto)
    - [GenesisState](#wasmx.websrv.GenesisState)
  
- [wasmx/websrv/proposal.proto](#wasmx/websrv/proposal.proto)
    - [DeregisterRouteProposal](#wasmx.websrv.DeregisterRouteProposal)
    - [RegisterRouteProposal](#wasmx.websrv.RegisterRouteProposal)
  
- [wasmx/websrv/query.proto](#wasmx/websrv/query.proto)
    - [OauthClientInfo](#wasmx.websrv.OauthClientInfo)
    - [QueryContractByRouteRequest](#wasmx.websrv.QueryContractByRouteRequest)
    - [QueryContractByRouteResponse](#wasmx.websrv.QueryContractByRouteResponse)
    - [QueryGetAllOauthClientsRequest](#wasmx.websrv.QueryGetAllOauthClientsRequest)
    - [QueryGetAllOauthClientsResponse](#wasmx.websrv.QueryGetAllOauthClientsResponse)
    - [QueryGetOauthClientRequest](#wasmx.websrv.QueryGetOauthClientRequest)
    - [QueryGetOauthClientResponse](#wasmx.websrv.QueryGetOauthClientResponse)
    - [QueryGetOauthClientsByOwnerRequest](#wasmx.websrv.QueryGetOauthClientsByOwnerRequest)
    - [QueryGetOauthClientsByOwnerResponse](#wasmx.websrv.QueryGetOauthClientsByOwnerResponse)
    - [QueryHttpRequestGet](#wasmx.websrv.QueryHttpRequestGet)
    - [QueryHttpResponseGet](#wasmx.websrv.QueryHttpResponseGet)
    - [QueryParamsRequest](#wasmx.websrv.QueryParamsRequest)
    - [QueryParamsResponse](#wasmx.websrv.QueryParamsResponse)
    - [QueryRouteByContractRequest](#wasmx.websrv.QueryRouteByContractRequest)
    - [QueryRouteByContractResponse](#wasmx.websrv.QueryRouteByContractResponse)
  
    - [Query](#wasmx.websrv.Query)
  
- [wasmx/websrv/tx.proto](#wasmx/websrv/tx.proto)
    - [MsgDeregisterOAuthClient](#wasmx.websrv.MsgDeregisterOAuthClient)
    - [MsgDeregisterOAuthClientResponse](#wasmx.websrv.MsgDeregisterOAuthClientResponse)
    - [MsgEditOAuthClient](#wasmx.websrv.MsgEditOAuthClient)
    - [MsgEditOAuthClientResponse](#wasmx.websrv.MsgEditOAuthClientResponse)
    - [MsgRegisterOAuthClient](#wasmx.websrv.MsgRegisterOAuthClient)
    - [MsgRegisterOAuthClientResponse](#wasmx.websrv.MsgRegisterOAuthClientResponse)
  
    - [Msg](#wasmx.websrv.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="wasmx/websrv/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/params.proto



<a name="wasmx.websrv.Params"></a>

### Params
Params defines the parameters for the module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `oauth_client_registration_only_e_id` | [bool](#bool) |  |  |





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



<a name="wasmx/websrv/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/proposal.proto



<a name="wasmx.websrv.DeregisterRouteProposal"></a>

### DeregisterRouteProposal
DisallowCosmosMessagesProposal is a gov Content type to remove a previously
allowed Cosmos message or query to be called from the EVM


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | description of the proposal |
| `path` | [string](#string) |  | Route path |
| `contract_address` | [string](#string) |  | Contract address in bech32 format |






<a name="wasmx.websrv.RegisterRouteProposal"></a>

### RegisterRouteProposal
RegisterRouteProposal is a gov Content type to register a web server route


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | description of the proposal |
| `path` | [string](#string) |  | Route path |
| `contract_address` | [string](#string) |  | Contract address in bech32 format |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/websrv/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/query.proto



<a name="wasmx.websrv.OauthClientInfo"></a>

### OauthClientInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `domain` | [string](#string) |  |  |
| `public` | [bool](#bool) |  |  |






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






<a name="wasmx.websrv.QueryGetAllOauthClientsRequest"></a>

### QueryGetAllOauthClientsRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="wasmx.websrv.QueryGetAllOauthClientsResponse"></a>

### QueryGetAllOauthClientsResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `clients` | [OauthClientInfo](#wasmx.websrv.OauthClientInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="wasmx.websrv.QueryGetOauthClientRequest"></a>

### QueryGetOauthClientRequest
QueryGetOauthClientRequest is the request type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |






<a name="wasmx.websrv.QueryGetOauthClientResponse"></a>

### QueryGetOauthClientResponse
QueryGetOauthClientResponse is the response type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client` | [OauthClientInfo](#wasmx.websrv.OauthClientInfo) |  |  |






<a name="wasmx.websrv.QueryGetOauthClientsByOwnerRequest"></a>

### QueryGetOauthClientsByOwnerRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |






<a name="wasmx.websrv.QueryGetOauthClientsByOwnerResponse"></a>

### QueryGetOauthClientsByOwnerResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [uint64](#uint64) | repeated |  |






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
| `data` | [bytes](#bytes) |  | HttpResponse data = 1; |






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





 <!-- end messages -->

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
| `GetAllOauthClients` | [QueryGetAllOauthClientsRequest](#wasmx.websrv.QueryGetAllOauthClientsRequest) | [QueryGetAllOauthClientsResponse](#wasmx.websrv.QueryGetAllOauthClientsResponse) | GetAllClients gets all the registered client apps for the oauth service | GET|/wasmx/websrv/v1/oauth/clients|
| `GetOauthClient` | [QueryGetOauthClientRequest](#wasmx.websrv.QueryGetOauthClientRequest) | [QueryGetOauthClientResponse](#wasmx.websrv.QueryGetOauthClientResponse) | GetOauthClient gets the registered oauth client by client id | GET|/wasmx/websrv/v1/oauth/client/{client_id}|
| `GetOauthClientsByOwner` | [QueryGetOauthClientsByOwnerRequest](#wasmx.websrv.QueryGetOauthClientsByOwnerRequest) | [QueryGetOauthClientsByOwnerResponse](#wasmx.websrv.QueryGetOauthClientsByOwnerResponse) | GetOauthClientsByOwner gets all the registered oauth client by an owner address | GET|/wasmx/websrv/v1/oauth/clients/{owner}|

 <!-- end services -->



<a name="wasmx/websrv/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/tx.proto



<a name="wasmx.websrv.MsgDeregisterOAuthClient"></a>

### MsgDeregisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |






<a name="wasmx.websrv.MsgDeregisterOAuthClientResponse"></a>

### MsgDeregisterOAuthClientResponse







<a name="wasmx.websrv.MsgEditOAuthClient"></a>

### MsgEditOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |
| `domain` | [string](#string) |  |  |






<a name="wasmx.websrv.MsgEditOAuthClientResponse"></a>

### MsgEditOAuthClientResponse







<a name="wasmx.websrv.MsgRegisterOAuthClient"></a>

### MsgRegisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `domain` | [string](#string) |  |  |






<a name="wasmx.websrv.MsgRegisterOAuthClientResponse"></a>

### MsgRegisterOAuthClientResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterOAuthClient` | [MsgRegisterOAuthClient](#wasmx.websrv.MsgRegisterOAuthClient) | [MsgRegisterOAuthClientResponse](#wasmx.websrv.MsgRegisterOAuthClientResponse) | Register OAuth client | |
| `EditOAuthClient` | [MsgEditOAuthClient](#wasmx.websrv.MsgEditOAuthClient) | [MsgEditOAuthClientResponse](#wasmx.websrv.MsgEditOAuthClientResponse) | Edit OAuth client | |
| `DeregisterOAuthClient` | [MsgDeregisterOAuthClient](#wasmx.websrv.MsgDeregisterOAuthClient) | [MsgDeregisterOAuthClientResponse](#wasmx.websrv.MsgDeregisterOAuthClientResponse) | Deregister OAuth client | |

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

