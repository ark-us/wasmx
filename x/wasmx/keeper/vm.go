package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm"
)

type WasmxEngine struct {
	DataDir    string
	printDebug bool
}

func NewVM(dataDir string, memoryLimit uint32, printDebug bool, cacheSize uint32) (*WasmxEngine, error) {
	return &WasmxEngine{
		DataDir:    dataDir,
		printDebug: printDebug,
	}, nil
}

func (k WasmxEngine) Create(code types.WasmCode) (types.Checksum, error) {
	// TODO analyze code

	// get checksum and save wasm
	return k.save_wasm(k.DataDir, code)
}

func (k WasmxEngine) AnalyzeWasm(code types.WasmCode) (types.AnalysisReport, error) {
	return vm.AnalyzeWasm(code)
}

func (k WasmxEngine) Instantiate(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	initMsg []byte,
	prefixStoreKey []byte,
	store types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
) (types.ContractResponse, uint64, error) {
	// TODO gas
	checksum := codeInfo.CodeHash
	pinned := codeInfo.Pinned
	var data types.ContractResponse
	var err error

	if len(codeInfo.InterpretedBytecodeDeployment) > 0 {
		data, err = vm.ExecuteWasmInterpreted(ctx, "instantiate", env, initMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, nil)
	} else {
		// TODO gas
		var filepath string
		if pinned {
			filepath = k.build_path_pinned(k.DataDir, checksum)
		} else {
			filepath = k.build_path(k.DataDir, checksum)
		}
		data, err = vm.ExecuteWasm(ctx, filepath, "instantiate", env, initMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, nil)
	}
	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k WasmxEngine) Execute(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	prefixStoreKey []byte,
	store types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error
	checksum := codeInfo.CodeHash
	pinned := codeInfo.Pinned

	if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
		data, err = vm.ExecuteWasmInterpreted(ctx, "main", env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies)
	} else {

		var filepath string
		if pinned {
			filepath = k.build_path_pinned(k.DataDir, checksum)
		} else {
			filepath = k.build_path(k.DataDir, checksum)
		}
		data, err = vm.ExecuteWasm(ctx, filepath, "main", env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies)
	}

	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k WasmxEngine) QueryExecute(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	prefixStoreKey []byte,
	store types.KVStore,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
) (types.WasmxQueryResponse, uint64, error) {
	var data types.ContractResponse
	var err error
	checksum := codeInfo.CodeHash
	pinned := codeInfo.Pinned

	if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
		data, err = vm.ExecuteWasmInterpreted(ctx, "main", env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies)
	} else {

		var filepath string
		if pinned {
			filepath = k.build_path_pinned(k.DataDir, checksum)
		} else {
			filepath = k.build_path(k.DataDir, checksum)
		}
		data, err = vm.ExecuteWasm(ctx, filepath, "main", env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies)
	}

	if err != nil {
		return types.WasmxQueryResponse{}, 0, err
	}
	return types.WasmxQueryResponse{Data: data.Data}, 0, nil
}

func (k WasmxEngine) GetCode(checksum types.Checksum) (types.WasmCode, error) {
	return k.load_wasm(k.DataDir, checksum)
}

func (k WasmxEngine) Cleanup() {

}

func (k WasmxEngine) Pin(checksum types.Checksum, compiledFolderPath string) error {
	pinnedPath := k.build_path_pinned(k.DataDir, checksum)
	fmt.Println("--WasmxEngine-Pin", compiledFolderPath)
	if compiledFolderPath != "" {
		compiledPath := k.build_path(compiledFolderPath, checksum) + ".so"
		err := copyFile(compiledPath, pinnedPath)
		if err != nil {
			return nil
		}
		return nil
	}
	return k.pin_code(k.build_path(k.DataDir, checksum), pinnedPath)
}

func (k WasmxEngine) Unpin(checksum types.Checksum) error {
	// TODO
	// remove pinned compiled code
	return nil
}

func (k WasmxEngine) pin_code(inPath string, outPath string) error {
	fmt.Println("--pin_code", inPath, outPath)
	return vm.AotCompile(inPath, outPath)
}

func (k WasmxEngine) checksum(wasmBytecode types.WasmCode) types.Checksum {
	h := sha256.New()
	h.Write(wasmBytecode)
	return h.Sum(nil)
}

func (k WasmxEngine) save_wasm(dataDir string, wasmBytecode types.WasmCode) (types.Checksum, error) {
	checksum := k.checksum(wasmBytecode)
	filepath := k.build_path(k.DataDir, checksum)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := os.WriteFile(filepath, wasmBytecode, 0644)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, nil
}

func (k WasmxEngine) load_wasm(dataDir string, checksum types.Checksum) (types.WasmCode, error) {
	filepath := k.build_path(k.DataDir, checksum)
	return os.ReadFile(filepath)
}

func (k WasmxEngine) build_path(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, hex.EncodeToString(checksum))
}

func (k WasmxEngine) build_path_pinned(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, types.PINNED_FOLDER, hex.EncodeToString(checksum)+".so")
}

func copyFile(sourceFile string, destinationFile string) error {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = os.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}
	return nil
}
