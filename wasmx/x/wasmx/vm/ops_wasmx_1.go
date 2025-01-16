package vm

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var (
	SSTORE_GAS_WASMX = 20_000
)

// cosmos-sdk key-value max sizes
// MaxKeyLength: 128K - 1
// MaxValueLength: 2G - 1

type WasmxJsonLog struct {
	Type   string
	Data   []byte
	Topics [][32]byte
}

// getCallData(): ArrayBuffer
func getCallData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ptr, err := rnh.AllocateWriteMem(ctx.Env.CurrentCall.CallData)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

func wasmxGetCaller(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr := types.PaddLeftTo32(ctx.Env.CurrentCall.Sender.Bytes())
	ptr, err := rnh.AllocateWriteMem(addr)
	if err != nil {
		return nil, err

	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

func wasmxGetAddress(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr := types.PaddLeftTo32(ctx.Env.Contract.Address.Bytes())
	ptr, err := rnh.AllocateWriteMem(addr)
	if err != nil {
		return nil, err

	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

// storageStore(key: ArrayBuffer, value: ArrayBuffer)
func wasmxStorageStore(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	key, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	data, err := rnh.ReadMemFromPtr(params[1])
	if err != nil {
		return nil, err
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_WASMX), "wasmxStorageStore")
	ctx.ContractStore.Set(key, data)
	returns := make([]interface{}, 0)
	return returns, nil
}

// storageLoad(key: ArrayBuffer): ArrayBuffer
func wasmxStorageLoad(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keybz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	data := ctx.ContractStore.Get(keybz)
	// if len(data) == 0 {
	// 	data = make([]byte, 32)
	// }
	newptr, err := rnh.AllocateWriteMem(data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, nil
}

// wasmxStorageDelete(key: ArrayBuffer)
func wasmxStorageDelete(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	key, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	// refund some gas?
	ctx.ContractStore.Delete(key)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxStorageDeleteRange(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StorageDeleteRange
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	startKey := req.StartKey
	endKey := req.EndKey
	if len(startKey) == 0 {
		startKey = nil
	}
	if len(endKey) == 0 {
		endKey = nil
	}

	iter := ctx.ContractStore.Iterator(startKey, endKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		ctx.ContractStore.Delete(iter.Key())
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxStorageLoadRange(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StorageRange
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	values := make([][]byte, 0)
	startKey := req.StartKey
	endKey := req.EndKey
	if len(startKey) == 0 {
		startKey = nil
	}
	if len(endKey) == 0 {
		endKey = nil
	}

	var iter types.Iterator
	if req.Reverse {
		iter = ctx.ContractStore.ReverseIterator(startKey, endKey)
	} else {
		iter = ctx.ContractStore.Iterator(startKey, endKey)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		values = append(values, iter.Value())
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	newptr, err := rnh.AllocateWriteMem(data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, nil
}

func wasmxStorageLoadRangePairs(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StorageRange
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	pairs := make([]StoragePair, 0)
	startKey := req.StartKey
	endKey := req.EndKey
	if len(startKey) == 0 {
		startKey = nil
	}
	if len(endKey) == 0 {
		endKey = nil
	}

	var iter types.Iterator
	if req.Reverse {
		iter = ctx.ContractStore.ReverseIterator(startKey, endKey)
	} else {
		iter = ctx.ContractStore.Iterator(startKey, endKey)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		pair := StoragePair{Key: iter.Key(), Value: iter.Value()}
		pairs = append(pairs, pair)
	}
	response := StoragePairs{Values: pairs}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	newptr, err := rnh.AllocateWriteMem(data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, nil
}

func wasmxLog(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var wlog WasmxJsonLog
	err = json.Unmarshal(data, &wlog)
	if err != nil {
		return nil, err
	}
	logtype := wlog.Type
	// cosmos log attributes cannot be empty
	if logtype == "" {
		logtype = LOG_TYPE_WASMX
	}
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}
	log := WasmxLog{
		Type:             logtype,
		ContractAddress:  ctx.Env.Contract.Address,
		SystemDependency: dependency,
		Data:             wlog.Data,
		Topics:           wlog.Topics,
	}
	ctx.Logs = append(ctx.Logs, log)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxGetReturnData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	newptr, err := rnh.AllocateWriteMem(ctx.ReturnData)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, nil
}

func wasmxGetFinishData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	newptr, err := rnh.AllocateWriteMem(ctx.FinishData)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, nil
}

func wasmxSetFinishData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	ctx.FinishData = data
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxFinish(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	return returns, nil
}

func wasmxRevert(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	return returns, fmt.Errorf("revert")
}

// message: usize, fileName: usize, line: u32, column: u32
func asAbort(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	message, _ := rnh.ReadMemFromPtr(params[0])
	fileName, _ := rnh.ReadMemFromPtr(params[1])
	ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx_env_1: ABORT: %s, %s. line: %d, column: %d", ctx.RuntimeHandler.ReadJsString(message), ctx.RuntimeHandler.ReadJsString(fileName), params[2], params[3]))
	return wasmxRevert(_context, rnh, params)
}

func asConsoleLog(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	message, err := rnh.ReadStringFromPtr(params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.log: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.log error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

func asConsoleInfo(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	message, err := rnh.ReadStringFromPtr(params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.info: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.info error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

func asConsoleError(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	message, err := rnh.ReadStringFromPtr(params[0])
	if err == nil {
		ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("wasmx: console.error: %s", message))
	} else {
		ctx.Logger(ctx.Ctx).Info(fmt.Sprintf("wasmx: console.error error: %s", err.Error()))
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

func asConsoleDebug(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	message, _ := rnh.ReadStringFromPtr(params[0])
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("wasmx: console.debug: %s", message))
	returns := make([]interface{}, 0)
	return returns, nil
}

func asDateNow(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	returns[0] = float64(time.Now().UTC().UnixMilli())
	return returns, nil
}

// TODO - move this only for non-deterministic contracts
func asSeed(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return nil, err
	}
	returns[0] = float64(binary.LittleEndian.Uint64(b[:]))
	return returns, nil
}

// function env.trace?(message: usize, n: i32, a0..4?: f64): void
// function env.seed?(): f64

func BuildWasmxEnv1(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("getCallData", getCallData, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("getCaller", wasmxGetCaller, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("getAddress", wasmxGetAddress, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("storageLoad", wasmxStorageLoad, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("storageStore", wasmxStorageStore, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("storageLoadRange", wasmxStorageLoadRange, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("storageLoadRangePairs", wasmxStorageLoadRangePairs, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("log", wasmxLog, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("getReturnData", wasmxGetReturnData, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("getFinishData", wasmxGetFinishData, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("setFinishData", wasmxSetFinishData, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("setReturnData", wasmxSetFinishData, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("finish", wasmxFinish, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("revert", wasmxRevert, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
	}

	return vm.BuildModule(rnh, "wasmx", context, fndefs)
}

func BuildAssemblyScriptEnv(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("abort", asAbort, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("console.log", asConsoleLog, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("console.info", asConsoleInfo, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("console.error", asConsoleError, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("console.debug", asConsoleDebug, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("Date.now", asDateNow, []interface{}{}, []interface{}{vm.ValType_F64()}, 0),
		vm.BuildFn("seed", asSeed, []interface{}{}, []interface{}{vm.ValType_F64()}, 0),
	}

	return vm.BuildModule(rnh, "env", context, fndefs)
}
