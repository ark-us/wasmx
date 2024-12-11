package multichain

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var FlagRegistryChainId = "registry-chain-id"
var CHAIN_CONFIG_REGISTRY = wasmxtypes.ROLE_METAREGISTRY

var EMPTY_BYTES20 = bytes.Repeat([]byte{0}, 20)

func AddMultiChainFlagsToCmd(cmd *cobra.Command) {
	f := cmd.Flags()
	f.String(FlagRegistryChainId, "", "multichain registry chain id")
}

type CliMultiChainCtx struct {
	ClientCtx       client.Context
	CustomAddrCodec mcodec.AccBech32Codec
	Config          *menc.ChainConfig
	SubchainId      string
	RegistryId      string
	// CtxChainId      string
	ForwardedTx bool
}

func MultiChainCtxByChainIdWithAppMsgs(
	clientCtx client.Context,
	flags *flag.FlagSet,
	customSigners []signing.CustomGetSigner,
	appCreatorFactory NewAppCreator,
) (*CliMultiChainCtx, error) {
	subchainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return &CliMultiChainCtx{}, fmt.Errorf("subchainId: %s", err)
	}
	registryId, err := flags.GetString(FlagRegistryChainId)
	if err != nil {
		return &CliMultiChainCtx{}, fmt.Errorf("registry chainId: %s", err)
	}
	clientCtx, customAddrCodec, config, err := MultiChainCtx(clientCtx, customSigners, subchainId, registryId)
	_, appCreator := CreateNoLoggerAppCreatorTemp(appCreatorFactory, 0)
	isubchainapp := appCreator(subchainId, config)
	subchainapp := isubchainapp.(AppwithTxConfig)

	clientCtx = clientCtx.
		WithCodec(subchainapp.AppCodec()).
		WithInterfaceRegistry(subchainapp.InterfaceRegistry()).
		WithTxConfig(subchainapp.TxConfig()).
		WithLegacyAmino(subchainapp.LegacyAmino())

	return &CliMultiChainCtx{
		ClientCtx:       clientCtx,
		CustomAddrCodec: customAddrCodec,
		Config:          config,
		SubchainId:      subchainId,
		RegistryId:      registryId,
		ForwardedTx:     false,
	}, err
}

func MultiChainCtxByChainId(clientCtx client.Context, flags *flag.FlagSet, customSigners []signing.CustomGetSigner) (*CliMultiChainCtx, error) {
	subchainId, err := flags.GetString(sdkflags.FlagChainID)
	if err != nil {
		return &CliMultiChainCtx{}, fmt.Errorf("subchainId: %s", err)
	}
	registryId, err := flags.GetString(FlagRegistryChainId)
	if err != nil {
		return &CliMultiChainCtx{}, fmt.Errorf("registry chainId: %s", err)
	}
	clientCtx, customAddrCodec, config, err := MultiChainCtx(clientCtx, customSigners, subchainId, registryId)

	return &CliMultiChainCtx{
		ClientCtx:       clientCtx,
		CustomAddrCodec: customAddrCodec,
		Config:          config,
		SubchainId:      subchainId,
		RegistryId:      registryId,
		// ForwardedTx: subchainId != ctxChainId,
		ForwardedTx: false,
	}, err
}

func MultiChainCtx(clientCtx client.Context, customSigners []signing.CustomGetSigner, chainId string, registryId string) (client.Context, mcodec.AccBech32Codec, *menc.ChainConfig, error) {
	config, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		if _, err := mcfg.GetChainConfig(registryId); err != nil {
			return clientCtx, mcodec.AccBech32Codec{}, nil, fmt.Errorf("could not get registry chain config: %s", err.Error())
		}
		config, err = GetSubChainConfig(clientCtx, chainId, registryId)
		if err != nil {
			return clientCtx, mcodec.AccBech32Codec{}, config, fmt.Errorf("could not get chain config by registry id: %s", err.Error())
		}
	}
	mcfg.SetGlobalChainConfig(chainId, *config)
	customEncoding := menc.MakeEncodingConfig(config, customSigners)

	customCdc := mcodec.NewAccBech32Codec(config.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	customAddrCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)

	clientCtx = clientCtx.WithAccountRetriever(cosmosmodtypes.NewAccountRetriever(chainId, customAddrCodec))

	clientCtx = clientCtx.
		WithCodec(customEncoding.Marshaler).
		WithInterfaceRegistry(customEncoding.InterfaceRegistry).
		WithTxConfig(customEncoding.TxConfig).
		WithLegacyAmino(customEncoding.Amino)

	return clientCtx, customAddrCodec, config, nil
}

func MultiChainWrap(clientCtx client.Context, msg sdk.Msg, fromAddr mcodec.AccAddressPrefixed) (*networktypes.MsgMultiChainWrap, error) {
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	msgMultiChain := &networktypes.MsgMultiChainWrap{
		MultiChainId: clientCtx.ChainID,
		Sender:       fromAddr.String(),
		Data:         msgAny,
	}
	return msgMultiChain, nil
}

func (ctx *CliMultiChainCtx) MultiChainWrap(msg sdk.Msg, fromAddr mcodec.AccAddressPrefixed) (sdk.Msg, error) {
	if ctx.ForwardedTx {
		return MultiChainWrap(ctx.ClientCtx, msg, fromAddr)
	}
	return msg, nil
}

func GetSubChainConfig(clientCtx client.Context, subchainId string, registryId string) (*menc.ChainConfig, error) {
	chainConfigBz, err := GetSubChainConfigBz(clientCtx, subchainId, registryId)
	if err != nil {
		return nil, err
	}
	var chainConfig menc.ChainConfig
	err = json.Unmarshal(chainConfigBz, &chainConfig)
	if err != nil {
		return nil, err
	}
	return &chainConfig, nil
}

func GetSubChainConfigBz(clientCtx client.Context, subchainId string, registryId string) ([]byte, error) {
	addrhex := wasmxtypes.ADDR_METAREGISTRY
	if CHAIN_CONFIG_REGISTRY == wasmxtypes.ROLE_MULTICHAIN_REGISTRY {
		addrhex = wasmxtypes.ADDR_MULTICHAIN_REGISTRY
	}
	registryAddress := wasmxtypes.AccAddressFromHex(addrhex)

	clientCtx, customAddrCodec, _, err := MultiChainCtx(clientCtx, []signing.CustomGetSigner{}, registryId, "")
	if err != nil {
		return nil, fmt.Errorf("cannot get registry config: %s", err.Error())
	}

	contractAddr, err := customAddrCodec.BytesToString(registryAddress)
	if err != nil {
		return nil, err
	}
	querymsg, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{GetSubChainConfigById: &wasmxtypes.QueryGetSubChainRequest{
		ChainId: subchainId,
	}})
	if err != nil {
		return nil, err
	}

	res, err := sendMultiChainQuery(
		clientCtx,
		customAddrCodec,
		contractAddr,
		EMPTY_BYTES20,
		querymsg,
		nil,
		nil,
		registryId,
	)
	if err != nil {
		return nil, err
	}
	data2, err := decodeQueryResponse(res.Data)
	if err != nil {
		return nil, fmt.Errorf("could not decode config query response: %s: %s", hex.EncodeToString(res.Data), err.Error())
	}
	return data2, nil
}

func decodeQueryResponse(resp []byte) ([]byte, error) {
	var data wasmxtypes.QuerySmartContractCallResponse
	err := data.Unmarshal(resp)
	if err != nil {
		return nil, fmt.Errorf("cannot decode query response, data: %s", hex.EncodeToString(resp))
	}
	var data2 wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(data.Data, &data2)
	if err != nil {
		return nil, err
	}
	return data2.Data, nil
}

func sendMultiChainQuery(
	clientCtx client.Context,
	addrCodec address.Codec,
	contractAddr string, sender sdk.AccAddress,
	queryData []byte, amount sdk.Coins,
	dependencies []string,
	chainId string,
) (*networktypes.QueryMultiChainResponse, error) {
	if !json.Valid(queryData) {
		return nil, errors.New("query data must be json")
	}
	msg := wasmxtypes.WasmxExecutionMessage{Data: queryData}
	msgbz, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	senderAddr, err := addrCodec.BytesToString(sender)
	if err != nil {
		return nil, err
	}
	queryClient := networktypes.NewQueryClient(clientCtx)

	query := wasmxtypes.QuerySmartContractCallRequest{
		Sender:       senderAddr,
		Address:      contractAddr,
		QueryData:    msgbz,
		Funds:        amount,
		Dependencies: dependencies,
	}
	bz, err := query.Marshal()
	if err != nil {
		return nil, err
	}

	req := &abci.RequestQuery{Data: bz, Path: "/mythos.wasmx.v1.Query/SmartContractCall"}
	reqbz, err := req.Marshal()
	if err != nil {
		return nil, err
	}

	res, err := queryClient.QueryMultiChain(
		context.Background(),
		&networktypes.QueryMultiChainRequest{
			MultiChainId: chainId,
			QueryData:    reqbz,
		},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}
