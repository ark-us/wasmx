package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	address "cosmossdk.io/core/address"
	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// RawContractMessage defines a json message that is sent or returned by a wasm contract.
// This type can hold any type of bytes. Until validateBasic is called there should not be
// any assumptions made that the data is valid syntax or semantic.
type RawContractMessage []byte

func (r *RawContractMessage) ValidateBasic() error {
	if r == nil {
		return ErrEmpty
	}
	if !json.Valid(*r) {
		return ErrInvalid
	}
	return nil
}

// Bytes returns raw bytes type
func (r RawContractMessage) Bytes() []byte {
	return r
}

// Equal content is equal json. Byte equal but this can change in the future.
func (r RawContractMessage) Equal(o RawContractMessage) bool {
	return bytes.Equal(r.Bytes(), o.Bytes())
}

func (msg MsgStoreCode) Route() string {
	return RouterKey
}

func (msg MsgStoreCode) Type() string {
	return "store-code"
}

func (msg MsgStoreCode) ValidateBasic() error {
	maxSize := GetMaxCodeSize(msg.Deps)
	if err := validateWasmCode(msg.ByteCode, maxSize); err != nil {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "code bytes %s", err.Error())
	}

	if msg.Metadata.Name == "" {
		msg.Metadata.Name = "unknown_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	return nil
}

func (msg MsgStoreCode) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return err
	}
	return nil
}

func (msg MsgDeployCode) Route() string {
	return RouterKey
}

func (msg MsgDeployCode) Type() string {
	return "deploy-code"
}

func (msg MsgDeployCode) ValidateBasic() error {
	maxSize := GetMaxCodeSize(msg.Deps)
	if err := validateCode(msg.ByteCode, maxSize); err != nil {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "code bytes %s", err.Error())
	}

	if !msg.Funds.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}

	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}

	return nil
}

func (msg MsgDeployCode) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return err
	}
	return nil
}

func (msg MsgInstantiateContract) Route() string {
	return RouterKey
}

func (msg MsgInstantiateContract) Type() string {
	return "instantiate"
}

func (msg MsgInstantiateContract) ValidateBasic() error {
	if msg.CodeId == 0 {
		return sdkerr.Wrap(sdkerrors.ErrInvalidRequest, "code id is required")
	}

	if err := ValidateLabel(msg.Label); err != nil {
		return sdkerr.Wrap(sdkerrors.ErrInvalidRequest, "label is required")
	}

	if !msg.Funds.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}

	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgInstantiateContract) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	return nil
}

func (msg MsgCompileContract) Route() string {
	return RouterKey
}

func (msg MsgCompileContract) Type() string {
	return "compile"
}

func (msg MsgCompileContract) ValidateBasic() error {
	if msg.CodeId == 0 {
		return sdkerr.Wrap(sdkerrors.ErrInvalidRequest, "code id is required")
	}
	return nil
}

func (msg MsgCompileContract) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Authority); err != nil {
		return sdkerr.Wrap(err, "authority")
	}
	return nil
}

func (msg MsgExecuteContract) Route() string {
	return RouterKey
}

func (msg MsgExecuteContract) Type() string {
	return "execute"
}

func (msg MsgExecuteContract) ValidateBasic() error {
	if !msg.Funds.IsValid() {
		return sdkerr.Wrap(sdkerrors.ErrInvalidCoins, "sentFunds")
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgExecuteContract) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := addressCodec.StringToBytes(msg.Contract)
	if err != nil {
		return sdkerr.Wrap(err, "contract")
	}
	if IsSystemAddress(contractAddress) {
		return sdkerr.Wrap(ErrUnauthorizedAddress, "cannot call system address")
	}
	return nil
}

func (msg MsgExecuteWithOriginContract) Route() string {
	return RouterKey
}

func (msg MsgExecuteWithOriginContract) Type() string {
	return "execute_with_origin"
}

func (msg MsgExecuteWithOriginContract) ValidateBasic() error {
	if !msg.Funds.IsValid() {
		return sdkerr.Wrap(sdkerrors.ErrInvalidCoins, "sentFunds")
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgExecuteWithOriginContract) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Origin); err != nil {
		return sdkerr.Wrap(err, "origin")
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := addressCodec.StringToBytes(msg.Contract)
	if err != nil {
		return sdkerr.Wrap(err, "contract")
	}
	if IsSystemAddress(contractAddress) {
		return sdkerr.Wrap(ErrUnauthorizedAddress, "cannot call system address")
	}
	return nil
}

func (msg MsgExecuteDelegateContract) Route() string {
	return RouterKey
}

func (msg MsgExecuteDelegateContract) Type() string {
	return "execute_delegate"
}

func (msg MsgExecuteDelegateContract) ValidateBasic() error {
	if !msg.Funds.IsValid() {
		return sdkerr.Wrap(sdkerrors.ErrInvalidCoins, "sentFunds")
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}
	return nil
}

func (msg MsgExecuteDelegateContract) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	codeAddress, err := addressCodec.StringToBytes(msg.CodeContract)
	if err != nil {
		return sdkerr.Wrap(err, "code_contract")
	}
	storageAddress, err := addressCodec.StringToBytes(msg.StorageContract)
	if err != nil {
		return sdkerr.Wrap(err, "storage_contract")
	}
	if IsSystemAddress(codeAddress) {
		return sdkerr.Wrap(ErrUnauthorizedAddress, "cannot call system address")
	}
	if IsSystemAddress(storageAddress) {
		return sdkerr.Wrap(ErrUnauthorizedAddress, "cannot call system address")
	}
	return nil
}

var _ sdk.Msg = &MsgInstantiateContract2{}

func (msg MsgInstantiateContract2) Route() string {
	return RouterKey
}

func (msg MsgInstantiateContract2) Type() string {
	return "instantiate2"
}

func (msg MsgInstantiateContract2) ValidateBasic() error {
	if msg.CodeId == 0 {
		return sdkerr.Wrap(sdkerrors.ErrInvalidRequest, "code id is required")
	}

	if err := ValidateLabel(msg.Label); err != nil {
		return sdkerr.Wrap(sdkerrors.ErrInvalidRequest, "label is required")
	}

	if !msg.Funds.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}

	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkerr.Wrap(err, "payload msg")
	}
	if err := ValidateSalt(msg.Salt); err != nil {
		return sdkerr.Wrap(err, "salt")
	}
	return nil
}

func (msg MsgInstantiateContract2) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	return nil
}

func (msg MsgExecuteEth) Route() string {
	return RouterKey
}

func (msg MsgExecuteEth) Type() string {
	return "execute-eth"
}

func (msg MsgExecuteEth) ValidateBasic() error {
	// TODO UnpackTxData, any.GetCachedValue
	// TODO validate tx arguments and signature
	return nil
}

func (msg MsgExecuteEth) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Sender); err != nil {
		return sdkerr.Wrap(err, "sender")
	}
	return nil
}

func (msg MsgExecuteEth) GetSignBytes() []byte {
	panic("MsgExecuteEth verifies ETH signature")
}

// AsTransaction creates an Ethereum Transaction type from the msg fields
func (msg MsgExecuteEth) AsTransaction() *ethtypes.Transaction {
	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil
	}
	return txData
}

// BuildTx builds the canonical cosmos tx from ethereum msg
func (msg MsgExecuteEth) BuildTx(txBuilder client.TxBuilder, evmDenom string) (signing.Tx, error) {
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errors.New("unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&ExtensionOptionEthereumTx{})
	if err != nil {
		return nil, err
	}

	txData, err := UnpackTxData(msg.Data)
	if err != nil {
		return nil, err
	}

	err = builder.SetMsgs(&msg)
	if err != nil {
		return nil, err
	}

	fees := make(sdk.Coins, 0)
	feeAmt := sdkmath.NewIntFromBigInt(msg.GetFee(txData.GasPrice(), txData.Gas()))
	if feeAmt.Sign() > 0 {
		fees = append(fees, sdk.NewCoin(evmDenom, feeAmt))
	}

	builder.SetExtensionOptions(option)
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(txData.Gas())
	tx := builder.GetTx()
	return tx, nil
}

func (msg MsgExecuteEth) GetFee(gasPrice *big.Int, gas uint64) *big.Int {
	gasLimit := new(big.Int).SetUint64(gas)
	return new(big.Int).Mul(gasPrice, gasLimit)
}

func (msg MsgExecuteEth) GetSignerFromSignature(ethSigner ethtypes.Signer) (sdk.AccAddress, error) {
	ethTx := msg.AsTransaction()
	sender, err := ethSigner.Sender(ethTx)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerrors.ErrorInvalidSigner,
			"couldn't retrieve sender address from the ethereum transaction: %s",
			err.Error(),
		)
	}
	return AccAddressFromEvm(sender), nil
}

// UnpackTxData unpacks an Any into a TxData. It returns an error if the
// client state can't be unpacked into a TxData.
func UnpackTxData(data []byte) (*ethtypes.Transaction, error) {
	if len(data) == 0 {
		return nil, sdkerr.Wrap(sdkerrors.ErrUnpackAny, "transaction data cannot be nil")
	}

	var tx ethtypes.Transaction
	err := tx.UnmarshalBinary(data)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack data into Transaction %T", data)
	}

	// txData, ok := any.GetCachedValue().(TxData)
	// if !ok {
	// 	return nil, sdkerr.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into TxData %T", any)
	// }

	return &tx, nil
}

// // PackTxData constructs a new Any packed with the given tx data value. It returns
// // an error if the client state can't be casted to a protobuf message or if the concrete
// // implementation is not registered to the protobuf codec.
// func PackTxData(txData TxData) (*codectypes.Any, error) {
// 	msg, ok := txData.(proto.Message)
// 	if !ok {
// 		return nil, errorsmod.Wrapf(errortypes.ErrPackAny, "cannot proto marshal %T", txData)
// 	}

// 	anyTxData, err := codectypes.NewAnyWithValue(msg)
// 	if err != nil {
// 		return nil, errorsmod.Wrap(errortypes.ErrPackAny, err.Error())
// 	}

// 	return anyTxData, nil
// }

func (msg MsgRegisterRole) Route() string {
	return RouterKey
}

func (msg MsgRegisterRole) Type() string {
	return "register-role"
}

func (msg MsgRegisterRole) ValidateBasic() error {
	if err := validateStringNonEmpty(msg.Title); err != nil {
		return sdkerr.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return sdkerr.Wrap(err, "description")
	}

	if err := validateStringNonEmpty(msg.Role); err != nil {
		return sdkerr.Wrap(err, "role")
	}

	if err := validateStringNonEmpty(msg.Label); err != nil {
		return sdkerr.Wrap(err, "label")
	}
	return nil
}

func (msg MsgRegisterRole) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Authority); err != nil {
		return sdkerr.Wrap(err, "authority")
	}
	if _, err := addressCodec.StringToBytes(msg.ContractAddress); err != nil {
		return sdkerr.Wrap(err, "contract address")
	}
	return nil
}

func (msg MsgDeregisterRole) Route() string {
	return RouterKey
}

func (msg MsgDeregisterRole) Type() string {
	return "deregister-role"
}

func (msg MsgDeregisterRole) ValidateBasic() error {
	if err := validateStringNonEmpty(msg.Title); err != nil {
		return sdkerr.Wrap(err, "title")
	}

	if err := validateStringNonEmpty(msg.Description); err != nil {
		return sdkerr.Wrap(err, "description")
	}
	return nil
}

func (msg MsgDeregisterRole) ValidateWithAddress(addressCodec address.Codec) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, err := addressCodec.StringToBytes(msg.Authority); err != nil {
		return sdkerr.Wrap(err, "authority")
	}
	if _, err := addressCodec.StringToBytes(msg.ContractAddress); err != nil {
		return sdkerr.Wrap(err, "contract address")
	}
	return nil
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateStringNonEmpty(i interface{}) error {
	if i == "" {
		return fmt.Errorf("empty string")
	}
	return validateString(i)
}
