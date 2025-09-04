package wasmxcore

import (
	"encoding/base64"
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func grpcRequestWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(grpcRequest_(utils.BytesToPackedPtr(bz)))
}

func startTimeoutWrap(bz []byte) {
	startTimeout_(utils.BytesToPackedPtr(bz))
}

func cancelTimeoutWrap(bz []byte) {
	cancelTimeout_(utils.BytesToPackedPtr(bz))
}

func startBackgroundProcessWrap(bz []byte) {
	startBackgroundProcess_(utils.BytesToPackedPtr(bz))
}

func writeToBackgroundProcessWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(writeToBackgroundProcess_(utils.BytesToPackedPtr(bz)))
}

func readFromBackgroundProcessWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(readFromBackgroundProcess_(utils.BytesToPackedPtr(bz)))
}

func migrateContractStateByStorageTypeWrap(bz []byte) {
	migrateContractStateByStorageType_(utils.BytesToPackedPtr(bz))
}

func migrateContractStateByAddressWrap(bz []byte) {
	migrateContractStateByAddress_(utils.BytesToPackedPtr(bz))
}

func storageLoadGlobalWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(storageLoadGlobal_(utils.BytesToPackedPtr(bz)))
}

func storageStoreGlobalWrap(bz []byte) {
	storageStoreGlobal_(utils.BytesToPackedPtr(bz))
}

func storageDeleteGlobalWrap(bz []byte) {
	storageDeleteGlobal_(utils.BytesToPackedPtr(bz))
}

func storageHasGlobalWrap(bz []byte) int32 {
	return storageHasGlobal_(utils.BytesToPackedPtr(bz))
}

func storageResetGlobalWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(storageResetGlobal_(utils.BytesToPackedPtr(bz)))
}

func updateSystemCacheWrap(bz []byte) []byte {
	return utils.PackedPtrToBytes(updateSystemCache_(utils.BytesToPackedPtr(bz)))
}

func GrpcRequest(ipAddress string, contractAddress wasmx.Bech32String, data string) (*GrpcResponse, error) {
	request := map[string]string{
		"ip_address": ipAddress,
		"contract":   string(contractAddress),
		"data":       data,
	}
	reqJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	LoggerDebugExtended("grpc request: ", []string{"request", string(reqJSON)})

	result := grpcRequestWrap(reqJSON)

	LoggerDebugExtended("grpc request: ", []string{"response", string(result)})

	var response GrpcResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}

	if response.Error == "" {
		decodedData, err := base64.StdEncoding.DecodeString(response.Data)
		if err != nil {
			return nil, err
		}
		response.Data = string(decodedData)
	}

	return &response, nil
}

func StartTimeout(id string, contract wasmx.Bech32String, delayMS int64, args []byte) error {
	req := StartTimeoutRequest{
		ID:       id,
		Contract: contract,
		Delay:    delayMS,
		Args:     args,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	startTimeoutWrap(reqJSON)
	return nil
}

func CancelTimeout(id string) error {
	req := CancelTimeoutRequest{ID: id}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	cancelTimeoutWrap(reqJSON)
	return nil
}

func StartBackgroundProcess(contract string, args []byte) error {
	msg := map[string][]byte{"data": args}
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req := StartBackgroundProcessRequest{
		Contract: contract,
		Args:     msgJSON,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	startBackgroundProcessWrap(reqJSON)
	return nil
}

func WriteToBackgroundProcess(contract, ptrFunc string, data []byte) (*WriteToBackgroundProcessResponse, error) {
	req := WriteToBackgroundProcessRequest{
		Contract: contract,
		Data:     data,
		PtrFunc:  ptrFunc,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result := writeToBackgroundProcessWrap(reqJSON)

	var response WriteToBackgroundProcessResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func ReadFromBackgroundProcess(contract wasmx.Bech32String, ptrFunc, lenFunc string) (*ReadFromBackgroundProcessResponse, error) {
	req := ReadFromBackgroundProcessRequest{
		Contract: contract,
		PtrFunc:  ptrFunc,
		LenFunc:  lenFunc,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result := readFromBackgroundProcessWrap(reqJSON)

	var response ReadFromBackgroundProcessResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func MigrateContractStateByStorageType(req MigrateContractStateByStorageRequest) error {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	migrateContractStateByStorageTypeWrap(reqJSON)
	return nil
}

func MigrateContractStateByAddress(req MigrateContractStateByAddressRequest) error {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	migrateContractStateByAddressWrap(reqJSON)
	return nil
}

func StorageLoadGlobal(req GlobalStorageLoadRequest) ([]byte, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result := storageLoadGlobalWrap(reqJSON)
	return result, nil
}

func StorageStoreGlobal(req GlobalStorageStoreRequest) error {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	storageStoreGlobalWrap(reqJSON)
	return nil
}

func StorageDeleteGlobal(req GlobalStorageLoadRequest) error {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	storageDeleteGlobalWrap(reqJSON)
	return nil
}

func StorageHasGlobal(req GlobalStorageLoadRequest) (bool, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	result := storageHasGlobalWrap(reqJSON)
	return result != 0, nil
}

func StorageResetGlobal(req GlobalStorageResetRequest) (*GlobalStorageResetResponse, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result := storageResetGlobalWrap(reqJSON)

	var response GlobalStorageResetResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func UpdateSystemCache(req UpdateSystemCacheRequest) (*UpdateSystemCacheResponse, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	LoggerInfo("update system cache: ", []string{"data", string(reqJSON)})

	result := updateSystemCacheWrap(reqJSON)

	LoggerInfo("update system cache: ", []string{"data", string(reqJSON), "host_response", string(result)})

	var response UpdateSystemCacheResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
