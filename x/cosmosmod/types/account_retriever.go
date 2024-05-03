package types

import (
	"context"
	"fmt"
	"strconv"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "mythos/v1/codec"
	networktypes "mythos/v1/x/network/types"
)

var (
	_ client.Account          = sdk.AccountI(nil)
	_ client.AccountRetriever = AccountRetriever{}
)

// AccountRetriever defines the properties of a type that can be used to
// retrieve accounts.
type AccountRetriever struct {
	ChainId      string
	AddressCodec mcodec.AccBech32Codec
}

func NewAccountRetriever(chainId string, addrCodec mcodec.AccBech32Codec) AccountRetriever {
	return AccountRetriever{ChainId: chainId, AddressCodec: addrCodec}
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ar AccountRetriever) GetAccount(clientCtx client.Context, addr sdk.AccAddress) (client.Account, error) {
	account, _, err := ar.GetAccountWithHeight(clientCtx, addr)
	return account, err
}

// GetAccountWithHeight queries for an account given an address. Returns the
// height of the query with the account. An error is returned if the query
// or decoding fails.
func (ar AccountRetriever) GetAccountWithHeight(clientCtx client.Context, addr sdk.AccAddress) (client.Account, int64, error) {
	var header metadata.MD
	addrPrefixed := ar.AddressCodec.BytesToAccAddressPrefixed(addr)
	fmt.Println("---GetAccountWithHeight--", addrPrefixed.String())
	queryClient := networktypes.NewQueryClient(clientCtx)

	accreq := &authtypes.QueryAccountRequest{Address: addrPrefixed.String()}

	// accreqAny, err := codectypes.NewAnyWithValue(accreq)
	// fmt.Println("---NewAnyWithValue--", err, accreqAny)
	// fmt.Println("---NewAnyWithValue somevAny TypeUrl--", accreqAny.TypeUrl)
	// fmt.Println("---NewAnyWithValue somevAny Value--", string(accreqAny.Value))
	// accreqAnyBz, err := clientCtx.Codec.Marshal(accreqAny)
	// if err != nil {
	// 	return nil, 0, err
	// }

	accreqbz, err := accreq.Marshal()
	if err != nil {
		return nil, 0, err
	}

	typeUrl := sdk.MsgTypeURL(accreq)
	queryPath, err := mcodec.GetPathFromType(typeUrl)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("---GetAccountWithHeight queryPath--", queryPath)

	abciQuery := &abci.RequestQuery{Data: accreqbz, Path: queryPath}
	abciQueryBz, err := abciQuery.Marshal()
	if err != nil {
		return nil, 0, err
	}

	req := &networktypes.QueryMultiChainRequest{
		MultiChainId: ar.ChainId,
		// QueryData:    accreqAnyBz,
		QueryData: abciQueryBz,
	}
	fmt.Println("---GetAccountWithHeight sending QueryMultiChain--")
	res, err := queryClient.QueryMultiChain(context.Background(), req, grpc.Header(&header))
	fmt.Println("---GetAccountWithHeight res--", err, res)
	if err != nil {
		return nil, 0, err
	}

	fmt.Println("---GetAccountWithHeight account--", err, string(res.Data))

	// anyRes, err := mcodec.AnyFromBz(clientCtx.Codec, res.Data)
	// fmt.Println("---GetAccountWithHeight account--", err, anyRes)
	// if err != nil {
	// 	return nil, 0, err
	// }

	var resAcc authtypes.QueryAccountResponse
	err = resAcc.Unmarshal(res.Data)
	fmt.Println("---GetAccountWithHeight resAcc(1)--", err, resAcc)

	// err = clientCtx.Codec.Unmarshal(res.Data, &resAcc)
	// fmt.Println("---GetAccountWithHeight resAcc(2)--", err, resAcc)

	if err != nil {
		return nil, 0, err
	}
	fmt.Println("---GetAccountWithHeight resAcc.Account--", resAcc.Account)

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	if l := len(blockHeight); l != 1 {
		return nil, 0, fmt.Errorf("unexpected '%s' header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, l, 1)
	}

	nBlockHeight, err := strconv.Atoi(blockHeight[0])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	// var acc sdk.AccountI
	// var acc mcodec.AccountI

	var acc2 BaseAccount
	err = clientCtx.Codec.UnmarshalInterface(resAcc.Account.Value, &acc2)
	fmt.Println("---GetAccountWithHeight resAcc.Account 1--", err, acc2)

	err = acc2.Unmarshal(resAcc.Account.Value)
	fmt.Println("---GetAccountWithHeight resAcc.Account 11--", err, acc2)

	// err = clientCtx.InterfaceRegistry.UnpackAny(resAcc.Account, &acc)
	// fmt.Println("---GetAccountWithHeight resAcc.Account 2--", err, acc)
	if err != nil {
		return nil, 0, err
	}

	return acc2.ToCosmosAccountI(), int64(nBlockHeight), nil
}

// GetAccountWithHeight queries for an account given an address. Returns the
// height of the query with the account. An error is returned if the query
// or decoding fails.
func (ar AccountRetriever) GetAccountWithHeight2(clientCtx client.Context, addr sdk.AccAddress) (client.Account, int64, error) {
	var header metadata.MD
	addrPrefixed := ar.AddressCodec.BytesToAccAddressPrefixed(addr)
	fmt.Println("---GetAccountWithHeight--", addrPrefixed.String())
	queryClient := authtypes.NewQueryClient(clientCtx)

	res, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: addrPrefixed.String()}, grpc.Header(&header))
	fmt.Println("---GetAccountWithHeight account--", err, res)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("---GetAccountWithHeight account--", res.Account)

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	if l := len(blockHeight); l != 1 {
		return nil, 0, fmt.Errorf("unexpected '%s' header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, l, 1)
	}

	nBlockHeight, err := strconv.Atoi(blockHeight[0])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	var acc sdk.AccountI
	if err := clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, 0, err
	}

	return acc, int64(nBlockHeight), nil
}

// EnsureExists returns an error if no account exists for the given address else nil.
func (ar AccountRetriever) EnsureExists(clientCtx client.Context, addr sdk.AccAddress) error {
	if _, err := ar.GetAccount(clientCtx, addr); err != nil {
		return err
	}

	return nil
}

// GetAccountNumberSequence returns sequence and account number for the given address.
// It returns an error if the account couldn't be retrieved from the state.
func (ar AccountRetriever) GetAccountNumberSequence(clientCtx client.Context, addr sdk.AccAddress) (uint64, uint64, error) {
	acc, err := ar.GetAccount(clientCtx, addr)
	if err != nil {
		return 0, 0, err
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
