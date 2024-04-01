package keeper

import (
	"context"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// QuerierSlashing is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierSlashing struct {
	Keeper *KeeperSlashing
}

var _ slashingtypes.QueryServer = QuerierSlashing{}

func NewQuerierSlashing(keeper *KeeperSlashing) QuerierSlashing {
	return QuerierSlashing{Keeper: keeper}
}

func (k QuerierSlashing) Params(goCtx context.Context, req *slashingtypes.QueryParamsRequest) (*slashingtypes.QueryParamsResponse, error) {
	return nil, nil
}

func (k QuerierSlashing) SigningInfo(goCtx context.Context, req *slashingtypes.QuerySigningInfoRequest) (*slashingtypes.QuerySigningInfoResponse, error) {
	return nil, nil
}

func (k QuerierSlashing) SigningInfos(goCtx context.Context, req *slashingtypes.QuerySigningInfosRequest) (*slashingtypes.QuerySigningInfosResponse, error) {
	return nil, nil
}
