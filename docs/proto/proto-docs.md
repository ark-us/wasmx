<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [wasmx/wasmx/contract.proto](#wasmx/wasmx/contract.proto)
    - [AbsoluteTxPosition](#wasmx.wasmx.AbsoluteTxPosition)
    - [CodeInfo](#wasmx.wasmx.CodeInfo)
    - [ContractInfo](#wasmx.wasmx.ContractInfo)
  
- [wasmx/wasmx/params.proto](#wasmx/wasmx/params.proto)
    - [Params](#wasmx.wasmx.Params)
  
- [wasmx/wasmx/genesis.proto](#wasmx/wasmx/genesis.proto)
    - [Code](#wasmx.wasmx.Code)
    - [Contract](#wasmx.wasmx.Contract)
    - [ContractStorage](#wasmx.wasmx.ContractStorage)
    - [GenesisState](#wasmx.wasmx.GenesisState)
    - [Sequence](#wasmx.wasmx.Sequence)
    - [SystemContract](#wasmx.wasmx.SystemContract)
  
- [wasmx/wasmx/query.proto](#wasmx/wasmx/query.proto)
    - [QueryParamsRequest](#wasmx.wasmx.QueryParamsRequest)
    - [QueryParamsResponse](#wasmx.wasmx.QueryParamsResponse)
  
    - [Query](#wasmx.wasmx.Query)
  
- [wasmx/wasmx/tx.proto](#wasmx/wasmx/tx.proto)
    - [MsgExecuteContract](#wasmx.wasmx.MsgExecuteContract)
    - [MsgExecuteContractResponse](#wasmx.wasmx.MsgExecuteContractResponse)
    - [MsgExecuteDelegateContract](#wasmx.wasmx.MsgExecuteDelegateContract)
    - [MsgExecuteDelegateContractResponse](#wasmx.wasmx.MsgExecuteDelegateContractResponse)
    - [MsgExecuteWithOriginContract](#wasmx.wasmx.MsgExecuteWithOriginContract)
    - [MsgInstantiateContract](#wasmx.wasmx.MsgInstantiateContract)
    - [MsgInstantiateContract2](#wasmx.wasmx.MsgInstantiateContract2)
    - [MsgInstantiateContract2Response](#wasmx.wasmx.MsgInstantiateContract2Response)
    - [MsgInstantiateContractResponse](#wasmx.wasmx.MsgInstantiateContractResponse)
    - [MsgStoreCode](#wasmx.wasmx.MsgStoreCode)
    - [MsgStoreCodeResponse](#wasmx.wasmx.MsgStoreCodeResponse)
  
    - [Msg](#wasmx.wasmx.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="wasmx/wasmx/contract.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/wasmx/contract.proto



<a name="wasmx.wasmx.AbsoluteTxPosition"></a>

### AbsoluteTxPosition
AbsoluteTxPosition is a unique transaction position that allows for global
ordering of transactions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | BlockHeight is the block the contract was created at |
| `tx_index` | [uint64](#uint64) |  | TxIndex is a monotonic counter within the block (actual transaction index, or gas consumed) |






<a name="wasmx.wasmx.CodeInfo"></a>

### CodeInfo
CodeInfo is data for the uploaded contract WASM code


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_hash` | [bytes](#bytes) |  | CodeHash is the unique identifier created by hashing the wasm code |
| `creator` | [string](#string) |  | Creator address who initially stored the code |
| `deps` | [string](#string) | repeated | deps can be hex-formatted contract addresses (32 bytes) or versioned interface labels |
| `abi` | [string](#string) |  |  |
| `json_schema` | [string](#string) |  |  |






<a name="wasmx.wasmx.ContractInfo"></a>

### ContractInfo
ContractInfo stores a WASM contract instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored Wasm code |
| `creator` | [string](#string) |  | Creator address who initially instantiated the contract |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `init_message` | [bytes](#bytes) |  | Initialization message |
| `created` | [AbsoluteTxPosition](#wasmx.wasmx.AbsoluteTxPosition) |  | Created Tx position when the contract was instantiated. |
| `ibc_port_id` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/wasmx/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/wasmx/params.proto



<a name="wasmx.wasmx.Params"></a>

### Params
Params defines the parameters for the module.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/wasmx/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/wasmx/genesis.proto



<a name="wasmx.wasmx.Code"></a>

### Code
Code - for importing and exporting code data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  |  |
| `code_info` | [CodeInfo](#wasmx.wasmx.CodeInfo) |  |  |
| `code_bytes` | [bytes](#bytes) |  |  |
| `pinned` | [bool](#bool) |  | Pinned to wasmvm cache |






<a name="wasmx.wasmx.Contract"></a>

### Contract
Contract struct encompasses ContractAddress, ContractInfo, and ContractState


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |
| `contract_info` | [ContractInfo](#wasmx.wasmx.ContractInfo) |  |  |
| `contract_state` | [ContractStorage](#wasmx.wasmx.ContractStorage) | repeated |  |






<a name="wasmx.wasmx.ContractStorage"></a>

### ContractStorage
ContractStorage


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  | hex-encode key |
| `value` | [bytes](#bytes) |  | raw value |






<a name="wasmx.wasmx.GenesisState"></a>

### GenesisState
GenesisState defines the wasmx module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.wasmx.Params) |  |  |
| `bootstrap_account_address` | [string](#string) |  | bootstrap address |
| `contract` | [SystemContract](#wasmx.wasmx.SystemContract) | repeated |  |
| `codes` | [Code](#wasmx.wasmx.Code) | repeated |  |
| `contracts` | [Contract](#wasmx.wasmx.Contract) | repeated |  |
| `sequences` | [Sequence](#wasmx.wasmx.Sequence) | repeated |  |






<a name="wasmx.wasmx.Sequence"></a>

### Sequence
Sequence key and value of an id generation counter


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id_key` | [bytes](#bytes) |  |  |
| `value` | [uint64](#uint64) |  |  |






<a name="wasmx.wasmx.SystemContract"></a>

### SystemContract



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `label` | [string](#string) |  |  |
| `init_message` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="wasmx/wasmx/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/wasmx/query.proto



<a name="wasmx.wasmx.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="wasmx.wasmx.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#wasmx.wasmx.Params) |  | params holds all the parameters of this module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.wasmx.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#wasmx.wasmx.QueryParamsRequest) | [QueryParamsResponse](#wasmx.wasmx.QueryParamsResponse) | Parameters queries the parameters of the module. | GET|/wasmx/wasmx/params|

 <!-- end services -->



<a name="wasmx/wasmx/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasmx/wasmx/tx.proto



<a name="wasmx.wasmx.MsgExecuteContract"></a>

### MsgExecuteContract
MsgExecuteContract submits the given message data to a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |






<a name="wasmx.wasmx.MsgExecuteContractResponse"></a>

### MsgExecuteContractResponse
MsgExecuteContractResponse returns execution result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="wasmx.wasmx.MsgExecuteDelegateContract"></a>

### MsgExecuteDelegateContract
MsgExecuteDelegateContract submits the given message data to a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `origin` | [string](#string) |  | Origin is the actor that originally signed the message |
| `sender` | [string](#string) |  | Sender is the storage contract, equivalent to the address that triggers the message (signer) |
| `caller` | [string](#string) |  | Caller is the address that will be used as sender |
| `code_contract` | [string](#string) |  | CodeContract is the address of the smart contract whose binary is used |
| `storage_contract` | [string](#string) |  | StorageContract is the address of the smart contract whose storage is used |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |






<a name="wasmx.wasmx.MsgExecuteDelegateContractResponse"></a>

### MsgExecuteDelegateContractResponse
MsgExecuteDelegateContractResponse returns execution result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="wasmx.wasmx.MsgExecuteWithOriginContract"></a>

### MsgExecuteWithOriginContract
MsgExecuteContract submits the given message data to a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `origin` | [string](#string) |  | Origin is the actor that originally signed the message |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |






<a name="wasmx.wasmx.MsgInstantiateContract"></a>

### MsgInstantiateContract
MsgInstantiateContract create a new smart contract instance for the given
code id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |






<a name="wasmx.wasmx.MsgInstantiateContract2"></a>

### MsgInstantiateContract2
MsgInstantiateContract2 create a new smart contract instance for the given
code id with a predicable address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `code_id` | [uint64](#uint64) |  | Admin is an optional address that can execute migrations |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |
| `salt` | [bytes](#bytes) |  | Salt is an arbitrary value provided by the sender. Size can be 1 to 64. |
| `fix_msg` | [bool](#bool) |  | FixMsg include the msg value into the hash for the predictable address. Default is false |






<a name="wasmx.wasmx.MsgInstantiateContract2Response"></a>

### MsgInstantiateContract2Response
MsgInstantiateContract2Response return instantiation result data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address of the new contract instance. |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="wasmx.wasmx.MsgInstantiateContractResponse"></a>

### MsgInstantiateContractResponse
MsgInstantiateContractResponse return instantiation result data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address of the new contract instance. |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="wasmx.wasmx.MsgStoreCode"></a>

### MsgStoreCode
MsgStoreCode submit Wasm code to the system


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `wasm_byte_code` | [bytes](#bytes) |  | WASMByteCode can be raw or gzip compressed |






<a name="wasmx.wasmx.MsgStoreCodeResponse"></a>

### MsgStoreCodeResponse
MsgStoreCodeResponse returns store result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `checksum` | [bytes](#bytes) |  | Checksum is the sha256 hash of the stored code |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="wasmx.wasmx.Msg"></a>

### Msg
Msg defines the wasm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `StoreCode` | [MsgStoreCode](#wasmx.wasmx.MsgStoreCode) | [MsgStoreCodeResponse](#wasmx.wasmx.MsgStoreCodeResponse) | StoreCode to submit Wasm code to the system | |
| `InstantiateContract` | [MsgInstantiateContract](#wasmx.wasmx.MsgInstantiateContract) | [MsgInstantiateContractResponse](#wasmx.wasmx.MsgInstantiateContractResponse) | InstantiateContract creates a new smart contract instance for the given code id. | |
| `InstantiateContract2` | [MsgInstantiateContract2](#wasmx.wasmx.MsgInstantiateContract2) | [MsgInstantiateContract2Response](#wasmx.wasmx.MsgInstantiateContract2Response) | InstantiateContract2 creates a new smart contract instance for the given code id with a predictable address | |
| `ExecuteContract` | [MsgExecuteContract](#wasmx.wasmx.MsgExecuteContract) | [MsgExecuteContractResponse](#wasmx.wasmx.MsgExecuteContractResponse) | Execute submits the given message data to a smart contract | |
| `ExecuteWithOriginContract` | [MsgExecuteWithOriginContract](#wasmx.wasmx.MsgExecuteWithOriginContract) | [MsgExecuteContractResponse](#wasmx.wasmx.MsgExecuteContractResponse) | ExecuteWithOrigin submits the given message data to a smart contract | |
| `ExecuteDelegateContract` | [MsgExecuteDelegateContract](#wasmx.wasmx.MsgExecuteDelegateContract) | [MsgExecuteDelegateContractResponse](#wasmx.wasmx.MsgExecuteDelegateContractResponse) | ExecuteDelegate submits the given message data to a smart contract | |

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

