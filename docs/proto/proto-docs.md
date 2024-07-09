<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [mythos/network/v1/custom.proto](#mythos/network/v1/custom.proto)
    - [MsgExecuteAtomicTxRequest](#mythos.network.v1.MsgExecuteAtomicTxRequest)
  
- [mythos/network/v1/genesis.proto](#mythos/network/v1/genesis.proto)
    - [GenesisState](#mythos.network.v1.GenesisState)
  
- [mythos/network/v1/query.proto](#mythos/network/v1/query.proto)
    - [QueryContractCallRequest](#mythos.network.v1.QueryContractCallRequest)
    - [QueryContractCallResponse](#mythos.network.v1.QueryContractCallResponse)
    - [QueryMultiChainRequest](#mythos.network.v1.QueryMultiChainRequest)
    - [QueryMultiChainResponse](#mythos.network.v1.QueryMultiChainResponse)
  
    - [Query](#mythos.network.v1.Query)
  
- [mythos/network/v1/tendermint.proto](#mythos/network/v1/tendermint.proto)
    - [Event](#mythos.network.v1.Event)
    - [EventAttribute](#mythos.network.v1.EventAttribute)
    - [ExecTxResult](#mythos.network.v1.ExecTxResult)
    - [RequestBroadcastTx](#mythos.network.v1.RequestBroadcastTx)
    - [RequestPing](#mythos.network.v1.RequestPing)
    - [ResponseBroadcastTx](#mythos.network.v1.ResponseBroadcastTx)
    - [ResponseCheckTx](#mythos.network.v1.ResponseCheckTx)
    - [ResponsePing](#mythos.network.v1.ResponsePing)
  
    - [BroadcastAPI](#mythos.network.v1.BroadcastAPI)
  
- [mythos/network/v1/tx.proto](#mythos/network/v1/tx.proto)
    - [AtomicTxCrossChainCallInfo](#mythos.network.v1.AtomicTxCrossChainCallInfo)
    - [CrossChainCallInfo](#mythos.network.v1.CrossChainCallInfo)
    - [ExtensionOptionAtomicMultiChainTx](#mythos.network.v1.ExtensionOptionAtomicMultiChainTx)
    - [ExtensionOptionMultiChainTx](#mythos.network.v1.ExtensionOptionMultiChainTx)
    - [MsgExecuteAtomicTxResponse](#mythos.network.v1.MsgExecuteAtomicTxResponse)
    - [MsgExecuteContract](#mythos.network.v1.MsgExecuteContract)
    - [MsgExecuteContractResponse](#mythos.network.v1.MsgExecuteContractResponse)
    - [MsgExecuteCrossChainCallRequest](#mythos.network.v1.MsgExecuteCrossChainCallRequest)
    - [MsgExecuteCrossChainCallRequestIndexed](#mythos.network.v1.MsgExecuteCrossChainCallRequestIndexed)
    - [MsgExecuteCrossChainCallResponse](#mythos.network.v1.MsgExecuteCrossChainCallResponse)
    - [MsgExecuteCrossChainCallResponseIndexed](#mythos.network.v1.MsgExecuteCrossChainCallResponseIndexed)
    - [MsgGrpcReceiveRequest](#mythos.network.v1.MsgGrpcReceiveRequest)
    - [MsgGrpcReceiveRequestResponse](#mythos.network.v1.MsgGrpcReceiveRequestResponse)
    - [MsgGrpcSendRequest](#mythos.network.v1.MsgGrpcSendRequest)
    - [MsgGrpcSendRequestResponse](#mythos.network.v1.MsgGrpcSendRequestResponse)
    - [MsgMultiChainWrap](#mythos.network.v1.MsgMultiChainWrap)
    - [MsgMultiChainWrapResponse](#mythos.network.v1.MsgMultiChainWrapResponse)
    - [MsgP2PReceiveMessageRequest](#mythos.network.v1.MsgP2PReceiveMessageRequest)
    - [MsgP2PReceiveMessageResponse](#mythos.network.v1.MsgP2PReceiveMessageResponse)
    - [MsgQueryContract](#mythos.network.v1.MsgQueryContract)
    - [MsgQueryContractResponse](#mythos.network.v1.MsgQueryContractResponse)
    - [MsgStartBackgroundProcessRequest](#mythos.network.v1.MsgStartBackgroundProcessRequest)
    - [MsgStartBackgroundProcessResponse](#mythos.network.v1.MsgStartBackgroundProcessResponse)
    - [MsgStartTimeoutRequest](#mythos.network.v1.MsgStartTimeoutRequest)
    - [MsgStartTimeoutResponse](#mythos.network.v1.MsgStartTimeoutResponse)
    - [SubTxCrossChainCallInfo](#mythos.network.v1.SubTxCrossChainCallInfo)
    - [WrappedResponse](#mythos.network.v1.WrappedResponse)
  
    - [Msg](#mythos.network.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="mythos/network/v1/custom.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/custom.proto



<a name="mythos.network.v1.MsgExecuteAtomicTxRequest"></a>

### MsgExecuteAtomicTxRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `txs` | [bytes](#bytes) | repeated | protobuf encoded transactions |
| `sender` | [bytes](#bytes) |  |  |





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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="mythos/network/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/query.proto



<a name="mythos.network.v1.QueryContractCallRequest"></a>

### QueryContractCallRequest
QueryContractCallRequest is the request type for the
Query/ContractCall RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `multi_chain_id` | [string](#string) |  |  |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `address` | [string](#string) |  | Address is the address of the smart contract |
| `query_data` | [bytes](#bytes) |  |  |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |
| `dependencies` | [string](#string) | repeated | Array of either hex-encoded contract addresses or contract labels on which the execution of this message depends on |






<a name="mythos.network.v1.QueryContractCallResponse"></a>

### QueryContractCallResponse
QueryContractCallResponse is the response type for the
Query/ContractCall RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains the json data returned from the smart contract |






<a name="mythos.network.v1.QueryMultiChainRequest"></a>

### QueryMultiChainRequest
QueryMultiChainRequest is the request type for the
Query/QueryMultiChain RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `multi_chain_id` | [string](#string) |  |  |
| `query_data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.QueryMultiChainResponse"></a>

### QueryMultiChainResponse
QueryMultiChainResponse is the response type for the


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains the json data returned from the smart contract |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.network.v1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ContractCall` | [QueryContractCallRequest](#mythos.network.v1.QueryContractCallRequest) | [QueryContractCallResponse](#mythos.network.v1.QueryContractCallResponse) | ContractCall | GET|/network/v1/{multi_chain_id}/contract/{address}/call/{query_data}|
| `QueryMultiChain` | [QueryMultiChainRequest](#mythos.network.v1.QueryMultiChainRequest) | [QueryMultiChainResponse](#mythos.network.v1.QueryMultiChainResponse) |  | GET|/network/v1/{multi_chain_id}/data/{query_data}|

 <!-- end services -->



<a name="mythos/network/v1/tendermint.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/tendermint.proto



<a name="mythos.network.v1.Event"></a>

### Event
Event allows application developers to attach additional information to
ResponseFinalizeBlock and ResponseCheckTx.
Later, transactions may be queried using these events.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [string](#string) |  |  |
| `attributes` | [EventAttribute](#mythos.network.v1.EventAttribute) | repeated |  |






<a name="mythos.network.v1.EventAttribute"></a>

### EventAttribute
EventAttribute is a single key-value pair, associated with an event.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |
| `index` | [bool](#bool) |  | nondeterministic |






<a name="mythos.network.v1.ExecTxResult"></a>

### ExecTxResult
ExecTxResult contains results of executing one individual transaction.

* Its structure is equivalent to #ResponseDeliverTx which will be deprecated/deleted
tendermint.abci.ExecTxResult


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code` | [uint32](#uint32) |  |  |
| `data` | [bytes](#bytes) |  |  |
| `log` | [string](#string) |  | nondeterministic |
| `info` | [string](#string) |  | nondeterministic |
| `gas_wanted` | [int64](#int64) |  |  |
| `gas_used` | [int64](#int64) |  |  |
| `events` | [Event](#mythos.network.v1.Event) | repeated | nondeterministic |
| `codespace` | [string](#string) |  |  |






<a name="mythos.network.v1.RequestBroadcastTx"></a>

### RequestBroadcastTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.RequestPing"></a>

### RequestPing







<a name="mythos.network.v1.ResponseBroadcastTx"></a>

### ResponseBroadcastTx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `check_tx` | [ResponseCheckTx](#mythos.network.v1.ResponseCheckTx) |  |  |
| `tx_result` | [ExecTxResult](#mythos.network.v1.ExecTxResult) |  |  |






<a name="mythos.network.v1.ResponseCheckTx"></a>

### ResponseCheckTx
tendermint.abci.ResponseCheckTx


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code` | [uint32](#uint32) |  |  |
| `data` | [bytes](#bytes) |  |  |
| `log` | [string](#string) |  | nondeterministic |
| `info` | [string](#string) |  | nondeterministic |
| `gas_wanted` | [int64](#int64) |  |  |
| `gas_used` | [int64](#int64) |  |  |
| `events` | [Event](#mythos.network.v1.Event) | repeated |  |
| `codespace` | [string](#string) |  |  |






<a name="mythos.network.v1.ResponsePing"></a>

### ResponsePing






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.network.v1.BroadcastAPI"></a>

### BroadcastAPI
BroadcastAPI

Deprecated: This API will be superseded by a more comprehensive gRPC-based
broadcast API, and is scheduled for removal after v0.38.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Ping` | [RequestPing](#mythos.network.v1.RequestPing) | [ResponsePing](#mythos.network.v1.ResponsePing) |  | |
| `BroadcastTx` | [RequestBroadcastTx](#mythos.network.v1.RequestBroadcastTx) | [ResponseBroadcastTx](#mythos.network.v1.ResponseBroadcastTx) |  | |

 <!-- end services -->



<a name="mythos/network/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mythos/network/v1/tx.proto



<a name="mythos.network.v1.AtomicTxCrossChainCallInfo"></a>

### AtomicTxCrossChainCallInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `subtx` | [SubTxCrossChainCallInfo](#mythos.network.v1.SubTxCrossChainCallInfo) | repeated |  |






<a name="mythos.network.v1.CrossChainCallInfo"></a>

### CrossChainCallInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `request` | [MsgExecuteCrossChainCallRequest](#mythos.network.v1.MsgExecuteCrossChainCallRequest) |  |  |
| `response` | [WrappedResponse](#mythos.network.v1.WrappedResponse) |  |  |






<a name="mythos.network.v1.ExtensionOptionAtomicMultiChainTx"></a>

### ExtensionOptionAtomicMultiChainTx
ExtensionOptionMultiChainTx is an extension option for multichain atomic transactions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `leader_chain_id` | [string](#string) |  |  |
| `chain_ids` | [string](#string) | repeated |  |






<a name="mythos.network.v1.ExtensionOptionMultiChainTx"></a>

### ExtensionOptionMultiChainTx
ExtensionOptionMultiChainTx is an extension option for multichain atomic transactions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_id` | [string](#string) |  | option (gogoproto.goproto_getters) = false; |
| `index` | [int32](#int32) |  | index of this transaction in the atomic set |
| `tx_count` | [int32](#int32) |  | total transactions in the atomic set |






<a name="mythos.network.v1.MsgExecuteAtomicTxResponse"></a>

### MsgExecuteAtomicTxResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `results` | [ExecTxResult](#mythos.network.v1.ExecTxResult) | repeated |  |






<a name="mythos.network.v1.MsgExecuteContract"></a>

### MsgExecuteContract
ExecuteContract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract

Funds coins that are transferred to the contract on execution repeated cosmos.base.v1beta1.Coin funds = 4 [ (gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins" ]; // Array of either hex-encoded contract addresses or contract labels // on which the execution of this message depends on repeated string dependencies = 5; |






<a name="mythos.network.v1.MsgExecuteContractResponse"></a>

### MsgExecuteContractResponse
ExecuteContractResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgExecuteCrossChainCallRequest"></a>

### MsgExecuteCrossChainCallRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | sender is the contract that prepared the message |
| `from` | [string](#string) |  | from is the contract that sent the cross-chain message |
| `to` | [string](#string) |  | to is the address of the smart contract on the current chain |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |
| `dependencies` | [string](#string) | repeated | Array of either hex-encoded contract addresses or contract labels on which the execution of this message depends on |
| `from_chain_id` | [string](#string) |  |  |
| `to_chain_id` | [string](#string) |  |  |
| `is_query` | [bool](#bool) |  |  |






<a name="mythos.network.v1.MsgExecuteCrossChainCallRequestIndexed"></a>

### MsgExecuteCrossChainCallRequestIndexed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `request` | [MsgExecuteCrossChainCallRequest](#mythos.network.v1.MsgExecuteCrossChainCallRequest) |  |  |
| `index` | [int32](#int32) |  |  |






<a name="mythos.network.v1.MsgExecuteCrossChainCallResponse"></a>

### MsgExecuteCrossChainCallResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |
| `error` | [string](#string) |  |  |






<a name="mythos.network.v1.MsgExecuteCrossChainCallResponseIndexed"></a>

### MsgExecuteCrossChainCallResponseIndexed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [MsgExecuteCrossChainCallResponse](#mythos.network.v1.MsgExecuteCrossChainCallResponse) |  |  |
| `index` | [int32](#int32) |  |  |






<a name="mythos.network.v1.MsgGrpcReceiveRequest"></a>

### MsgGrpcReceiveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | this will always be the network module |
| `contract` | [string](#string) |  |  |
| `data` | [bytes](#bytes) |  |  |
| `encoding` | [string](#string) |  | evm, json, protobuf // ? |






<a name="mythos.network.v1.MsgGrpcReceiveRequestResponse"></a>

### MsgGrpcReceiveRequestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgGrpcSendRequest"></a>

### MsgGrpcSendRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | this will always be the network module |
| `contract` | [string](#string) |  |  |
| `ip_address` | [string](#string) |  |  |
| `data` | [bytes](#bytes) |  |  |
| `encoding` | [string](#string) |  | evm, json, protobuf // ? |






<a name="mythos.network.v1.MsgGrpcSendRequestResponse"></a>

### MsgGrpcSendRequestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgMultiChainWrap"></a>

### MsgMultiChainWrap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `multi_chain_id` | [string](#string) |  |  |
| `sender` | [string](#string) |  |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="mythos.network.v1.MsgMultiChainWrapResponse"></a>

### MsgMultiChainWrapResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgP2PReceiveMessageRequest"></a>

### MsgP2PReceiveMessageRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | this will always be the network module |
| `contract` | [string](#string) |  |  |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgP2PReceiveMessageResponse"></a>

### MsgP2PReceiveMessageResponse







<a name="mythos.network.v1.MsgQueryContract"></a>

### MsgQueryContract
QueryContract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Address is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Funds coins that are transferred to the contract on execution repeated cosmos.base.v1beta1.Coin funds = 4 [ (gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins" ]; // Array of either hex-encoded contract addresses or contract labels // on which the execution of this message depends on repeated string dependencies = 5; |






<a name="mythos.network.v1.MsgQueryContractResponse"></a>

### MsgQueryContractResponse
QueryContractResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgStartBackgroundProcessRequest"></a>

### MsgStartBackgroundProcessRequest
MsgStartBackgroundProcessRequest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | this will always be the wasmx module // TODO authority |
| `contract` | [string](#string) |  | contract address |
| `args` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgStartBackgroundProcessResponse"></a>

### MsgStartBackgroundProcessResponse
MsgStartBackgroundProcessResponse






<a name="mythos.network.v1.MsgStartTimeoutRequest"></a>

### MsgStartTimeoutRequest
MsgStartTimeoutRequest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | this will always be the wasmx module // TODO authority |
| `contract` | [string](#string) |  | contract address |
| `delay` | [int64](#int64) |  |  |
| `args` | [bytes](#bytes) |  |  |






<a name="mythos.network.v1.MsgStartTimeoutResponse"></a>

### MsgStartTimeoutResponse
MsgStartTimeoutResponse






<a name="mythos.network.v1.SubTxCrossChainCallInfo"></a>

### SubTxCrossChainCallInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `calls` | [CrossChainCallInfo](#mythos.network.v1.CrossChainCallInfo) | repeated |  |






<a name="mythos.network.v1.WrappedResponse"></a>

### WrappedResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |
| `error` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="mythos.network.v1.Msg"></a>

### Msg
Msg defines the grpc server

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GrpcSendRequest` | [MsgGrpcSendRequest](#mythos.network.v1.MsgGrpcSendRequest) | [MsgGrpcSendRequestResponse](#mythos.network.v1.MsgGrpcSendRequestResponse) | GrpcSendRequest | |
| `StartTimeout` | [MsgStartTimeoutRequest](#mythos.network.v1.MsgStartTimeoutRequest) | [MsgStartTimeoutResponse](#mythos.network.v1.MsgStartTimeoutResponse) | StartTimeout | |
| `StartBackgroundProcess` | [MsgStartBackgroundProcessRequest](#mythos.network.v1.MsgStartBackgroundProcessRequest) | [MsgStartBackgroundProcessResponse](#mythos.network.v1.MsgStartBackgroundProcessResponse) |  | |
| `MultiChainWrap` | [MsgMultiChainWrap](#mythos.network.v1.MsgMultiChainWrap) | [MsgMultiChainWrapResponse](#mythos.network.v1.MsgMultiChainWrapResponse) | MultiChainWrap wraps a message to be executed on one of the available chains | |
| `GrpcReceiveRequest` | [MsgGrpcReceiveRequest](#mythos.network.v1.MsgGrpcReceiveRequest) | [MsgGrpcReceiveRequestResponse](#mythos.network.v1.MsgGrpcReceiveRequestResponse) | GrpcReceiveRequest | |
| `P2PReceiveMessage` | [MsgP2PReceiveMessageRequest](#mythos.network.v1.MsgP2PReceiveMessageRequest) | [MsgP2PReceiveMessageResponse](#mythos.network.v1.MsgP2PReceiveMessageResponse) | P2PReceiveMessage | |
| `ExecuteAtomicTx` | [MsgExecuteAtomicTxRequest](#mythos.network.v1.MsgExecuteAtomicTxRequest) | [MsgExecuteAtomicTxResponse](#mythos.network.v1.MsgExecuteAtomicTxResponse) |  | |
| `ExecuteCrossChainTx` | [MsgExecuteCrossChainCallRequest](#mythos.network.v1.MsgExecuteCrossChainCallRequest) | [MsgExecuteCrossChainCallResponse](#mythos.network.v1.MsgExecuteCrossChainCallResponse) | only executed internally, sent by wasmx | |

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

