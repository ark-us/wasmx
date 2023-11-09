package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cometbft/cometbft/node"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var NETWORK_HEX_ADDRESS = "0x0000000000000000000000000000000000000028"

type msgServer struct {
	Keeper
	ClientCtx client.Context
	TmNode    *node.Node
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper, clientCtx client.Context) types.MsgServer {
	return &msgServer{Keeper: keeper, ClientCtx: clientCtx}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) Ping(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
	fmt.Println("---------Ping", msg.Data, goCtx)
	ctx := sdk.UnwrapSDKContext(goCtx)
	fmt.Println("---------Ping ctx", ctx)

	// fmt.Println("---------Ping validators", m.GetValidators(ctx))

	tmNode := m.TmNode
	fmt.Println("==Ping=peers===", tmNode.ConsensusReactor().Switch.Peers())
	fmt.Println("==Ping=ProposerAddress===", tmNode.BlockStore().LoadBaseMeta().Header.ProposerAddress)

	fmt.Println("==Validators.GetProposer()===", tmNode.EvidencePool().State().Validators.GetProposer())
	fmt.Println("==NextValidators.GetProposer()===", tmNode.EvidencePool().State().NextValidators.GetProposer())

	fmt.Println("==Validators.Validators()===", tmNode.EvidencePool().State().Validators.Validators)

	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")

	bz, err := hex.DecodeString("0000000000000000000000000000000000000005")
	fmt.Println("--network-bz--", bz)
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
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

func (m msgServer) SetValidators(goCtx context.Context, msg *types.MsgSetValidators) (*types.MsgSetValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// fmt.Println("==SetValidators===")

	tmNode := m.TmNode
	var validators []*cmttypes.Validator
	validators = tmNode.EvidencePool().State().Validators.Validators
	fmt.Println("=SetValidators=Validators.Validators()===", len(validators), validators)

	validatorAddresses := make([]sdk.AccAddress, len(validators))
	for i, valid := range validators {
		fmt.Println("---validatorAddresses---", i, valid)
		validatorAddresses[i] = sdk.AccAddress(valid.Address)
		fmt.Println("---validatorAddresses---", i, validatorAddresses[i].String(), hex.EncodeToString(validatorAddresses[i]))
	}
	fmt.Println("---validatorAddresses---", validatorAddresses)

	// validatorAddresses := []sdk.AccAddress{
	// 	wasmxtypes.AccAddressFromHex("1111111111111111111111111111111111111111"),
	// 	wasmxtypes.AccAddressFromHex("2222222222222222222222222222222222222222"),
	// }

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	datalen := big.NewInt(int64(len(validatorAddresses))).FillBytes(make([]byte, 32))
	bz, err := hex.DecodeString("9300c9260000000000000000000000000000000000000000000000000000000000000020")
	if err != nil {
		return nil, err
	}
	bz = append(bz, datalen...)

	for _, valid := range validatorAddresses {
		// fmt.Println("--SetValidators-bz-0-", hex.EncodeToString(bz))
		bz = append(bz, make([]byte, 12)...)
		bz = append(bz, valid.Bytes()...)
		// fmt.Println("--SetValidators-bz-1-", hex.EncodeToString(bz))
	}
	// fmt.Println("--SetValidators-bz--", hex.EncodeToString(bz))

	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	// fmt.Println("--SetValidators-execmsgbz--", hex.EncodeToString(execmsgbz))
	// TODO have authority network + governance for these contracts
	// TODO sender must be network module
	// sender := sdk.AccAddress("network") // must have 20 bytes
	sender := contractAddress
	_, err = m.wasmxKeeper.Execute(ctx, contractAddress, sender, execmsgbz, nil, nil)
	// fmt.Println("-SetValidators-resp---", resp, err)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetValidatorsResponse{}, nil
}

func (m msgServer) GetValidators(goCtx context.Context, msg *types.MsgGetValidators) (*types.MsgGetValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// fmt.Println("==GetValidators===")

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	bz, err := hex.DecodeString("b7ab4db5")
	if err != nil {
		return nil, err
	}

	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	// fmt.Println("--GetValidators-execmsgbz--", hex.EncodeToString(execmsgbz))
	// TODO have authority network + governance for these contracts
	// TODO sender must be network module
	// sender := sdk.AccAddress("network") // must have 20 bytes
	sender := contractAddress
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, sender, execmsgbz, nil, nil)
	// fmt.Println("-GetValidators-resp---", resp, err)
	if err != nil {
		return nil, err
	}

	return &types.MsgGetValidatorsResponse{
		Validators: []string{hex.EncodeToString(resp)},
	}, nil
}

// 0x0000000000000000000000000000000000000028
// setValidators(address[] memory _validators)
// getProposer()
// makeProposal(string memory message, address currentNode)

func (m msgServer) MakeProposal(goCtx context.Context, msg *types.MsgMakeProposal) (*types.MsgMakeProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tmNode := m.TmNode
	currentValidator, err := tmNode.PrivValidator().GetPubKey()
	if err != nil {
		return nil, err
	}
	// tmNode.NodeInfo().ID()
	// tmNode.Switch().NetAddress()
	// tmNode.Switch().

	fmt.Println("==currentValidator", currentValidator.Address())

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	bz, err := hex.DecodeString("589f5dd70000000000000000000000000000000000000000000000000000000000000040" + hex.EncodeToString(currentValidator.Address()) + "000000000000000000000000000000000000000000000000000000000000000568656c6c6f000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-bz--", hex.EncodeToString(bz))
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("-network-resp---", resp)

	return &types.MsgMakeProposalResponse{
		Data: "",
	}, nil
}

func (m msgServer) IsProposer(goCtx context.Context, msg *types.MsgIsProposer) (*types.MsgIsProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tmNode := m.TmNode
	currentValidator, err := tmNode.PrivValidator().GetPubKey()
	if err != nil {
		return nil, err
	}
	// tmNode.NodeInfo().ID()
	// tmNode.Switch().NetAddress()
	// tmNode.Switch().

	fmt.Println("==currentValidator", currentValidator.Address())

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	bz, err := hex.DecodeString("e9790d02")
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-bz--", bz)
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("-network-resp---", resp)

	return &types.MsgIsProposerResponse{
		IsProposer: hex.EncodeToString(resp) == "0000000000000000000000000000000000000001",
	}, nil
}

func (m msgServer) SetCurrentNode(goCtx context.Context, msg *types.MsgSetCurrentNode) (*types.MsgSetCurrentNodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tmNode := m.TmNode
	currentValidator, err := tmNode.PrivValidator().GetPubKey()
	if err != nil {
		return nil, err
	}

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	bz, err := hex.DecodeString("9a25709f000000000000000000000000" + hex.EncodeToString(currentValidator.Address()))
	if err != nil {
		return nil, err
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	_, err = m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetCurrentNodeResponse{}, nil
}

func (m msgServer) GetCurrentNode(goCtx context.Context, msg *types.MsgGetCurrentNode) (*types.MsgGetCurrentNodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress := wasmxtypes.AccAddressFromHex(NETWORK_HEX_ADDRESS)
	bz, err := hex.DecodeString("14f26bc3")
	if err != nil {
		return nil, err
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgGetCurrentNodeResponse{CurrentNode: hex.EncodeToString(resp)}, nil
}
