package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm"
)

type WasmxEngine struct {
	DataDir    string
	SourcesDir string
	printDebug bool
}

func NewVM(dataDir string, sourcesDir string, memoryLimit uint32, printDebug bool, cacheSize uint32) (*WasmxEngine, error) {
	return &WasmxEngine{
		DataDir:    dataDir,
		SourcesDir: sourcesDir,
		printDebug: printDebug,
	}, nil
}

func (k WasmxEngine) Create(wasmBytecode types.WasmCode) (types.Checksum, error) {
	// get checksum and save wasm
	checksum := k.checksum(wasmBytecode)
	filepath := k.build_path(k.DataDir, checksum)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := os.WriteFile(filepath, wasmBytecode, 0644)
	if err != nil {
		return nil, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, nil
}

func (k WasmxEngine) CreateUtf8(sourceCode []byte, extension string) (types.Checksum, error) {
	// get checksum and save source code
	checksum := k.checksum(sourceCode)
	filepath := k.build_path_utf8(k.SourcesDir, checksum, extension)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := os.WriteFile(filepath, sourceCode, 0644)
	if err != nil {
		return nil, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, nil
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
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
) (types.ContractResponse, uint64, error) {
	// TODO gas
	var data types.ContractResponse
	var err error

	if len(codeInfo.InterpretedBytecodeDeployment) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(ctx, types.ENTRY_POINT_INSTANTIATE, env, initMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, nil, false)
	} else {
		// TODO gas
		data, err = vm.ExecuteWasm(ctx, types.ENTRY_POINT_INSTANTIATE, env, initMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, nil, false)
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
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error

	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(ctx, types.ENTRY_POINT_EXECUTE, env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, false)
	} else {
		data, err = vm.ExecuteWasm(ctx, types.ENTRY_POINT_EXECUTE, env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, false)
	}

	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k WasmxEngine) Reply(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	prefixStoreKey []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency, // TODO remove
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error
	wrappedMsg := types.WasmxExecutionMessage{Data: executeMsg}
	wrappedMsgBz, err := json.Marshal(wrappedMsg)
	if err != nil {
		return types.ContractResponse{}, 0, err
	}

	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(ctx, types.ENTRY_POINT_REPLY, env, wrappedMsgBz, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, false)
	} else {
		data, err = vm.ExecuteWasm(ctx, types.ENTRY_POINT_REPLY, env, wrappedMsgBz, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, false)
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
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	systemDeps []types.SystemDep,
	dependencies []types.ContractDependency,
	isdebug bool,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(ctx, types.ENTRY_POINT_QUERY, env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, isdebug)
	} else {
		data, err = vm.ExecuteWasm(ctx, types.ENTRY_POINT_QUERY, env, executeMsg, prefixStoreKey, store, cosmosHandler, gasMeter, systemDeps, dependencies, isdebug)
	}

	if err != nil {
		return data, 0, err
	}
	return data, 0, nil
}

func (k WasmxEngine) GetCode(checksum types.Checksum, deps []string) (types.WasmCode, error) {
	if types.HasUtf8Dep(deps) {
		extension := GetExtensionFromDeps(deps)
		return k.load_utf8(extension, checksum)
	}
	return k.load_wasm(checksum)
}

func (k WasmxEngine) Cleanup() {

}

func (k WasmxEngine) Pin(checksum types.Checksum, compiledFolderPath string) error {
	pinnedPath := k.build_path_pinned(k.DataDir, checksum)
	if compiledFolderPath != "" {
		compiledPath := k.build_path(compiledFolderPath, checksum) + ".so"
		err := copyFile(compiledPath, pinnedPath)
		if err == nil {
			return nil
		}
	}
	return k.pin_code(k.build_path(k.DataDir, checksum), pinnedPath)
}

func (k WasmxEngine) Unpin(checksum types.Checksum) error {
	// TODO
	// remove pinned compiled code
	return nil
}

func (k WasmxEngine) pin_code(inPath string, outPath string) error {
	return vm.AotCompile(inPath, outPath)
}

func (k WasmxEngine) checksum(wasmBytecode types.WasmCode) types.Checksum {
	h := sha256.New()
	h.Write(wasmBytecode)
	return h.Sum(nil)
}

func (k WasmxEngine) load_utf8(extension string, checksum types.Checksum) ([]byte, error) {
	filepath := k.build_path_utf8(k.SourcesDir, checksum, extension)
	return os.ReadFile(filepath)
}

func (k WasmxEngine) load_wasm(checksum types.Checksum) (types.WasmCode, error) {
	filepath := k.build_path(k.DataDir, checksum)
	return os.ReadFile(filepath)
}

func (k WasmxEngine) build_path(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, hex.EncodeToString(checksum))
}

func (k WasmxEngine) build_path_pinned(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, types.PINNED_FOLDER, hex.EncodeToString(checksum)+".so")
}

func (k WasmxEngine) build_path_utf8(dataDir string, checksum types.Checksum, extension string) string {
	filename := fmt.Sprintf("%s_%s.%s", extension, hex.EncodeToString(checksum), extension)
	return path.Join(dataDir, extension, filename)
}

func (k WasmxEngine) GetFilePath(codeInfo types.CodeInfo) string {
	filepath := ""
	if codeInfo.Pinned {
		filepath = k.build_path_pinned(k.DataDir, codeInfo.CodeHash)
	} else {
		if types.HasUtf8Dep(codeInfo.Deps) {
			extension := GetExtensionFromDeps(codeInfo.Deps)
			filepath = k.build_path_utf8(k.SourcesDir, codeInfo.CodeHash, extension)
		} else {
			if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
				filepath = ""
			} else {
				filepath = k.build_path(k.DataDir, codeInfo.CodeHash)
			}
		}
	}
	return filepath
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
