package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// QuerierBank is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierBank struct {
	Keeper *KeeperBank
}

var _ banktypes.QueryServer = QuerierBank{}

func NewQuerierBank(keeper *KeeperBank) QuerierBank {
	return QuerierBank{Keeper: keeper}
}

func (k QuerierBank) Balance(goCtx context.Context, req *banktypes.QueryBalanceRequest) (*banktypes.QueryBalanceResponse, error) {
	addr, err := k.Keeper.AccBech32Codec().StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, err
	}
	amount := k.Keeper.GetBalancePrefixed(goCtx, addr, req.Denom)
	return &banktypes.QueryBalanceResponse{Balance: &amount}, nil
}

func (k QuerierBank) AllBalances(goCtx context.Context, req *banktypes.QueryAllBalancesRequest) (*banktypes.QueryAllBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := k.Keeper.AccBech32Codec().StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := &banktypes.QueryAllBalancesRequest{
		Address:      addr.String(),
		Pagination:   nil,
		ResolveDenom: false,
	}

	bankmsgbz, err := k.Keeper.cdc.MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetAllBalances":%s}`, string(bankmsgbz)))
	resp, err := k.Keeper.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var contractResp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil, err
	}

	var response banktypes.QueryAllBalancesResponse
	err = k.Keeper.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (k QuerierBank) SpendableBalances(goCtx context.Context, req *banktypes.QuerySpendableBalancesRequest) (*banktypes.QuerySpendableBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.SpendableBalances not implemented")
	return &banktypes.QuerySpendableBalancesResponse{}, nil
}

func (k QuerierBank) SpendableBalanceByDenom(goCtx context.Context, req *banktypes.QuerySpendableBalanceByDenomRequest) (*banktypes.QuerySpendableBalanceByDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.SpendableBalanceByDenom not implemented")
	return &banktypes.QuerySpendableBalanceByDenomResponse{}, nil
}

func (k QuerierBank) TotalSupply(goCtx context.Context, req *banktypes.QueryTotalSupplyRequest) (*banktypes.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	reqbz, err := k.Keeper.cdc.MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetTotalSupply":%s}`, string(reqbz)))
	resp, err := k.Keeper.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var contractResp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil, err
	}

	var response banktypes.QueryTotalSupplyResponse
	err = k.Keeper.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (k QuerierBank) SupplyOf(goCtx context.Context, req *banktypes.QuerySupplyOfRequest) (*banktypes.QuerySupplyOfResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	reqbz, err := k.Keeper.cdc.MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetSupplyOf":%s}`, string(reqbz)))
	resp, err := k.Keeper.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var contractResp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil, err
	}

	var response banktypes.QuerySupplyOfResponse
	err = k.Keeper.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (k QuerierBank) Params(goCtx context.Context, req *banktypes.QueryParamsRequest) (*banktypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.Params not implemented")
	return &banktypes.QueryParamsResponse{}, nil
}

func (k QuerierBank) DenomMetadata(goCtx context.Context, req *banktypes.QueryDenomMetadataRequest) (*banktypes.QueryDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.DenomMetadata not implemented")
	return &banktypes.QueryDenomMetadataResponse{}, nil
}

func (k QuerierBank) DenomMetadataByQueryString(goCtx context.Context, req *banktypes.QueryDenomMetadataByQueryStringRequest) (*banktypes.QueryDenomMetadataByQueryStringResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.DenomMetadataByQueryString not implemented")
	return &banktypes.QueryDenomMetadataByQueryStringResponse{}, nil
}

func (k QuerierBank) DenomsMetadata(goCtx context.Context, req *banktypes.QueryDenomsMetadataRequest) (*banktypes.QueryDenomsMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.DenomsMetadata not implemented")
	return &banktypes.QueryDenomsMetadataResponse{}, nil
}

func (k QuerierBank) DenomOwners(goCtx context.Context, req *banktypes.QueryDenomOwnersRequest) (*banktypes.QueryDenomOwnersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.DenomOwners not implemented")
	return &banktypes.QueryDenomOwnersResponse{}, nil
}

func (k QuerierBank) DenomOwnersByQuery(goCtx context.Context, req *banktypes.QueryDenomOwnersByQueryRequest) (*banktypes.QueryDenomOwnersByQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.DenomOwnersByQuery not implemented")
	return &banktypes.QueryDenomOwnersByQueryResponse{}, nil
}

func (k QuerierBank) SendEnabled(goCtx context.Context, req *banktypes.QuerySendEnabledRequest) (*banktypes.QuerySendEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("QuerierBank.SendEnabled not implemented")
	return &banktypes.QuerySendEnabledResponse{}, nil
}
