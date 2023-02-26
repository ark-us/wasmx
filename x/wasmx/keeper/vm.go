package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"path"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/keeper/ewasm"
	"wasmx/x/wasmx/types"
)

type WasmxEngine struct {
	DataDir    string
	Cache      map[string]*wasmedge.VM
	printDebug bool
}

func NewVM(dataDir string, memoryLimit uint32, printDebug bool, cacheSize uint32) (*WasmxEngine, error) {
	// cache, err := api.InitCache(dataDir, supportedFeatures, cacheSize, memoryLimit)
	// if err != nil {
	// 	return nil, err
	// }
	return &WasmxEngine{
		DataDir:    dataDir,
		Cache:      make(map[string]*wasmedge.VM, 0),
		printDebug: printDebug,
	}, nil
}

func (k WasmxEngine) Create(code types.WasmCode) (types.Checksum, error) {
	// TODO analyze code

	// get checksum and save wasm
	return k.save_wasm(k.DataDir, code)
}

func (k WasmxEngine) AnalyzeCode(checksum types.Checksum) (types.AnalysisReport, error) {
	return types.AnalysisReport{}, nil
}

func (k WasmxEngine) Instantiate(
	checksum types.Checksum,
	env types.Env,
	info types.MessageInfo,
	initMsg []byte,
	store types.KVStore,
	// querier types.Querier,
	// gasMeter types.GasMeter,
	// gasLimit uint64,
	// deserCost types.UFraction,
) (types.ContractResponse, uint64, error) {
	// load wasm
	// execute instantiate export

	// TODO gas
	filepath := k.build_path(k.DataDir, checksum)
	data, err := ewasm.ExecuteWasm(filepath, "instantiate", initMsg)
	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return types.ContractResponse{Data: data}, 0, nil
}

func (k WasmxEngine) Execute(
	checksum types.Checksum,
	env types.Env,
	info types.MessageInfo,
	executeMsg []byte,
	store types.KVStore,
	// querier types.Querier,
	// gasMeter types.GasMeter,
	gasLimit uint64,
	// deserCost types.UFraction,
) (types.ContractResponse, uint64, error) {
	// load wasm
	// execute instantiate export
	filepath := k.build_path(k.DataDir, checksum)
	data, err := ewasm.ExecuteWasm(filepath, "main", executeMsg)
	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return types.ContractResponse{Data: data}, 0, nil
}

func (k WasmxEngine) QueryExecute(
	code types.Checksum,
	env types.Env,
	info types.MessageInfo,
	executeMsg []byte,
	store types.KVStore,
	querier types.Querier,
	gasMeter types.GasMeter,
	gasLimit uint64,
	// deserCost types.UFraction,
) (types.ContractResponse, uint64, error) {
	return types.ContractResponse{}, 0, nil
}

func (k WasmxEngine) GetCode(checksum types.Checksum) (types.WasmCode, error) {
	return k.load_wasm(k.DataDir, checksum)
}

func (k WasmxEngine) Cleanup() {

}

func (k WasmxEngine) Pin(checksum types.Checksum) error {
	return nil
}

func (k WasmxEngine) Unpin(checksum types.Checksum) error {
	return nil
}

func (k WasmxEngine) save_wasm(dataDir string, wasmBytecode types.WasmCode) (types.Checksum, error) {
	// TODO analyze code

	h := sha256.New()
	h.Write(wasmBytecode)
	checksum := h.Sum(nil)
	filepath := k.build_path(k.DataDir, checksum)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := ioutil.WriteFile(filepath, wasmBytecode, 0644)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	return checksum, nil
}

func (k WasmxEngine) load_wasm(dataDir string, checksum types.Checksum) (types.WasmCode, error) {
	filepath := k.build_path(k.DataDir, checksum)
	return ioutil.ReadFile(filepath)
}

func (k WasmxEngine) build_path(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, hex.EncodeToString(checksum))
}
