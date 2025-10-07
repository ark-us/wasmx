package types

import (
	"context"
	"fmt"
	"sync"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type SystemBootstrapContextKey string

const SystemBootstrapKey SystemBootstrapContextKey = "SystemBootstrapData"

type SystemBootstrapPerChain struct {
	RoleAddress              mcodec.AccAddressPrefixed `json:"role_address"`
	RoleRegistryId           uint64                    `json:"role_registry_id"`
	RoleRegistryCodeInfo     *CodeInfo                 `json:"role_registry_code_info"`
	RoleRegistryContractInfo *ContractInfo             `json:"role_registry_contract_info"`
	CodeRegistryAddress      mcodec.AccAddressPrefixed `json:"code_registry_address"`
	CodeRegistryId           uint64                    `json:"code_registry_id"`
	CodeRegistryCodeInfo     *CodeInfo                 `json:"code_registry_code_info"`
	CodeRegistryContractInfo *ContractInfo             `json:"code_registry_contract_info"`
}

type SystemBootstrap struct {
	data    map[string]SystemBootstrapPerChain
	mtxData sync.Mutex
}

func (v *SystemBootstrap) GetData(chainId string) (*SystemBootstrapPerChain, bool) {
	v.mtxData.Lock()
	defer v.mtxData.Unlock()
	data, found := v.data[chainId]
	return &data, found
}

func (v *SystemBootstrap) SetData(chainId string, data SystemBootstrapPerChain) {
	v.mtxData.Lock()
	defer v.mtxData.Unlock()
	v.data[chainId] = data
}

func NewSystemBootstrapData(roleAddress string, roleCodeId uint64, roleCodeInfo CodeInfoPB, roleContractInfo ContractInfoPB, scAddress string, scCodeId uint64, scCodeInfo CodeInfoPB, scContractInfo ContractInfoPB) *SystemBootstrapData {
	return &SystemBootstrapData{
		RoleAddress:              roleAddress,
		RoleRegistryId:           roleCodeId,
		RoleRegistryCodeInfo:     &roleCodeInfo,
		RoleRegistryContractInfo: &roleContractInfo,
		CodeRegistryAddress:      scAddress,
		CodeRegistryId:           scCodeId,
		CodeRegistryCodeInfo:     &scCodeInfo,
		CodeRegistryContractInfo: &scContractInfo,
	}
}

func WithSystemBootstrap(ctx context.Context) (context.Context, *SystemBootstrap) {
	data := &SystemBootstrap{
		data: map[string]SystemBootstrapPerChain{},
	}
	newctx := context.WithValue(ctx, SystemBootstrapKey, data)
	return newctx, data
}

func GetSystemBootstrap(ctx context.Context) (*SystemBootstrap, error) {
	datai := ctx.Value(SystemBootstrapKey)
	data, ok := (datai).(*SystemBootstrap)
	if !ok {
		return nil, fmt.Errorf("SystemBootstrap invalid type")
	}
	if data == nil {
		return nil, fmt.Errorf("SystemBootstrap not set on context")
	}
	return data, nil
}
