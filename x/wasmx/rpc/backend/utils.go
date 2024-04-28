package backend

import (
	"encoding/hex"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/log"

	address "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	abci "github.com/cometbft/cometbft/abci/types"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"

	rpctypes "mythos/v1/x/wasmx/rpc/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// GetLogsFromBlockResults returns the list of event logs from the tendermint block result response
func GetLogsFromBlockResults(addressCodec address.Codec, blockRes *tmrpctypes.ResultBlockResults) ([][]*ethtypes.Log, error) {
	blockLogs := [][]*ethtypes.Log{}
	for _, txResult := range blockRes.TxsResults {
		logs, err := TxLogsFromEvents(addressCodec, txResult.Events)
		if err != nil {
			return nil, err
		}

		blockLogs = append(blockLogs, logs)
	}
	return blockLogs, nil
}

// TxLogsFromEvents parses ethereum logs from cosmos events for specific msg index
func TxLogsFromEvents(addressCodec address.Codec, events []abci.Event) ([]*ethtypes.Log, error) {
	var logs []*ethtypes.Log
	for _, event := range events {
		if event.Type != (wasmxtypes.CustomContractEventPrefix + wasmxtypes.EventTypeWasmxLog) {
			continue
		}
		log, err := ParseTxLogsFromEvent(addressCodec, event)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// TODO only ewasm logs?
// ParseTxLogsFromEvent parse tx logs from one event
func ParseTxLogsFromEvent(addressCodec address.Codec, event abci.Event) (*ethtypes.Log, error) {
	var log ethtypes.Log
	for _, attr := range event.Attributes {
		// we now parse all wasmx logs, regardless of AttributeKeyEventType
		if attr.Key == wasmxtypes.AttributeKeyIndex {
			index, err := strconv.Atoi(string(attr.Value))
			if err != nil {
				return nil, err
			}
			log.Index = uint(index)
		} else if attr.Key == wasmxtypes.AttributeKeyContractAddr {
			contractAddress, err := addressCodec.StringToBytes(string(attr.Value))
			if err != nil {
				return nil, err
			}
			log.Address = wasmxtypes.EvmAddressFromAcc(contractAddress)
		} else if attr.Key == wasmxtypes.AttributeKeyTopic {
			log.Topics = append(log.Topics, common.HexToHash(string(attr.Value)))
		} else if attr.Key == wasmxtypes.AttributeKeyData {
			data, err := hex.DecodeString(string(attr.Value[2:]))
			if err != nil {
				return nil, err
			}
			log.Data = data
		}
	}
	return &log, nil
}

// ContractAddressFromEvents returns a deployed contract address from cosmos events
func ContractAddressFromEvents(addressCodec address.Codec, events []abci.Event) *common.Address {
	for _, event := range events {
		if event.Type != wasmxtypes.EventTypeDeploy {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key == wasmxtypes.AttributeKeyContractAddr {
				contractAddress, err := addressCodec.StringToBytes(string(attr.Value))
				if err != nil {
					return nil
				}
				ethAddress := wasmxtypes.EvmAddressFromAcc(contractAddress)
				return &ethAddress
			}
		}
	}
	return nil
}

// getAccountNonce returns the account nonce for the given account address.
// If the pending value is true, it will iterate over the mempool (pending)
// txs in order to compute and return the pending tx sequence.
func (b *Backend) getAccountNonce(accAddr common.Address, pending bool, height int64, logger log.Logger) (uint64, error) {
	queryClient := authtypes.NewQueryClient(b.clientCtx)
	addrCodec := b.clientCtx.InterfaceRegistry.SigningContext().AddressCodec()
	adr, err := addrCodec.BytesToString(sdk.AccAddress(accAddr.Bytes()))
	if err != nil {
		return 0, err
	}
	ctx := rpctypes.ContextWithHeight(height)
	res, err := queryClient.Account(ctx, &authtypes.QueryAccountRequest{Address: adr})
	if err != nil {
		st, ok := status.FromError(err)
		// treat as account doesn't exist yet
		if ok && st.Code() == codes.NotFound {
			return 0, nil
		}
		return 0, err
	}
	var acc authtypes.AccountI
	if err := b.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return 0, err
	}

	nonce := acc.GetSequence()

	if !pending {
		return nonce, nil
	}

	// the account retriever doesn't include the uncommitted transactions on the nonce so we need to
	// to manually add them.
	pendingTxs, _, err := b.PendingTransactions()
	if err != nil {
		logger.Error("failed to fetch pending transactions", "error", err.Error())
		return nonce, nil
	}

	nonce += uint64(len(pendingTxs))

	return nonce, nil
}
