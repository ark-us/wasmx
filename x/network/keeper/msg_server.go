package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/node"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgServer struct {
	Keeper
	DB        dbm.DB
	ClientCtx client.Context
	TmNode    *node.Node
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper, db dbm.DB, clientCtx client.Context) types.MsgServer {
	return &msgServer{Keeper: keeper, DB: db, ClientCtx: clientCtx}
}

var _ types.MsgServer = msgServer{}

// This is the entrypoint for transactions signed by Ethereum wallets
// Works with both EVM & CosmWasm contracts, both interpreted and wasm-based
func (m msgServer) Ping(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
	fmt.Println("---------Ping", msg.Data, goCtx)
	ctx := sdk.UnwrapSDKContext(goCtx)
	fmt.Println("---------Ping ctx", ctx)

	fmt.Println("---------Ping validators", m.GetValidators(ctx))

	tmNode := m.TmNode
	fmt.Println("==startCmtNode=peers===", tmNode.ConsensusReactor().Switch.Peers())
	fmt.Println("==startCmtNode=ProposerAddress===", tmNode.BlockStore().LoadBaseMeta().Header.ProposerAddress)

	fmt.Println("==Validators.GetProposer()===", tmNode.EvidencePool().State().Validators.GetProposer())
	fmt.Println("==NextValidators.GetProposer()===", tmNode.EvidencePool().State().NextValidators.GetProposer())

	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")

	bz, err := hex.DecodeString("0000000000000000000000000000000000000005")
	fmt.Println("--network-bz--", bz)
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-execmsgbz--", execmsgbz)
	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("-network-resp---", resp)

	response := msg.Data + hex.EncodeToString(resp)

	return &types.MsgPingResponse{
		Data: response,
	}, nil
}
