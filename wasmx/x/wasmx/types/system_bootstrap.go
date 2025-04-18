package types

import (
	"context"
	"fmt"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type SystemBootstrapContextKey string

const SystemBootstrapKey SystemBootstrapContextKey = "SystemBootstrapData"

type SystemBootstrap struct {
	RoleAddress              mcodec.AccAddressPrefixed
	CodeRegistryAddress      mcodec.AccAddressPrefixed
	CodeRegistryId           uint64
	CodeRegistryCodeInfo     *CodeInfo
	CodeRegistryContractInfo *ContractInfo
}

func NewSystemBootstrapData(roleAddress string, scAddress string, scCodeId uint64, scCodeInfo CodeInfoPB, scContractInfo ContractInfoPB) *SystemBootstrapData {
	return &SystemBootstrapData{
		RoleAddress:              roleAddress,
		CodeRegistryAddress:      scAddress,
		CodeRegistryId:           scCodeId,
		CodeRegistryCodeInfo:     &scCodeInfo,
		CodeRegistryContractInfo: &scContractInfo,
	}
}

func NewSystemBootstrap(roleAddress mcodec.AccAddressPrefixed, scAddress mcodec.AccAddressPrefixed, scCodeId uint64, scCodeInfo CodeInfo, scContractInfo ContractInfo) *SystemBootstrap {
	return &SystemBootstrap{
		RoleAddress:              roleAddress,
		CodeRegistryAddress:      scAddress,
		CodeRegistryId:           scCodeId,
		CodeRegistryCodeInfo:     &scCodeInfo,
		CodeRegistryContractInfo: &scContractInfo,
	}
}

func WithSystemBootstrap(ctx context.Context) (context.Context, *SystemBootstrap) {
	data := &SystemBootstrap{
		RoleAddress:         mcodec.AccAddressPrefixed{},
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
