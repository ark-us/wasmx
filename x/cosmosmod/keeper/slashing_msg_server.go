package keeper

import (
	"context"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type msgSlashingServer struct {
	Keeper *KeeperSlashing
}

// NewMsgSlashingServerImpl returns an implementation of the MsgServer interface
func NewMsgSlashingServerImpl(keeper *KeeperSlashing) slashingtypes.MsgServer {
	return &msgSlashingServer{
		Keeper: keeper,
	}
}

var _ slashingtypes.MsgServer = msgSlashingServer{}

func (m msgSlashingServer) Unjail(goCtx context.Context, msg *slashingtypes.MsgUnjail) (*slashingtypes.MsgUnjailResponse, error) {
	return nil, nil
}

func (m msgSlashingServer) UpdateParams(goCtx context.Context, msg *slashingtypes.MsgUpdateParams) (*slashingtypes.MsgUpdateParamsResponse, error) {
	return nil, nil
}
