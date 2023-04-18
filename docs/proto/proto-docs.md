<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [mythos/websrv/v1/params.proto](#mythos/websrv/v1/params.proto)
    - [Params](#mythos.websrv.v1.Params)
  
- [mythos/websrv/v1/genesis.proto](#mythos/websrv/v1/genesis.proto)
    - [GenesisState](#mythos.websrv.v1.GenesisState)
  
- [mythos/websrv/v1/proposal.proto](#mythos/websrv/v1/proposal.proto)
    - [DeregisterRouteProposal](#mythos.websrv.v1.DeregisterRouteProposal)
    - [RegisterRouteProposal](#mythos.websrv.v1.RegisterRouteProposal)
  
- [mythos/websrv/v1/query.proto](#mythos/websrv/v1/query.proto)
    - [OauthClientInfo](#mythos.websrv.v1.OauthClientInfo)
    - [QueryContractByRouteRequest](#mythos.websrv.v1.QueryContractByRouteRequest)
    - [QueryContractByRouteResponse](#mythos.websrv.v1.QueryContractByRouteResponse)
    - [QueryGetAllOauthClientsRequest](#mythos.websrv.v1.QueryGetAllOauthClientsRequest)
    - [QueryGetAllOauthClientsResponse](#mythos.websrv.v1.QueryGetAllOauthClientsResponse)
    - [QueryGetOauthClientRequest](#mythos.websrv.v1.QueryGetOauthClientRequest)
    - [QueryGetOauthClientResponse](#mythos.websrv.v1.QueryGetOauthClientResponse)
    - [QueryGetOauthClientsByOwnerRequest](#mythos.websrv.v1.QueryGetOauthClientsByOwnerRequest)
    - [QueryGetOauthClientsByOwnerResponse](#mythos.websrv.v1.QueryGetOauthClientsByOwnerResponse)
    - [QueryHttpRequestGet](#mythos.websrv.v1.QueryHttpRequestGet)
    - [QueryHttpResponseGet](#mythos.websrv.v1.QueryHttpResponseGet)
    - [QueryParamsRequest](#mythos.websrv.v1.QueryParamsRequest)
    - [QueryParamsResponse](#mythos.websrv.v1.QueryParamsResponse)
    - [QueryRouteByContractRequest](#mythos.websrv.v1.QueryRouteByContractRequest)
    - [QueryRouteByContractResponse](#mythos.websrv.v1.QueryRouteByContractResponse)
  
    - [Query](#mythos.websrv.v1.Query)
  
- [mythos/websrv/v1/tx.proto](#mythos/websrv/v1/tx.proto)
    - [MsgDeregisterOAuthClient](#mythos.websrv.v1.MsgDeregisterOAuthClient)
    - [MsgDeregisterOAuthClientResponse](#mythos.websrv.v1.MsgDeregisterOAuthClientResponse)
    - [MsgEditOAuthClient](#mythos.websrv.v1.MsgEditOAuthClient)
    - [MsgEditOAuthClientResponse](#mythos.websrv.v1.MsgEditOAuthClientResponse)
    - [MsgRegisterOAuthClient](#mythos.websrv.v1.MsgRegisterOAuthClient)
    - [MsgRegisterOAuthClientResponse](#mythos.websrv.v1.MsgRegisterOAuthClientResponse)
  
    - [Msg](#mythos.websrv.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="mythos/websrv/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/websrv/v1/params.proto



<a name="mythos.websrv.v1.Params"></a>

### Params
Params defines the parameters for the module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `oauth_client_registration_only_e_id` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mythos/websrv/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/websrv/v1/genesis.proto



<a name="mythos.websrv.v1.GenesisState"></a>

### GenesisState
GenesisState defines the websrv module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#mythos.websrv.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mythos/websrv/v1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/websrv/v1/proposal.proto



<a name="mythos.websrv.v1.DeregisterRouteProposal"></a>

### DeregisterRouteProposal
DisallowCosmosMessagesProposal is a gov Content type to remove a previously
allowed Cosmos message or query to be called from the EVM


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | description of the proposal |
| `path` | [string](#string) |  | Route path |
| `contract_address` | [string](#string) |  | Contract address in bech32 format |






<a name="mythos.websrv.v1.RegisterRouteProposal"></a>

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



<a name="mythos/websrv/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/websrv/v1/query.proto



<a name="mythos.websrv.v1.OauthClientInfo"></a>

### OauthClientInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `domain` | [string](#string) |  |  |
| `public` | [bool](#bool) |  |  |






<a name="mythos.websrv.v1.QueryContractByRouteRequest"></a>

### QueryContractByRouteRequest
QueryContractByRouteRequest is the request type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |






<a name="mythos.websrv.v1.QueryContractByRouteResponse"></a>

### QueryContractByRouteResponse
QueryContractByRouteResponse is the response type for the
Query/ContractByRoute RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="mythos.websrv.v1.QueryGetAllOauthClientsRequest"></a>

### QueryGetAllOauthClientsRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="mythos.websrv.v1.QueryGetAllOauthClientsResponse"></a>

### QueryGetAllOauthClientsResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetAllOauthClients RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `clients` | [OauthClientInfo](#mythos.websrv.v1.OauthClientInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="mythos.websrv.v1.QueryGetOauthClientRequest"></a>

### QueryGetOauthClientRequest
QueryGetOauthClientRequest is the request type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |






<a name="mythos.websrv.v1.QueryGetOauthClientResponse"></a>

### QueryGetOauthClientResponse
QueryGetOauthClientResponse is the response type for the
Query/GetOauthClient RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client` | [OauthClientInfo](#mythos.websrv.v1.OauthClientInfo) |  |  |






<a name="mythos.websrv.v1.QueryGetOauthClientsByOwnerRequest"></a>

### QueryGetOauthClientsByOwnerRequest
QueryGetAllOauthClientsRequest is the request type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |






<a name="mythos.websrv.v1.QueryGetOauthClientsByOwnerResponse"></a>

### QueryGetOauthClientsByOwnerResponse
QueryGetAllOauthClientsResponse is the response type for the
Query/GetOauthClientsByOwner RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [uint64](#uint64) | repeated |  |






<a name="mythos.websrv.v1.QueryHttpRequestGet"></a>

### QueryHttpRequestGet
QueryHttpGetRequest is the request type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `http_request` | [bytes](#bytes) |  |  |






<a name="mythos.websrv.v1.QueryHttpResponseGet"></a>

### QueryHttpResponseGet
QueryHttpResponseGet is the response type for the
Query/HttpGet RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | HttpResponse data = 1; |






<a name="mythos.websrv.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="mythos.websrv.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#mythos.websrv.v1.Params) |  | params holds all the parameters of this module. |






<a name="mythos.websrv.v1.QueryRouteByContractRequest"></a>

### QueryRouteByContractRequest
QueryRouteByContractRequest is the request type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |






<a name="mythos.websrv.v1.QueryRouteByContractResponse"></a>

### QueryRouteByContractResponse
QueryRouteByContractResponse is the response type for the
Query/RouteByContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.websrv.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `HttpGet` | [QueryHttpRequestGet](#mythos.websrv.v1.QueryHttpRequestGet) | [QueryHttpResponseGet](#mythos.websrv.v1.QueryHttpResponseGet) | HttpGet makes a get request to the webserver | GET|/websrv/v1/get/{http_request}|
| `ContractByRoute` | [QueryContractByRouteRequest](#mythos.websrv.v1.QueryContractByRouteRequest) | [QueryContractByRouteResponse](#mythos.websrv.v1.QueryContractByRouteResponse) | ContractByRoute gets the contract controlling a given route | GET|/websrv/v1/route/{path}|
| `RouteByContract` | [QueryRouteByContractRequest](#mythos.websrv.v1.QueryRouteByContractRequest) | [QueryRouteByContractResponse](#mythos.websrv.v1.QueryRouteByContractResponse) | RouteByContract gets the route controlled by a given contract | GET|/websrv/v1/contract/{contract_address}|
| `Params` | [QueryParamsRequest](#mythos.websrv.v1.QueryParamsRequest) | [QueryParamsResponse](#mythos.websrv.v1.QueryParamsResponse) | Parameters queries the parameters of the module. | GET|/websrv/v1/params|
| `GetAllOauthClients` | [QueryGetAllOauthClientsRequest](#mythos.websrv.v1.QueryGetAllOauthClientsRequest) | [QueryGetAllOauthClientsResponse](#mythos.websrv.v1.QueryGetAllOauthClientsResponse) | GetAllClients gets all the registered client apps for the oauth service | GET|/websrv/v1/oauth/clients|
| `GetOauthClient` | [QueryGetOauthClientRequest](#mythos.websrv.v1.QueryGetOauthClientRequest) | [QueryGetOauthClientResponse](#mythos.websrv.v1.QueryGetOauthClientResponse) | GetOauthClient gets the registered oauth client by client id | GET|/websrv/v1/oauth/client/{client_id}|
| `GetOauthClientsByOwner` | [QueryGetOauthClientsByOwnerRequest](#mythos.websrv.v1.QueryGetOauthClientsByOwnerRequest) | [QueryGetOauthClientsByOwnerResponse](#mythos.websrv.v1.QueryGetOauthClientsByOwnerResponse) | GetOauthClientsByOwner gets all the registered oauth client by an owner address | GET|/websrv/v1/oauth/clients/{owner}|

 <!-- end services -->



<a name="mythos/websrv/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/websrv/v1/tx.proto



<a name="mythos.websrv.v1.MsgDeregisterOAuthClient"></a>

### MsgDeregisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |






<a name="mythos.websrv.v1.MsgDeregisterOAuthClientResponse"></a>

### MsgDeregisterOAuthClientResponse







<a name="mythos.websrv.v1.MsgEditOAuthClient"></a>

### MsgEditOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `client_id` | [uint64](#uint64) |  |  |
| `domain` | [string](#string) |  |  |






<a name="mythos.websrv.v1.MsgEditOAuthClientResponse"></a>

### MsgEditOAuthClientResponse







<a name="mythos.websrv.v1.MsgRegisterOAuthClient"></a>

### MsgRegisterOAuthClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | bech32 address |
| `domain` | [string](#string) |  |  |






<a name="mythos.websrv.v1.MsgRegisterOAuthClientResponse"></a>

### MsgRegisterOAuthClientResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.websrv.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterOAuthClient` | [MsgRegisterOAuthClient](#mythos.websrv.v1.MsgRegisterOAuthClient) | [MsgRegisterOAuthClientResponse](#mythos.websrv.v1.MsgRegisterOAuthClientResponse) | Register OAuth client | |
| `EditOAuthClient` | [MsgEditOAuthClient](#mythos.websrv.v1.MsgEditOAuthClient) | [MsgEditOAuthClientResponse](#mythos.websrv.v1.MsgEditOAuthClientResponse) | Edit OAuth client | |
| `DeregisterOAuthClient` | [MsgDeregisterOAuthClient](#mythos.websrv.v1.MsgDeregisterOAuthClient) | [MsgDeregisterOAuthClientResponse](#mythos.websrv.v1.MsgDeregisterOAuthClientResponse) | Deregister OAuth client | |

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

