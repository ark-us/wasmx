package wasmxcore

// #include <stdlib.h>
import "C"

// Core environment host API imports

//go:wasmimport wasmxcore grpcRequest
func grpcRequest_(dataPtr int64) int64

//go:wasmimport wasmxcore startTimeout
func startTimeout_(reqPtr int64)

//go:wasmimport wasmxcore cancelTimeout
func cancelTimeout_(reqPtr int64)

//go:wasmimport wasmxcore startBackgroundProcess
func startBackgroundProcess_(reqPtr int64)

//go:wasmimport wasmxcore writeToBackgroundProcess
func writeToBackgroundProcess_(reqPtr int64) int64

//go:wasmimport wasmxcore readFromBackgroundProcess
func readFromBackgroundProcess_(reqPtr int64) int64

//go:wasmimport wasmxcore externalCall
func externalCall_(dataPtr int64) int64

//go:wasmimport wasmxcore migrateContractStateByStorageType
func migrateContractStateByStorageType_(dataPtr int64)

//go:wasmimport wasmxcore migrateContractStateByAddress
func migrateContractStateByAddress_(dataPtr int64)

//go:wasmimport wasmxcore storageLoadGlobal
func storageLoadGlobal_(reqPtr int64) int64

//go:wasmimport wasmxcore storageStoreGlobal
func storageStoreGlobal_(addressPtr int64)

//go:wasmimport wasmxcore storageDeleteGlobal
func storageDeleteGlobal_(addressPtr int64)

//go:wasmimport wasmxcore storageHasGlobal
func storageHasGlobal_(addressPtr int64) int32

//go:wasmimport wasmxcore storageResetGlobal
func storageResetGlobal_(addressPtr int64) int64

//go:wasmimport wasmxcore updateSystemCache
func updateSystemCache_(reqPtr int64) int64
