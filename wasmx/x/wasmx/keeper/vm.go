package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"

	sdkerr "cosmossdk.io/errors"
	log "cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/sync/errgroup"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	utils "github.com/loredanacirstea/wasmx/x/wasmx/vm/utils"
)

type WasmxEngine struct {
	goRoutineGroup  *errgroup.Group
	goContextParent context.Context
	DataDir         string
	SourcesDir      string
	printDebug      bool
	app             types.Application
	Logger          func(ctx sdk.Context) log.Logger
	WasmRuntime     memc.IWasmVmMeta
}

func NewVM(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	dataDir string,
	sourcesDir string,
	printDebug bool,
	cacheSize uint32,
	app types.Application,
	logger func(ctx sdk.Context) log.Logger,
	wasmRuntime memc.IWasmVmMeta,
) (*WasmxEngine, error) {
	return &WasmxEngine{
		goRoutineGroup:  goRoutineGroup,
		goContextParent: goContextParent,
		DataDir:         dataDir,
		SourcesDir:      sourcesDir,
		printDebug:      printDebug,
		app:             app,
		Logger:          logger,
		WasmRuntime:     wasmRuntime,
	}, nil
}

func (k *WasmxEngine) Create(wasmBytecode types.WasmCode) (types.Checksum, error) {
	// get checksum and save wasm
	checksum := k.checksum(wasmBytecode)
	filepath := k.build_path(k.DataDir, checksum)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := utils.SafeWriteFile(filepath, wasmBytecode)
	if err != nil {
		return nil, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, nil
}

func (k *WasmxEngine) CreateUtf8(sourceCode []byte, extension string) (types.Checksum, error) {
	// get checksum and save source code
	checksum := k.checksum(sourceCode)
	filepath := k.build_path_utf8(k.SourcesDir, checksum, extension)

	// Read and write permissions for the owner and read-only permissions for everyone else
	err := utils.SafeWriteFile(filepath, sourceCode)
	if err != nil {
		return nil, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, nil
}

func (k *WasmxEngine) AnalyzeWasm(ctx sdk.Context, code types.WasmCode) (memc.AnalysisReport, error) {
	return vm.AnalyzeWasm(ctx, k.WasmRuntime, code)
}

func (k *WasmxEngine) Instantiate(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	initMsg []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
) (types.ContractResponse, uint64, error) {
	// TODO gas
	var data types.ContractResponse
	var err error

	if len(codeInfo.InterpretedBytecodeDeployment) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_INSTANTIATE, env, initMsg, store, cosmosHandler, gasMeter, contractInfo, nil, false, false, k.app, k.WasmRuntime.NewWasmVm)
	} else {
		// TODO gas
		data, err = vm.ExecuteWasm(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_INSTANTIATE, env, initMsg, store, cosmosHandler, gasMeter, contractInfo, nil, false, false, k.app, k.WasmRuntime.NewWasmVm)
	}
	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k *WasmxEngine) Execute(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
	dependencies []types.ContractDependency,
	inBackground bool,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error

	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_EXECUTE, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, inBackground, k.app, k.WasmRuntime.NewWasmVm)
	} else {
		data, err = vm.ExecuteWasm(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_EXECUTE, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, inBackground, k.app, k.WasmRuntime.NewWasmVm)
	}

	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k *WasmxEngine) ExecuteEntryPoint(
	ctx sdk.Context,
	contractEntryPoint string,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
	dependencies []types.ContractDependency,
	inBackground bool,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error

	// TODO if it has interpreter deps
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, contractEntryPoint, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, inBackground, k.app, k.WasmRuntime.NewWasmVm)
	} else {
		data, err = vm.ExecuteWasm(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, contractEntryPoint, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, inBackground, k.app, k.WasmRuntime.NewWasmVm)
	}

	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k *WasmxEngine) Reply(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
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
		data, err = vm.ExecuteWasmInterpreted(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_REPLY, env, wrappedMsgBz, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, false, k.app, k.WasmRuntime.NewWasmVm)
	} else {
		data, err = vm.ExecuteWasm(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_REPLY, env, wrappedMsgBz, store, cosmosHandler, gasMeter, contractInfo, dependencies, false, false, k.app, k.WasmRuntime.NewWasmVm)
	}

	if err != nil {
		return types.ContractResponse{}, 0, err
	}
	return data, 0, nil
}

func (k *WasmxEngine) QueryExecute(
	ctx sdk.Context,
	codeInfo *types.CodeInfo,
	env types.Env,
	executeMsg []byte,
	store prefix.Store,
	cosmosHandler types.WasmxCosmosHandler,
	gasMeter types.GasMeter,
	contractInfo types.ContractDependency,
	dependencies []types.ContractDependency,
	isdebug bool,
) (types.ContractResponse, uint64, error) {
	var data types.ContractResponse
	var err error
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 || types.HasUtf8Dep(codeInfo.Deps) {
		data, err = vm.ExecuteWasmInterpreted(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_QUERY, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, isdebug, false, k.app, k.WasmRuntime.NewWasmVm)
	} else {
		data, err = vm.ExecuteWasm(k.goRoutineGroup, k.goContextParent, ctx, k.Logger, types.ENTRY_POINT_QUERY, env, executeMsg, store, cosmosHandler, gasMeter, contractInfo, dependencies, isdebug, false, k.app, k.WasmRuntime.NewWasmVm)
	}
	if err != nil {
		return data, 0, err
	}
	return data, 0, nil
}

func (k *WasmxEngine) GetCode(checksum types.Checksum, deps []string) (types.WasmCode, error) {
	if types.HasUtf8Dep(deps) {
		return k.GetCodeUtf8(checksum, deps)
	}
	return k.load_wasm(checksum)
}

func (k *WasmxEngine) GetCodeUtf8(checksum types.Checksum, deps []string) (types.WasmCode, error) {
	extension := GetExtensionFromDeps(deps)
	return k.load_utf8(extension, checksum)
}

func (k *WasmxEngine) Cleanup() {

}

func (k *WasmxEngine) Pin(ctx sdk.Context, checksum types.Checksum, compiledFolderPath string, meteringOff bool) error {
	pinnedPath := k.build_path_pinned(k.DataDir, checksum)
	// if we do not find a precompiled contract in our cache, we compile it here
	if compiledFolderPath != "" {
		compiledPath := k.build_path(compiledFolderPath, checksum)
		err := copyFile(compiledPath, pinnedPath)
		if err == nil {
			return nil
		}
	}
	err := k.pin_code(ctx, k.build_path(k.DataDir, checksum), pinnedPath, meteringOff)
	if err != nil {
		return err
	}
	// we cache this contract if cache exists
	if compiledFolderPath != "" {
		compiledPath := k.build_path(compiledFolderPath, checksum)
		copyFile(pinnedPath, compiledPath)
	}
	return nil
}

func (k *WasmxEngine) Unpin(checksum types.Checksum) error {
	// TODO
	// remove pinned compiled code
	return nil
}

func (k *WasmxEngine) pin_code(ctx sdk.Context, inPath string, outPath string, meteringOff bool) error {
	return k.WasmRuntime.AotCompile(ctx, inPath, outPath, meteringOff)
}

func (k *WasmxEngine) checksum(wasmBytecode types.WasmCode) types.Checksum {
	h := sha256.New()
	h.Write(wasmBytecode)
	return h.Sum(nil)
}

func (k *WasmxEngine) load_utf8(extension string, checksum types.Checksum) ([]byte, error) {
	filepath := k.build_path_utf8(k.SourcesDir, checksum, extension)
	return os.ReadFile(filepath)
}

func (k *WasmxEngine) load_wasm(checksum types.Checksum) (types.WasmCode, error) {
	filepath := k.build_path(k.DataDir, checksum)
	return os.ReadFile(filepath)
}

func (k *WasmxEngine) build_path(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, hex.EncodeToString(checksum))
}

func (k *WasmxEngine) build_path_pinned(dataDir string, checksum types.Checksum) string {
	return path.Join(dataDir, types.PINNED_FOLDER, hex.EncodeToString(checksum))
}

func (k *WasmxEngine) build_path_utf8(dataDir string, checksum types.Checksum, extension string) string {
	filename := fmt.Sprintf("%s_%s.%s", extension, hex.EncodeToString(checksum), extension)
	return path.Join(dataDir, extension, filename)
}

func (k *WasmxEngine) BuildPathPinned(codeHash []byte) string {
	return k.build_path_pinned(k.DataDir, codeHash)
}

func (k *WasmxEngine) BuildPath(codeHash []byte) string {
	return k.build_path(k.DataDir, codeHash)
}

func (k *WasmxEngine) BuildPathUtf8(codeHash []byte, extension string) string {
	return k.build_path_utf8(k.SourcesDir, codeHash, extension)
}

func (k *WasmxEngine) GetPathUtf8(codeInfo types.CodeInfo) string {
	return k.BuildPathUtf8(codeInfo.CodeHash, GetExtensionFromDeps(codeInfo.Deps))
}

func (k *WasmxEngine) GetCodeFilePath(codeInfo types.CodeInfo) string {
	filepath := ""
	if types.HasUtf8Dep(codeInfo.Deps) {
		filepath = k.GetPathUtf8(codeInfo)
	} else {
		if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
			filepath = ""
		} else {
			filepath = k.BuildPath(codeInfo.CodeHash)
		}
	}
	return filepath
}

func (k *WasmxEngine) GetFilePath(codeInfo types.CodeInfo) string {
	filepath := ""
	if codeInfo.Pinned {
		filepath = k.BuildPathPinned(codeInfo.CodeHash)
	} else {
		if types.HasUtf8Dep(codeInfo.Deps) {
			filepath = k.GetPathUtf8(codeInfo)
		} else {
			if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
				filepath = ""
			} else {
				filepath = k.BuildPath(codeInfo.CodeHash)
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
	err = utils.SafeWriteFile(destinationFile, input)
	if err != nil {
		return err
	}
	return nil
}
