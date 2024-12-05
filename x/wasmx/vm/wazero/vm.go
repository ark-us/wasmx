package wazero

// import (
//     "context"
//     "encoding/hex"
//     "encoding/json"
//     "fmt"
//     "strings"

//     "golang.org/x/exp/slices"
//     "golang.org/x/sync/errgroup"

//     sdkerr "cosmossdk.io/errors"
//     log "cosmossdk.io/log"
//     "cosmossdk.io/store/prefix"
//     sdk "github.com/cosmos/cosmos-sdk/types"

//     "github.com/tetratelabs/wazero"
//     "github.com/tetratelabs/wazero/api"

//     "mythos/v1/x/wasmx/types"
//     vmi "mythos/v1/x/wasmx/vm/interfaces"
//     "mythos/v1/x/wasmx/vm/wasmutils"
// )

// // Runtime is a global wazero runtime instance
// var Runtime wazero.Runtime

// func init() {
//     Runtime = wazero.NewRuntime(context.Background())
// }

// // EwasmWrapper wraps wasm function calls
// func EwasmWrapper(ctx context.Context, module api.Module, name string, params ...uint64) ([]uint64, error) {
//     fn := module.ExportedFunction(name)
//     if fn == nil {
//         return nil, fmt.Errorf("function %s not found", name)
//     }

//     return fn.Call(ctx, params...)
// }

// func InitiateWasm(context *Context, filePath string, wasmbuffer []byte, systemDeps []types.SystemDep) (api.Module, []func(), error) {
//     var cleanups []func()
//     var err error

//     // Create new module
//     builder := Runtime.NewHostModuleBuilder("env")

//     // Set default system dependency if none provided
//     if len(systemDeps) == 0 {
//         label := types.DEFAULT_SYS_DEP
//         systemDeps = append(systemDeps, types.SystemDep{Role: label, Label: label})
//     }

//     // Initialize dependencies
//     _cleanups, err := initiateWasmDeps(context, builder, systemDeps)
//     cleanups = append(cleanups, _cleanups...)
//     if err != nil {
//         return nil, cleanups, err
//     }

//     // Instantiate module
//     var module api.Module
//     if filePath != "" || len(wasmbuffer) > 0 {
//         module, err = wasmutils.InstantiateWasm(context.GoContextParent, Runtime, builder, filePath, wasmbuffer)
//         if err != nil {
//             return nil, cleanups, err
//         }
//         cleanups = append(cleanups, func() { module.Close(context.GoContextParent) })
//     }

//     return module, cleanups, nil
// }

// func initiateWasmDeps(context *Context, builder wazero.HostModuleBuilder, systemDeps []types.SystemDep) ([]func(), error) {
//     cleanups := make([]func(), 0)

//     for _, systemDep := range systemDeps {
//         // Handle nested dependencies
//         _cleanups, err := initiateWasmDeps(context, builder, systemDep.Deps)
//         cleanups = append(cleanups, _cleanups...)
//         if err != nil {
//             return cleanups, err
//         }

//         // Find and apply system dependency handler
//         handler, found := SystemDepHandler[systemDep.Role]
//         if !found {
//             handler, found = SystemDepHandler[systemDep.Label]
//         }
//         if found {
//             releaseFn, err := handler(context, builder, &systemDep)
//             cleanups = append(cleanups, releaseFn...)
//             if err != nil {
//                 return cleanups, err
//             }
//         }
//     }

//     return cleanups, nil
// }

// func getMemoryHandler(systemDeps []types.SystemDep) vmi.MemoryHandler {
//     handler := getMemoryHandlerFromDeps(systemDeps)
//     if handler != nil {
//         return handler
//     }
//     return MemoryDepHandler[types.WASMX_MEMORY_ASSEMBLYSCRIPT]
// }

// func getMemoryHandlerFromDeps(systemDeps []types.SystemDep) vmi.MemoryHandler {
//     for _, systemDep := range systemDeps {
//         handler, found := MemoryDepHandler[systemDep.Role]
//         if !found {
//             handler, found = MemoryDepHandler[systemDep.Label]
//         }
//         if found {
//             return handler
//         }
//     }

//     for _, systemDep := range systemDeps {
//         handler := getMemoryHandlerFromDeps(systemDep.Deps)
//         if handler != nil {
//             return handler
//         }
//     }
//     return nil
// }

// func runCleanups(cleanups []func()) {
//     for i := len(cleanups) - 1; i >= 0; i-- {
//         cleanups[i]()
//     }
// }

// func ExecuteWasmInterpreted(
//     goRoutineGroup *errgroup.Group,
//     goContextParent context.Context,
//     ctx sdk.Context,
//     logger func(ctx sdk.Context) log.Logger,
//     funcName string,
//     env types.Env,
//     msg []byte,
//     storeKey []byte,
//     kvstore prefix.Store,
//     storageType types.ContractStorageType,
//     cosmosHandler types.WasmxCosmosHandler,
//     gasMeter types.GasMeter,
//     systemDeps []types.SystemDep,
//     dependencies []types.ContractDependency,
//     isdebug bool,
//     inBackground bool,
//     app types.Application,
// ) (types.ContractResponse, error) {
//     var ethMsg types.WasmxExecutionMessage
//     if err := json.Unmarshal(msg, &ethMsg); err != nil {
//         return types.ContractResponse{}, sdkerr.Wrapf(err, "could not decode wasm execution message")
//     }

//     var cleanups []func()
//     defer func() { runCleanups(cleanups) }()

//     var contractRouter ContractRouter = make(map[string]*ContractContext)
//     context := &Context{
//         GoRoutineGroup:  goRoutineGroup,
//         GoContextParent: goContextParent,
//         Ctx:             ctx,
//         Logger:          GetVmLogger(logger, env.Chain.ChainIdFull, env.Contract.Address.String()),
//         GasMeter:        gasMeter,
//         Env:             &env,
//         ContractStore:   kvstore,
//         CosmosHandler:   cosmosHandler,
//         App:             app,
//         ContractRouter:  contractRouter,
//         NativeHandler:   NativeMap,
//         dbIterators:     map[int32]types.Iterator{},
//         MemoryHandler:   getMemoryHandler(systemDeps),
//     }
//     context.Env.CurrentCall.CallData = ethMsg.Data

//     // Setup contract contexts
//     for _, dep := range dependencies {
//         contractContext := buildExecutionContextClassic(dep)
//         if contractContext == nil {
//             return types.ContractResponse{}, fmt.Errorf("could not build dependency execution context for %s", dep.Address)
//         }
//         context.ContractRouter[dep.Address.String()] = contractContext
//     }

//     // Setup self context
//     selfContext := buildExecutionContextClassic(types.ContractDependency{
//         FilePath:    "",
//         Bytecode:    []byte{},
//         CodeHash:    []byte{},
//         StoreKey:    storeKey,
//         StorageType: storageType,
//         SystemDeps:  systemDeps,
//     })
//     if selfContext == nil {
//         return types.ContractResponse{}, fmt.Errorf("could not build execution context for self")
//     }
//     contractstr := env.Contract.Address.String()
//     context.ContractRouter[contractstr] = selfContext

//     // Initialize Wasm module
//     module, _cleanups, err := InitiateWasm(context, "", nil, systemDeps)
//     cleanups = append(cleanups, _cleanups...)
//     if err != nil {
//         return types.ContractResponse{}, err
//     }

//     selfContext.Module = module
//     setExecutionBytecode(context, module, funcName)
//     selfContext.ContractInfo.Bytecode = context.Env.Contract.Bytecode
//     selfContext.ContractInfo.CodeHash = context.Env.Contract.CodeHash

//     executeHandler := GetExecuteFunctionHandler(systemDeps)

//     if inBackground {
//         err = types.AddBackgroundProcesses(goContextParent, &types.BackgroundProcess{
//             Label:   contractstr,
//             Module:  module,
//             ExecuteHandler: func(funcName_ string) ([]byte, error) {
//                 _, err := executeHandler(context, module, funcName_, []uint64{})
//                 if err != nil {
//                     return nil, err
//                 }
//                 return context.FinishData, nil
//             },
//         })
//         if err != nil {
//             return types.ContractResponse{}, err
//         }
//     }

//     _, err = executeHandler(context, module, funcName, []uint64{})
//     if err != nil {
//         wrapErr := sdkerr.Wrapf(
//             err,
//             "chain_id: %s; contract: %s; entry point: %s; revert: %s",
//             context.Ctx.ChainID(),
//             context.Env.Contract.Address.String(),
//             funcName,
//             string(context.FinishData),
//         )
//         resp := handleContractErrorResponse(module, context.FinishData, isdebug, wrapErr)
//         if isdebug {
//             return resp, nil
//         }
//         return resp, wrapErr
//     }

//     return handleContractResponse(context, module, isdebug), nil
// }

// func ExecuteWasm(
//     goRoutineGroup *errgroup.Group,
//     goContextParent context.Context,
//     ctx sdk.Context,
//     logger func(ctx sdk.Context) log.Logger,
//     funcName string,
//     env types.Env,
//     msg []byte,
//     storeKey []byte,
//     kvstore prefix.Store,
//     storageType types.ContractStorageType,
//     cosmosHandler types.WasmxCosmosHandler,
//     gasMeter types.GasMeter,
//     systemDeps []types.SystemDep,
//     dependencies []types.ContractDependency,
//     isdebug bool,
//     inBackground bool,
//     app types.Application,
// ) (types.ContractResponse, error) {
//     var ethMsg types.WasmxExecutionMessage
//     if err := json.Unmarshal(msg, &ethMsg); err != nil {
//         return types.ContractResponse{}, sdkerr.Wrapf(err, "could not decode wasm execution message")
//     }

//     var cleanups []func()
//     defer func() { runCleanups(cleanups) }()

//     var contractRouter ContractRouter = make(map[string]*ContractContext)
//     context := &Context{
//         GoRoutineGroup:  goRoutineGroup,
//         GoContextParent: goContextParent,
//         Ctx:             ctx,
//         Logger:          GetVmLogger(logger, env.Chain.ChainIdFull, env.Contract.Address.String()),
//         GasMeter:        gasMeter,
//         Env:             &env,
//         ContractStore:   kvstore,
//         CosmosHandler:   cosmosHandler,
//         ContractRouter:  contractRouter,
//         App:             app,
//         NativeHandler:   NativeMap,
//         dbIterators:     map[int32]types.Iterator{},
//         MemoryHandler:   getMemoryHandler(systemDeps),
//     }
//     context.Env.CurrentCall.CallData = ethMsg.Data

//     // Check for native implementations
//     if context.NativeHandler.IsPrecompile(env.Contract.Address.Bytes()) {
//         data, err := context.NativeHandler.Execute(context, env.Contract.Address, ethMsg.Data)
//         if err != nil {
//             return types.ContractResponse{Data: data}, err
//         }
//         return types.ContractResponse{Data: data}, nil
//     }

//     // Setup contract contexts
//     for _, dep := range dependencies {
//         contractContext := buildExecutionContextClassic(dep)
//         if contractContext == nil {
//             return types.ContractResponse{}, fmt.Errorf("could not build dependency execution context for %s", dep.Address)
//         }
//         context.ContractRouter[dep.Address.String()] = contractContext
//     }

//     // Setup self context
//     selfContext := buildExecutionContextClassic(types.ContractDependency{
//         FilePath:    env.Contract.FilePath,
//         Bytecode:    []byte{},
//         CodeHash:    []byte{},
//         StoreKey:    storeKey,
//         StorageType: storageType,
//         SystemDeps:  systemDeps,
//     })
//     if selfContext == nil {
//         return types.ContractResponse{}, fmt.Errorf("could not build execution context for self")
//     }
//     contractstr := env.Contract.Address.String()
//     context.ContractRouter[contractstr] = selfContext

//     // Initialize Wasm module
//     module, _cleanups, err := InitiateWasm(context, env.Contract.FilePath, nil, systemDeps)
//     cleanups = append(cleanups, _cleanups...)
//     if err != nil {
//         return types.ContractResponse{}, err
//     }

//     selfContext.Module = module
//     setExecutionBytecode(context, module, funcName)
//     selfContext.ContractInfo.Bytecode = context.Env.Contract.Bytecode
//     selfContext.ContractInfo.CodeHash = context.Env.Contract.CodeHash

//     executeHandler := GetExecuteFunctionHandler(systemDeps)

//     if inBackground {
//         err = types.AddBackgroundProcesses(goContextParent, &types.BackgroundProcess{
//             Label:   contractstr,
//             Module:  module,
//             ExecuteHandler: func(funcName_ string) ([]byte, error) {
//                 _, err := executeHandler(context, module, funcName_, []uint64{})
//                 if err != nil {
//                     return nil, err
//                 }
//                 return context.FinishData, nil
//             },
//         })
//         if err != nil {
//             return types.ContractResponse{}, err
//         }
//     }

//     _, err = executeHandler(context, module, funcName, []uint64{})
//     if err != nil {
//         wrapErr := sdkerr.Wrapf(
//             err,
//             "chain_id: %s; contract: %s; entry point: %s; revert: %s",
//             context.Ctx.ChainID(),
//             context.Env.Contract.Address.String(),
//             funcName,
//             hex.EncodeToString(context.FinishData),
//         )
//         resp := handleContractErrorResponse(module, context.FinishData, isdebug, wrapErr)
//         if isdebug {
//             return resp, nil
//         }
//         return resp, wrapErr
//     }

//     return handleContractResponse(context, module, isdebug), nil
// }

// func setExecutionBytecode(context *Context, module api.Module, funcName string) {
//     if len(context.Env.Contract.Bytecode) > 0 {
//         if funcName == types.ENTRY_POINT_INSTANTIATE {
//             context.Env.Contract.Bytecode = append(context.Env.Contract.Bytecode, context.Env.CurrentCall.CallData...)
//         }

//         runtimeLen, err := hex.DecodeString(fmt.Sprintf("%064x", len(context.Env.Contract.Bytecode)))
//         if err != nil {
//             return
//         }
//         calldLen, err := hex.DecodeString(fmt.Sprintf("%064x", len(context.Env.CurrentCall.CallData)))
//         if err != nil {
//             return
//         }
//         context.Env.CurrentCall.CallData = append(append(append(
//             runtimeLen,
//             context.Env.Contract.Bytecode...),
//             calldLen...),
//             context.Env.CurrentCall.CallData...,
//         )
//         return
//     }

//     // Get bytecode from exported function if available
//     if fn := module.ExportedFunction("evm_bytecode"); fn != nil {
//         results, err := fn.Call(context.GoContextParent)
//         if err != nil {
//             return
//         }

//         memoffset := uint32(results[0])
//         constructorLength := uint32(results[1])
//         runtimeLength := uint32(results[2])

//         memory := module.Memory()
//         if memory == nil {
//             return
//         }

//         executionBytecode, ok := memory.Read(memoffset+constructorLength, runtimeLength)
//         if !ok {
//             return
//         }

//         if funcName == types.ENTRY_POINT_INSTANTIATE {
//             constructorBytecode, ok := memory.Read(memoffset, constructorLength)
//             if !ok {
//                 return
//             }
//             executionBytecode = append(constructorBytecode, executionBytecode...)
//             executionBytecode = append(executionBytecode, context.Env.CurrentCall.CallData...)
//         }

//         context.Env.Contract.Bytecode = executionBytecode
//     }
// }

// func handleContractResponse(context *Context, module api.Module, isdebug bool) types.ContractResponse {
//     data := context.FinishData
//     logs := context.Logs
//     messages := context.Messages
//     var events []types.Event

//     for i, log := range logs {
//         contractAddressStr := log.ContractAddress.String()
//         attributes := []types.EventAttribute{
//             {Key: types.AttributeKeyIndex, Value: fmt.Sprint(i)},
//             {Key: types.AttributeKeyData, Value: "0x" + hex.EncodeToString(log.Data)},
//             {Key: types.AttributeKeyEventType, Value: log.Type},
//             {Key: types.AttributeKeyCallContractAddress, Value: contractAddressStr},
//             {Key: types.AttributeKeyDependency, Value: log.SystemDependency},
//         }

//         for _, topic := range log.Topics {
//             attributes = append(attributes, types.EventAttribute{
//                 Key:   types.AttributeKeyTopic,
//                 Value: "0x" + hex.EncodeToString(topic[:]),
//             })
//         }

//         events = append(events, types.Event{
//             Type:       types.EventTypeWasmxLog,
//             Attributes: attributes,
//         })
//     }
//     events = append(events, context.CosmosEvents...)

//     var mem []byte
//     if isdebug {
//         mem = getMemory(module)
//     }

//     return types.ContractResponse{
//         Data:           data,
//         Events:         events,
//         MemorySnapshot: mem,
//         Messages:       messages,
//     }
// }

// func handleContractErrorResponse(module api.Module, data []byte, isdebug bool, err error) types.ContractResponse {
//     var mem []byte
//     if isdebug {
//         mem = getMemory(module)
//     }

//     return types.ContractResponse{
//         Data:           data,
//         MemorySnapshot: mem,
//         ErrorMessage:   err.Error(),
//     }
// }

// const MEM_PAGE_SIZE = 64 * 1024 // 64KiB

// func getMemory(module api.Module) []byte {
//     memory := module.Memory()
//     if memory == nil {
//         return nil
//     }

//     data, ok := memory.Read(0, memory.Size())
//     if !ok {
//         return nil
//     }

//     dst := make([]byte, len(data))
//     copy(dst, data)
//     return dst
// }

// func GetVmLogger(
//     logger func(ctx sdk.Context) log.Logger,
//     chainId string,
//     contractAddress string,
// ) func(ctx sdk.Context) log.Logger {
//     return func(ctx sdk.Context) log.Logger {
//         moduleName := fmt.Sprintf("x/%s_%s_%s", types.ModuleName, chainId, contractAddress)
//         return logger(ctx).With(log.ModuleKey, moduleName)
//     }
// }

// func GetVmLoggerExtended(
//     logger func(ctx sdk.Context) log.Logger,
//     chainId string,
//     contractAddress string,
// ) func(ctx sdk.Context) log.Logger {
//     return func(ctx sdk.Context) log.Logger {
//         moduleName := fmt.Sprintf("x/%s_extended_%s_%s", types.ModuleName, chainId, contractAddress)
//         return logger(ctx).With(log.ModuleKey, moduleName)
//     }
// }
