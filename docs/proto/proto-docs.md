<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [wasmx/websrv/v1/params.proto](#wasmx/websrv/v1/params.proto)
    - [Params](#wasmx.websrv.v1.Params)
  
- [wasmx/websrv/v1/genesis.proto](#wasmx/websrv/v1/genesis.proto)
    - [GenesisState](#wasmx.websrv.v1.GenesisState)
  
- [wasmx/websrv/v1/proposal.proto](#wasmx/websrv/v1/proposal.proto)
    - [DeregisterRouteProposal](#wasmx.websrv.v1.DeregisterRouteProposal)
    - [RegisterRouteProposal](#wasmx.websrv.v1.RegisterRouteProposal)
  
- [wasmx/websrv/v1/query.proto](#wasmx/websrv/v1/query.proto)
    - [OauthClientInfo](#wasmx.websrv.v1.OauthClientInfo)
    - [QueryContractByRouteRequest](#wasmx.websrv.v1.QueryContractByRouteRequest)
    - [QueryContractByRouteResponse](#wasmx.websrv.v1.QueryContractByRouteResponse)
    - [QueryGetAllOauthClientsRequest](#wasmx.websrv.v1.QueryGetAllOauthClientsRequest)
    - [QueryGetAllOauthClientsResponse](#wasmx.websrv.v1.QueryGetAllOauthClientsResponse)
    - [QueryGetOauthClientRequest](#wasmx.websrv.v1.QueryGetOauthClientRequest)
    - [QueryGetOauthClientResponse](#wasmx.websrv.v1.QueryGetOauthClientResponse)
    - [QueryGetOauthClientsByOwnerRequest](#wasmx.websrv.v1.QueryGetOauthClientsByOwnerRequest)
    - [QueryGetOauthClientsByOwnerResponse](#wasmx.websrv.v1.QueryGetOauthClientsByOwnerResponse)
    - [QueryHttpRequestGet](#wasmx.websrv.v1.QueryHttpRequestGet)
    - [QueryHttpResponseGet](#wasmx.websrv.v1.QueryHttpResponseGet)
    - [QueryParamsRequest](#wasmx.websrv.v1.QueryParamsRequest)
    - [QueryParamsResponse](#wasmx.websrv.v1.QueryParamsResponse)
    - [QueryRouteByContractRequest](#wasmx.websrv.v1.QueryRouteByContractRequest)
    - [QueryRouteByContractResponse](#wasmx.websrv.v1.QueryRouteByContractResponse)
  
    - [Query](#wasmx.websrv.v1.Query)
  
- [wasmx/websrv/v1/tx.proto](#wasmx/websrv/v1/tx.proto)
    - [MsgDeregisterOAuthClient](#wasmx.websrv.v1.MsgDeregisterOAuthClient)
    - [MsgDeregisterOAuthClientResponse](#wasmx.websrv.v1.MsgDeregisterOAuthClientResponse)
    - [MsgEditOAuthClient](#wasmx.websrv.v1.MsgEditOAuthClient)
    - [MsgEditOAuthClientResponse](#wasmx.websrv.v1.MsgEditOAuthClientResponse)
    - [MsgRegisterOAuthClient](#wasmx.websrv.v1.MsgRegisterOAuthClient)
    - [MsgRegisterOAuthClientResponse](#wasmx.websrv.v1.MsgRegisterOAuthClientResponse)
  
    - [Msg](#wasmx.websrv.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="wasmx/websrv/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/v1/params.proto



<a name="wasmx.websrv.v1.Params"></a>

### Params
Params defines the parameters for the module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `oauth_client_registration_only_e_id` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/websrv/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/v1/genesis.proto



<a name="wasmx.websrv.v1.GenesisState"></a>

### GenesisState
GenesisState defines the websrv module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.websrv.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/websrv/v1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/v1/proposal.proto



<a name="wasmx.websrv.v1.DeregisterRouteProposal"></a>

### DeregisterRouteProposal
DisallowCosmosMessagesProposal is a gov Content type to remove a previously
allowed Cosmos message or query to be called from the EVM


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | description of the proposal |
| `path` | [string](#string) |  | Route path |
| `contract_address` | [string](#string) |  | Contract address in bech32 format |






<a name="wasmx.websrv.v1.RegisterRouteProposal"></a>

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



<a name="wasmx/websrv/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/v1/query.proto



<a name="wasmx.websrv.v1.OauthClientInfo"></a>

### OauthClientInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `domain` | [string](#string) |  |  |
| `public` | [bool](#bool) |  |  |






<a name="wasmx.websrv.v1.QueryContractByRouteRequest"></a>

### QueryContractByRouteRequest
QueryContractByRouteRequest is the request type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |






<a name="wasmx.websrv.v1.QueryContractByRouteResponse"></a>

### QueryContractByRouteResponse
QueryContractByRouteResponse is the response type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="wasmx.websrv.v1.QueryGetAllOauthClientsRequest"></a>

### QueryGetAllOauthClientsRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="wasmx.websrv.v1.QueryGetAllOauthClientsResponse"></a>

### QueryGetAllOauthClientsResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `clients` | [OauthClientInfo](#wasmx.websrv.v1.OauthClientInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="wasmx.websrv.v1.QueryGetOauthClientRequest"></a>

### QueryGetOauthClientRequest
QueryGetOauthClientRequest is the request type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |






<a name="wasmx.websrv.v1.QueryGetOauthClientResponse"></a>

### QueryGetOauthClientResponse
QueryGetOauthClientResponse is the response type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client` | [OauthClientInfo](#wasmx.websrv.v1.OauthClientInfo) |  |  |






<a name="wasmx.websrv.v1.QueryGetOauthClientsByOwnerRequest"></a>

### QueryGetOauthClientsByOwnerRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |






<a name="wasmx.websrv.v1.QueryGetOauthClientsByOwnerResponse"></a>

### QueryGetOauthClientsByOwnerResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [uint64](#uint64) | repeated |  |






<a name="wasmx.websrv.v1.QueryHttpRequestGet"></a>

### QueryHttpRequestGet
QueryHttpGetRequest is the request type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `http_request` | [bytes](#bytes) |  |  |






<a name="wasmx.websrv.v1.QueryHttpResponseGet"></a>

### QueryHttpResponseGet
QueryHttpResponseGet is the response type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | HttpResponse data = 1; |






<a name="wasmx.websrv.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="wasmx.websrv.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.websrv.v1.Params) |  | params holds all the parameters of this module. |






<a name="wasmx.websrv.v1.QueryRouteByContractRequest"></a>

### QueryRouteByContractRequest
QueryRouteByContractRequest is the request type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="wasmx.websrv.v1.QueryRouteByContractResponse"></a>

### QueryRouteByContractResponse
QueryRouteByContractResponse is the response type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `HttpGet` | [QueryHttpRequestGet](#wasmx.websrv.v1.QueryHttpRequestGet) | [QueryHttpResponseGet](#wasmx.websrv.v1.QueryHttpResponseGet) | HttpGet makes a get request to the webserver | GET|/wasmx/websrv/v1/get/{http_request}|
| `ContractByRoute` | [QueryContractByRouteRequest](#wasmx.websrv.v1.QueryContractByRouteRequest) | [QueryContractByRouteResponse](#wasmx.websrv.v1.QueryContractByRouteResponse) | ContractByRoute gets the contract controlling a given route | GET|/wasmx/websrv/v1/route/{path}|
| `RouteByContract` | [QueryRouteByContractRequest](#wasmx.websrv.v1.QueryRouteByContractRequest) | [QueryRouteByContractResponse](#wasmx.websrv.v1.QueryRouteByContractResponse) | RouteByContract gets the route controlled by a given contract | GET|/wasmx/websrv/v1/contract/{contract_address}|
| `Params` | [QueryParamsRequest](#wasmx.websrv.v1.QueryParamsRequest) | [QueryParamsResponse](#wasmx.websrv.v1.QueryParamsResponse) | Parameters queries the parameters of the module. | GET|/wasmx/websrv/v1/params|
| `GetAllOauthClients` | [QueryGetAllOauthClientsRequest](#wasmx.websrv.v1.QueryGetAllOauthClientsRequest) | [QueryGetAllOauthClientsResponse](#wasmx.websrv.v1.QueryGetAllOauthClientsResponse) | GetAllClients gets all the registered client apps for the oauth service | GET|/wasmx/websrv/v1/oauth/clients|
| `GetOauthClient` | [QueryGetOauthClientRequest](#wasmx.websrv.v1.QueryGetOauthClientRequest) | [QueryGetOauthClientResponse](#wasmx.websrv.v1.QueryGetOauthClientResponse) | GetOauthClient gets the registered oauth client by client id | GET|/wasmx/websrv/v1/oauth/client/{client_id}|
| `GetOauthClientsByOwner` | [QueryGetOauthClientsByOwnerRequest](#wasmx.websrv.v1.QueryGetOauthClientsByOwnerRequest) | [QueryGetOauthClientsByOwnerResponse](#wasmx.websrv.v1.QueryGetOauthClientsByOwnerResponse) | GetOauthClientsByOwner gets all the registered oauth client by an owner address | GET|/wasmx/websrv/v1/oauth/clients/{owner}|

 <!-- end services -->



<a name="wasmx/websrv/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/websrv/v1/tx.proto



<a name="wasmx.websrv.v1.MsgDeregisterOAuthClient"></a>

### MsgDeregisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |






<a name="wasmx.websrv.v1.MsgDeregisterOAuthClientResponse"></a>

### MsgDeregisterOAuthClientResponse







<a name="wasmx.websrv.v1.MsgEditOAuthClient"></a>

### MsgEditOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |
| `domain` | [string](#string) |  |  |






<a name="wasmx.websrv.v1.MsgEditOAuthClientResponse"></a>

### MsgEditOAuthClientResponse







<a name="wasmx.websrv.v1.MsgRegisterOAuthClient"></a>

### MsgRegisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `domain` | [string](#string) |  |  |






<a name="wasmx.websrv.v1.MsgRegisterOAuthClientResponse"></a>

### MsgRegisterOAuthClientResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.websrv.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterOAuthClient` | [MsgRegisterOAuthClient](#wasmx.websrv.v1.MsgRegisterOAuthClient) | [MsgRegisterOAuthClientResponse](#wasmx.websrv.v1.MsgRegisterOAuthClientResponse) | Register OAuth client | |
| `EditOAuthClient` | [MsgEditOAuthClient](#wasmx.websrv.v1.MsgEditOAuthClient) | [MsgEditOAuthClientResponse](#wasmx.websrv.v1.MsgEditOAuthClientResponse) | Edit OAuth client | |
| `DeregisterOAuthClient` | [MsgDeregisterOAuthClient](#wasmx.websrv.v1.MsgDeregisterOAuthClient) | [MsgDeregisterOAuthClientResponse](#wasmx.websrv.v1.MsgDeregisterOAuthClientResponse) | Deregister OAuth client | |

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

