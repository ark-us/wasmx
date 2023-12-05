package keeper

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/cometbft/cometbft/node"

	sdkerr "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type IntervalAction struct {
	Delay  int64
	Repeat int32
	Args   []byte
}

type msgServer struct {
	app           types.BaseApp
	intervalCount int32
	intervals     map[int32]*IntervalAction
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper, app types.BaseApp) types.MsgServer {
	return &msgServer{
		Keeper:        keeper,
		app:           app,
		intervalCount: 0,
		intervals:     map[int32]*IntervalAction{},
	}
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServer(
	keeper Keeper,
	tmNode *node.Node,
) types.MsgServer {
	return &msgServer{
		Keeper:        keeper,
		intervalCount: 0,
		intervals:     map[int32]*IntervalAction{},
	}
}

var _ types.MsgServer = msgServer{}

func (m *msgServer) incrementIntervalId() {
	m.intervalCount += 1
}

// TODO this must not be called from outside, only from wasmx... (authority)
// maybe only from the contract that the interval is for?
func (m msgServer) StartInterval(goCtx context.Context, msg *types.MsgStartIntervalRequest) (*types.MsgStartIntervalResponse, error) {
	fmt.Println("Go - start interval request", msg.Contract, string(msg.Args))
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	intervalId := m.intervalCount
	fmt.Println("Go - intervalId", intervalId)
	m.incrementIntervalId()

	fmt.Println("Go - intervalCount", m.intervalCount)

	m.intervals[intervalId] = &IntervalAction{
		Delay:  msg.Delay,
		Repeat: msg.Repeat,
		Args:   msg.Args,
		// Cancel: cancelFn, // just through a stop error TODO
	}
	fmt.Println("Go - post interval set")

	description := fmt.Sprintf("timed action: id %d, delay %dms, repeat %d, args: %s ", intervalId, msg.Delay, msg.Repeat, string(msg.Args))
	_intervalId := big.NewInt(int64(intervalId))
	data := append(_intervalId.FillBytes(make([]byte, 4)), msg.Args...)
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	msgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	fmt.Println("--StartInterval--", intervalId, msg.Delay, msg.Repeat, string(msg.Args))

	timeDelay := msg.Delay
	httpSrvDone := make(chan struct{}, 1)
	height := ctx.BlockHeight() - 1 // TODO fixme
	fmt.Println("---height--", height)
	logger := m.Keeper.Logger(ctx)
	// errCh := make(chan error)
	m.goRoutineGroup.Go(func() error {
		fmt.Println("--StartInterval--set new ctx")

		// ctx := ht.req.Context()
		// var cancel context.CancelFunc
		// if ht.timeoutSet {
		// 	ctx, cancel = context.WithTimeout(ctx, ht.timeout)
		// } else {
		// 	ctx, cancel = context.WithCancel(ctx)
		// }

		// set context
		// grpcCtx := goCtx
		sdkCtx_, ctxcachems, err := CreateQueryContext(m.app, logger, height, false)
		fmt.Println("--StartInterval--CreateQueryContext", err)
		if err != nil {
			logger.Error("failed to create query context", "err", err)
			return err
		}
		sdkCtx, commitCacheCtx := sdkCtx_.CacheContext()

		// Add relevant gRPC headers
		if height == 0 {
			height = sdkCtx.BlockHeight() // If height was not set in the request, set it to the latest
		}

		// Attach the sdk.Context into the gRPC's context.Context.
		// grpcCtx = context.WithValue(grpcCtx, sdk.SdkContextKey, sdkCtx)

		// md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))
		// if err = grpc.SetHeader(grpcCtx, md); err != nil {
		// 	logger.Error("failed to set gRPC header", "err", err)
		// }

		// NOW execute
		fmt.Println("--StartInterval--sleeping...")

		time.Sleep(time.Duration(timeDelay) * time.Millisecond)
		fmt.Println("--StartInterval--ExecuteEventual...")

		goCtx := context.Background()
		goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)
		ctx := sdk.UnwrapSDKContext(goCtx)

		_, err = m.Keeper.wasmxKeeper.ExecuteEventual(ctx, contractAddress, contractAddress, msgbz, make([]string, 0))
		fmt.Println("--StartInterval--ExecuteEventual", err)
		if err != nil {
			// TODO - stop without propagating a stop to parent
			if err == types.ErrGoroutineClosed {
				m.Logger(ctx).Error("Closing thread", "description", description, err.Error())
				close(httpSrvDone)
				return nil
			}

			m.Logger(ctx).Error("failed to start a new thread", "error", err.Error())
			// errCh <- err
			return err
		}

		commitCacheCtx()
		// commit original context
		mythosapp, ok := m.app.(MythosApp)
		if !ok {
			return fmt.Errorf("failed to get MythosApp from server Application")
		}
		origtstore := ctxcachems.GetStore(mythosapp.GetMKey(wasmxtypes.MemStoreKey))
		origtstore.(storetypes.CacheWrap).Write()

		// <-ctx.Done() ?
		// svrCtx.Logger.Info("stopping the thread...")
		return nil
	})

	return &types.MsgStartIntervalResponse{
		IntervalId: intervalId,
	}, nil
}

func (m msgServer) GrpcSendRequest(goCtx context.Context, msg *types.MsgGrpcSendRequest) (*types.MsgGrpcSendRequestResponse, error) {
	fmt.Println("Go - send grpc request", msg.IpAddress, msg.Data, string(msg.Data))
	ctx := sdk.UnwrapSDKContext(goCtx)

	ip := msg.IpAddress
	client := StartGRPCClient(ip)
	req := &types.MsgGrpcReceiveRequest{
		Data:     msg.Data,
		Sender:   msg.Sender,
		Contract: msg.Contract,
		Encoding: msg.Encoding,
	}
	res, err := client.GrpcReceiveRequest(ctx, req)
	fmt.Println("Go - grpc request sent", err)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrpcSendRequestResponse{
		Data: res.Data,
	}, nil
}

func (m msgServer) GrpcReceiveRequest(goCtx context.Context, msg *types.MsgGrpcReceiveRequest) (*types.MsgGrpcReceiveRequestResponse, error) {
	fmt.Println("Go - received grpc request", msg.Data, string(msg.Data))
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	datab64 := base64.StdEncoding.EncodeToString(msg.Data)
	data := []byte(fmt.Sprintf(`{"run":{"event":{"type":"receiveHeartbeat","params":[{"key":"entry","value":"%s"}]}}}`, datab64))
	fmt.Println("===receiveHeartbeat==", string(data))
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-GrpcReceiveRequest-network-execmsgbz--", hex.EncodeToString(execmsgbz))
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	fmt.Println("Go - received grpc request and executed state machine", string(resp), err)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-GrpcReceiveRequest-network-resp---", string(resp))

	// test state
	qmsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"getCurrentState":{}}`)}
	qmsgbz, err := json.Marshal(qmsg)
	if err != nil {
		return nil, err
	}
	qres, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, qmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Go - current state query: ", string(qres))
	var qresbz wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(qres, &qresbz)
	if err != nil {
		return nil, err
	}

	fmt.Println("Go - current state: ", string(qresbz.Data))

	// logs_count0
	qmsg = wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"getContextValue":{"key":"logs_count"}}`)}
	qmsgbz, err = json.Marshal(qmsg)
	if err != nil {
		return nil, err
	}
	qres, err = m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, qmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Go - logs count: ", string(qres))
	err = json.Unmarshal(qres, &qresbz)
	if err != nil {
		return nil, err
	}
	fmt.Println("Go - logs count: ", string(qresbz.Data))

	return &types.MsgGrpcReceiveRequestResponse{
		Data: resp,
	}, nil
}

func (m msgServer) Setup(goCtx context.Context, msg *types.MsgSetup) (*types.MsgSetupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "ExecuteEth could not parse sender address")
	}

	data := []byte(`{"create":{"context":[{"key":"data","value":"aGVsbG8="},{"key":"address","value":"0.0.0.0:8091"}],"id":"AB-Req-Res-timer","initial":"uninitialized","states":[{"name":"uninitialized","after":[],"on":[{"name":"initialize","target":"active","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"active","after":[],"on":[{"name":"receiveRequest","target":"received","guard":"","actions":[]},{"name":"send","target":"sender","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"received","after":[{"name":"1000","target":"#AB-Req-Res-timer.active","guard":"","actions":[]}],"on":[],"exit":[],"entry":[],"initial":"","states":[]},{"name":"sender","after":[{"name":"10000","target":"#AB-Req-Res-timer.sending","guard":"","actions":[{"type":"xstate.raise","event":{"type":"sendRequest","params":[{"key":"data","value":"data"},{"key":"address","value":"address"}]},"params":[]}]}],"on":[],"exit":[],"entry":[],"initial":"","states":[]},{"name":"sending","after":[],"on":[{"name":"sendRequest","target":"sender","guard":"","actions":[{"type":"sendRequest","params":[]}]}],"exit":[],"entry":[],"initial":"","states":[]}]}}`)

	var resp []byte
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	resp, err = m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetupResponse{
		Data: string(resp),
	}, nil
}

func (m msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-Ping--network-resp---", string(resp))

	return &types.MsgExecuteContractResponse{
		Data: resp,
	}, nil
}

func (m msgServer) QueryContract(goCtx context.Context, msg *types.MsgQueryContract) (*types.MsgQueryContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: msg.Msg}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}

	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, senderAddr, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-Ping--network-resp---", string(resp))

	return &types.MsgQueryContractResponse{
		Data: resp,
	}, nil
}

func (m msgServer) Ping(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}

	data := []byte(`{"run":{"event":{"type":"send","params":[]}}}`)
	execmsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	execmsgbz, err := json.Marshal(execmsg)
	if err != nil {
		return nil, err
	}
	// fmt.Println("--Ping--network-execmsgbz--", hex.EncodeToString(execmsgbz))
	resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	// fmt.Println("-Ping--network-resp---", string(resp))

	// test state
	qmsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"getCurrentState":{}}`)}
	qmsgbz, err := json.Marshal(qmsg)
	if err != nil {
		return nil, err
	}
	qres, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, qmsgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Go - Ping current state query: ", string(qres))
	var qresbz wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(qres, &qresbz)
	if err != nil {
		return nil, err
	}

	fmt.Println("Go - Ping current state: ", string(qresbz.Data))

	response := msg.Data + hex.EncodeToString(resp)

	return &types.MsgPingResponse{
		Data: response,
	}, nil
}

// func (m msgServer) Ping2(goCtx context.Context, msg *types.MsgPing) (*types.MsgPingResponse, error) {
// 	fmt.Println("---------Ping", msg.Data, goCtx)
// 	ctx := sdk.UnwrapSDKContext(goCtx)
// 	fmt.Println("---------Ping ctx", ctx)

// 	// fmt.Println("---------Ping validators", m.GetValidators(ctx))

// 	tmNode := m.TmNode
// 	fmt.Println("==Ping=peers===", tmNode.ConsensusReactor().Switch.Peers())
// 	fmt.Println("==Ping=ProposerAddress===", tmNode.BlockStore().LoadBaseMeta().Header.ProposerAddress)

// 	fmt.Println("==Validators.GetProposer()===", tmNode.EvidencePool().State().Validators.GetProposer())
// 	fmt.Println("==NextValidators.GetProposer()===", tmNode.EvidencePool().State().NextValidators.GetProposer())

// 	fmt.Println("==Validators.Validators()===", tmNode.EvidencePool().State().Validators.Validators)

// 	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")

// 	bz, err := hex.DecodeString("0000000000000000000000000000000000000005")
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println("--network-bz--", bz)
// 	execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
// 	execmsgbz, err := json.Marshal(execmsg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
// 	resp, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println("-network-resp---", resp)

// 	response := msg.Data + hex.EncodeToString(resp)

// 	return &types.MsgPingResponse{
// 		Data: response,
// 	}, nil
// }

func (m msgServer) SetValidators(goCtx context.Context, msg *types.MsgSetValidators) (*types.MsgSetValidatorsResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// // fmt.Println("==SetValidators===")

	// tmNode := m.TmNode
	// var validators []*cmttypes.Validator
	// validators = tmNode.EvidencePool().State().Validators.Validators
	// fmt.Println("=SetValidators=Validators.Validators()===", len(validators), validators)

	// validatorAddresses := make([]sdk.AccAddress, len(validators))
	// for i, valid := range validators {
	// 	fmt.Println("---validatorAddresses---", i, valid)
	// 	validatorAddresses[i] = sdk.AccAddress(valid.Address)
	// 	fmt.Println("---validatorAddresses---", i, validatorAddresses[i].String(), hex.EncodeToString(validatorAddresses[i]))
	// }
	// fmt.Println("---validatorAddresses---", validatorAddresses)

	// // validatorAddresses := []sdk.AccAddress{
	// // 	wasmxtypes.AccAddressFromHex("1111111111111111111111111111111111111111"),
	// // 	wasmxtypes.AccAddressFromHex("2222222222222222222222222222222222222222"),
	// // }

	// contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "contract")
	// }
	// datalen := big.NewInt(int64(len(validatorAddresses))).FillBytes(make([]byte, 32))
	// bz, err := hex.DecodeString("9300c9260000000000000000000000000000000000000000000000000000000000000020")
	// if err != nil {
	// 	return nil, err
	// }
	// bz = append(bz, datalen...)

	// for _, valid := range validatorAddresses {
	// 	// fmt.Println("--SetValidators-bz-0-", hex.EncodeToString(bz))
	// 	bz = append(bz, make([]byte, 12)...)
	// 	bz = append(bz, valid.Bytes()...)
	// 	// fmt.Println("--SetValidators-bz-1-", hex.EncodeToString(bz))
	// }
	// // fmt.Println("--SetValidators-bz--", hex.EncodeToString(bz))

	// execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	// execmsgbz, err := json.Marshal(execmsg)
	// if err != nil {
	// 	return nil, err
	// }
	// // fmt.Println("--SetValidators-execmsgbz--", hex.EncodeToString(execmsgbz))
	// // TODO have authority network + governance for these contracts
	// // TODO sender must be network module
	// // sender := sdk.AccAddress("network") // must have 20 bytes
	// sender := contractAddress
	// _, err = m.wasmxKeeper.Execute(ctx, contractAddress, sender, execmsgbz, nil, nil)
	// // fmt.Println("-SetValidators-resp---", resp, err)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgSetValidatorsResponse{}, nil
}

func (m msgServer) GetValidators(goCtx context.Context, msg *types.MsgGetValidators) (*types.MsgGetValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// fmt.Println("==GetValidators===")

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
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

func (m msgServer) MakeProposal(goCtx context.Context, msg *types.MsgMakeProposal) (*types.MsgMakeProposalResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// tmNode := m.TmNode
	// currentValidator, err := tmNode.PrivValidator().GetPubKey()
	// if err != nil {
	// 	return nil, err
	// }
	// // tmNode.NodeInfo().ID()
	// // tmNode.Switch().NetAddress()
	// // tmNode.Switch().

	// fmt.Println("==currentValidator", currentValidator.Address())

	// contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "contract")
	// }
	// bz, err := hex.DecodeString("589f5dd70000000000000000000000000000000000000000000000000000000000000040" + hex.EncodeToString(currentValidator.Address()) + "000000000000000000000000000000000000000000000000000000000000000568656c6c6f000000000000000000000000000000000000000000000000000000")
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("--network-bz--", hex.EncodeToString(bz))
	// execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	// execmsgbz, err := json.Marshal(execmsg)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
	// resp, err := m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("-network-resp---", resp)

	return &types.MsgMakeProposalResponse{
		Data: "",
	}, nil
}

func (m msgServer) IsProposer(goCtx context.Context, msg *types.MsgIsProposer) (*types.MsgIsProposerResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// tmNode := m.TmNode
	// currentValidator, err := tmNode.PrivValidator().GetPubKey()
	// if err != nil {
	// 	return nil, err
	// }
	// // tmNode.NodeInfo().ID()
	// // tmNode.Switch().NetAddress()
	// // tmNode.Switch().

	// fmt.Println("==currentValidator", currentValidator.Address())

	// contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "contract")
	// }
	// bz, err := hex.DecodeString("e9790d02")
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("--network-bz--", bz)
	// execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	// execmsgbz, err := json.Marshal(execmsg)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("--network-execmsgbz--", hex.EncodeToString(execmsgbz))
	// resp, err := m.wasmxKeeper.Query(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("-network-resp---", resp)

	// return &types.MsgIsProposerResponse{
	// 	IsProposer: hex.EncodeToString(resp) == "0000000000000000000000000000000000000001",
	// }, nil

	return &types.MsgIsProposerResponse{
		IsProposer: false,
	}, nil
}

func (m msgServer) SetCurrentNode(goCtx context.Context, msg *types.MsgSetCurrentNode) (*types.MsgSetCurrentNodeResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// tmNode := m.TmNode
	// currentValidator, err := tmNode.PrivValidator().GetPubKey()
	// if err != nil {
	// 	return nil, err
	// }

	// contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "contract")
	// }
	// bz, err := hex.DecodeString("9a25709f000000000000000000000000" + hex.EncodeToString(currentValidator.Address()))
	// if err != nil {
	// 	return nil, err
	// }
	// execmsg := wasmxtypes.WasmxExecutionMessage{Data: bz}
	// execmsgbz, err := json.Marshal(execmsg)
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = m.wasmxKeeper.Execute(ctx, contractAddress, contractAddress, execmsgbz, nil, nil)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgSetCurrentNodeResponse{}, nil
}

func (m msgServer) GetCurrentNode(goCtx context.Context, msg *types.MsgGetCurrentNode) (*types.MsgGetCurrentNodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddress, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}
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
