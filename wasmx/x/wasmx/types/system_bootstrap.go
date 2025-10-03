package types

import (
	"context"
	"fmt"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type SystemBootstrapContextKey string

const SystemBootstrapKey SystemBootstrapContextKey = "SystemBootstrapData"

type SystemBootstrap struct {
	RoleAddress              mcodec.AccAddressPrefixed `json:"role_address"`
	RoleRegistryId           uint64                    `json:"role_registry_id"`
	RoleRegistryCodeInfo     *CodeInfo                 `json:"role_registry_code_info"`
	RoleRegistryContractInfo *ContractInfo             `json:"role_registry_contract_info"`
	CodeRegistryAddress      mcodec.AccAddressPrefixed `json:"code_registry_address"`
	CodeRegistryId           uint64                    `json:"code_registry_id"`
	CodeRegistryCodeInfo     *CodeInfo                 `json:"code_registry_code_info"`
	CodeRegistryContractInfo *ContractInfo             `json:"code_registry_contract_info"`
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

func NewSystemBootstrap(roleAddress mcodec.AccAddressPrefixed, roleCodeId uint64, roleCodeInfo CodeInfo, roleContractInfo ContractInfo, scAddress mcodec.AccAddressPrefixed, scCodeId uint64, scCodeInfo CodeInfo, scContractInfo ContractInfo) *SystemBootstrap {
	return &SystemBootstrap{
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
		RoleAddress:         mcodec.AccAddressPrefixed{},
		RoleRegistryId:      0,
		CodeRegistryAddress: mcodec.AccAddressPrefixed{},
		CodeRegistryId:      0,
	}
	newctx := context.WithValue(ctx, SystemBootstrapKey, data)
	return newctx, data
}

func SetSystemBootstrap(ctx context.Context, newdata *SystemBootstrap) error {
	datai := ctx.Value(SystemBootstrapKey)
	data, ok := (datai).(*SystemBootstrap)
	if !ok {
		return fmt.Errorf("SystemBootstrap invalid type")
	}
	if data == nil {
		return fmt.Errorf("SystemBootstrap not set on context")
	}
	data.RoleAddress = newdata.RoleAddress
	data.RoleRegistryId = newdata.RoleRegistryId
	data.RoleRegistryCodeInfo = newdata.RoleRegistryCodeInfo
	data.RoleRegistryContractInfo = newdata.RoleRegistryContractInfo
	data.CodeRegistryAddress = newdata.CodeRegistryAddress
	data.CodeRegistryId = newdata.CodeRegistryId
	data.CodeRegistryCodeInfo = newdata.CodeRegistryCodeInfo
	data.CodeRegistryContractInfo = newdata.CodeRegistryContractInfo
	return nil
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
