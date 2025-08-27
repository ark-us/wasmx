package vm

import (
	"encoding/hex"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	errorsmod "cosmossdk.io/errors"
	cometbftenc "github.com/cometbft/cometbft/crypto/encoding"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	mctx "github.com/loredanacirstea/wasmx/context"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type ResponseOptimisticExecution struct {
	MetaInfo map[string][]byte `json:"metainfo"`
}

type WrapRequestFinalizeBlock struct {
	Request  abci.RequestFinalizeBlock `json:"request"`
	MetaInfo map[string][]byte         `json:"metainfo"`
}

type WrapResult struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

// PrepareProposal(*abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
func PrepareProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}

	var req abci.RequestPrepareProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	resp, err := ctx.GetApplication().PrepareProposal(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "PrepareProposal")
		return nil, err
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

// ProcessProposal(*abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
func ProcessProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req abci.RequestProcessProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ProcessProposal")
		return nil, err
	}
	bapp := ctx.GetApplication()
	resp, err := bapp.ProcessProposal(&req)

	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ProcessProposal")
		return nil, err
	}

	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func OptimisticExecution(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, ndx := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	resptr, ndx := memc.GetPointerFromParams(rnh, params, ndx)
	resbz, err := rnh.ReadMemFromPtr(resptr)
	if err != nil {
		return nil, err
	}
	var req abci.RequestProcessProposal
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, err
	}
	var res abci.ResponseProcessProposal
	err = json.Unmarshal(resbz, &res)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, err
	}
	bapp := ctx.GetApplication()
	oe := bapp.GetOptimisticExecution()
	oe.Enable()

	// reset meta info from previous optimistic executions
	mctx.ResetExecutionMetaInfo(ctx.GoContextParent)

	bapp.OptimisticExecution(&req, &res)
	oe.Disable()

	// TODO we should return the error, not throw
	_, err = oe.WaitResult()
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, err
	}
	metainfo, err := mctx.GetExecutionMetaInfoEncoded(ctx.GoContextParent, ctx.GetCosmosHandler().Codec())
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "OptimisticExecution")
		return nil, err
	}

	respbz, err := json.Marshal(&ResponseOptimisticExecution{MetaInfo: metainfo})
	if err != nil {
		return nil, err
	}

	return rnh.AllocateWriteMem(respbz)
}

// FinalizeBlock(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
func FinalizeBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req WrapRequestFinalizeBlock
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	// set metainfo on the parent context, so it is available during execution
	err = mctx.SetExecutionMetaInfo(ctx.GoContextParent, ctx.CosmosHandler.Codec(), req.MetaInfo)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		return nil, err
	}

	bapp := ctx.GetApplication()
	resp, err := bapp.FinalizeBlockSimple(&req.Request)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "FinalizeBlock")
		errmsg = err.Error()
	}
	oe := bapp.GetOptimisticExecution()
	if oe.Initialized() {
		oe.Reset()
	}

	// hooks for finalize block result
	err = ctx.CosmosHandler.FinalizeBlockResultHandler(ctx.Ctx, resp)
	if err != nil {
		return nil, err
	}

	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respwrapbz)
}

// BeginBlock(*abci.RequestFinalizeBlock) (sdk.BeginBlock, error)
func BeginBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req abci.RequestFinalizeBlock
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	resp, err := ctx.GetApplication().BeginBlock(&req)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "BeginBlock")
		errmsg = err.Error()
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respwrapbz)
}

// EndBlock(metadata string) (*abci.ResponseFinalizeBlock, error)
func EndBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	metadata, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	resp, err := ctx.GetApplication().EndBlock(metadata)
	errmsg := ""
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "EndBlock")
		errmsg = err.Error()
	}

	// hooks for finalize block result, e.g. system cache hooks
	err = ctx.CosmosHandler.EndBlockResultHandler(ctx.Ctx, resp)
	if err != nil {
		return nil, err
	}

	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	respwrap := &FinalizeBlockWrap{
		Error: errmsg,
		Data:  respbz,
	}
	respwrapbz, err := json.Marshal(respwrap)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respwrapbz)
}

// Commit() (*abci.ResponseCommit, error)
func Commit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp, err := ctx.GetApplication().Commit()
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "Commit")
		return nil, err
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func RollbackToVersion(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	err := ctx.GetApplication().CommitMultiStore().RollbackToVersion(params[0].(int64))
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return rnh.AllocateWriteMem([]byte(errMsg))
}

func CheckTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req abci.RequestCheckTx
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	resp, err := ctx.GetApplication().CheckTx(&req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "CheckTx")
		return nil, err
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func wasmxHeaderHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req cmttypes.Header
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "HeaderHash")
		return nil, err
	}
	hash := req.Hash()
	return rnh.AllocateWriteMem(hash)
}

func validatorsToCmtValidators(interfaceRegistry cdctypes.InterfaceRegistry, vals []networktypes.TendermintValidator) ([]*cmttypes.Validator, error) {
	cmtvals := make([]*cmttypes.Validator, len(vals))
	for i, val := range vals {
		var pubkey cryptotypes.PubKey
		err := interfaceRegistry.UnpackAny(val.PubKey, &pubkey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert unpack cryptotypes.PubKey")
		}
		tmPk, err := cryptocodec.ToCmtProtoPublicKey(pubkey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert cryptotypes.PubKey to proto")
		}
		tmPk2, err := cometbftenc.PubKeyFromProto(tmPk)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert cryptotypes.PubKey from proto")
		}
		valaddr, err := hex.DecodeString(val.HexAddress)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to decode hex address")
		}
		v := &cmttypes.Validator{
			Address:          valaddr,
			PubKey:           tmPk2,
			VotingPower:      val.VotingPower,
			ProposerPriority: val.ProposerPriority,
		}
		cmtvals[i] = v
	}
	return cmtvals, nil
}

func wasmxValidatorsHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var vals networktypes.TendermintValidators
	err = ctx.CosmosHandler.Codec().UnmarshalJSON(reqbz, &vals)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ValidatorsHash")
		return nil, err
	}
	cmtvals, err := validatorsToCmtValidators(ctx.CosmosHandler.Codec().InterfaceRegistry(), vals.Validators)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ValidatorsHash")
		return nil, err
	}
	valSet, err := cmttypes.ValidatorSetFromExistingValidators(cmtvals)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ValidatorsHash")
		return nil, err
	}
	hash := valSet.Hash()
	return rnh.AllocateWriteMem(hash)
}

func wasmxConsensusParamsHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var cparams *cmttypes.ConsensusParams
	err = json.Unmarshal(reqbz, &cparams)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "ConsensusParamsHash")
		return nil, err
	}
	hash := cparams.Hash()
	return rnh.AllocateWriteMem(hash)
}

func wasmxBlockCommitVoteBytes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var vote cmtproto.Vote
	err = ctx.CosmosHandler.Codec().UnmarshalJSON(reqbz, &vote)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "BlockCommitVoteBytes", "reason", "unmarshal cmtproto.Vote")
		return nil, err
	}

	pb := cmttypes.CanonicalizeVote(ctx.Ctx.ChainID(), &vote)
	bz, err := protoio.MarshalDelimited(&pb)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error(), "consensus", "BlockCommitVoteBytes", "reason", "marshal cmtproto.CanonicalVote")
		return nil, err
	}
	return rnh.AllocateWriteMem(bz)
}
