package keeper

import (
	"context"

	metrics "github.com/hashicorp/go-metrics"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type msgBankServer struct {
	Keeper *KeeperBank
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgBankServerImpl(keeper *KeeperBank) banktypes.MsgServer {
	return &msgBankServer{
		Keeper: keeper,
	}
}

var _ banktypes.MsgServer = msgBankServer{}

func (m msgBankServer) Send(goCtx context.Context, msg *banktypes.MsgSend) (*banktypes.MsgSendResponse, error) {

	var (
		from, to []byte
		err      error
	)

	from, err = m.Keeper.ak.AddressCodec().StringToBytes(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	to, err = m.Keeper.ak.AddressCodec().StringToBytes(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	// if err := k.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
	// 	return nil, err
	// }

	// if k.BlockedAddr(to) {
	// 	return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	// }

	err = m.Keeper.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "send"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	return &banktypes.MsgSendResponse{}, nil
}

func (m msgBankServer) MultiSend(goCtx context.Context, msg *banktypes.MsgMultiSend) (*banktypes.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("msgBankServer.MultiSend not implemented")
	return &banktypes.MsgMultiSendResponse{}, nil
}

func (m msgBankServer) UpdateParams(goCtx context.Context, msg *banktypes.MsgUpdateParams) (*banktypes.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("msgBankServer.UpdateParams not implemented")
	return &banktypes.MsgUpdateParamsResponse{}, nil
}

func (m msgBankServer) SetSendEnabled(goCtx context.Context, msg *banktypes.MsgSetSendEnabled) (*banktypes.MsgSetSendEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("msgBankServer.SetSendEnabled not implemented")
	return &banktypes.MsgSetSendEnabledResponse{}, nil
}
