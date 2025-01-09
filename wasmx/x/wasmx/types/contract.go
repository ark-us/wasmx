package types

import (
	"encoding/json"
	"errors"
)

type CodeOrigin struct {
	ChainId string `json:"chain_id"`
	Address string `json:"address"`
}

type CodeMetadata struct {
	Name       string      `json:"name"`
	Categ      []string    `json:"categ"`
	Icon       string      `json:"icon"`
	Author     string      `json:"author"`
	Site       string      `json:"site"`
	Abi        string      `json:"abi"`
	JsonSchema string      `json:"json_schema"`
	Origin     *CodeOrigin `json:"origin"`
}

type CodeInfo struct {
	CodeHash                      Checksum           `json:"code_hash"`
	Creator                       string             `json:"creator"`
	Deps                          []string           `json:"deps"`
	Pinned                        bool               `json:"pinned"`
	MeteringOff                   bool               `json:"metering_off"`
	Metadata                      CodeMetadata       `json:"metadata"`
	InterpretedBytecodeDeployment RawContractMessage `json:"interpreted_bytecode_deployment"`
	InterpretedBytecodeRuntime    RawContractMessage `json:"interpreted_bytecode_runtime"`
	RuntimeHash                   Checksum           `json:"runtime_hash"`
}

type ContractInfo struct {
	CodeId      uint64              `json:"code_id"`
	Creator     string              `json:"creator"`
	Label       string              `json:"label"`
	StorageType ContractStorageType `json:"storage_type"`
	InitMessage RawContractMessage  `json:"init_message"`
	Provenance  string              `json:"provenance"`
	IbcPortId   string              `json:"ibc_port_id"`
}

type ContractInstance struct {
	CodeInfo     *CodeInfo     `json:"code_info"`
	ContractInfo *ContractInfo `json:"contract_info"`
}

type QueryLastCodeIdResponse struct {
	CodeId uint64 `json:"code_id"`
}

type MsgSetContractInfoRequest struct {
	Address      string       `json:"address"`
	ContractInfo ContractInfo `json:"contract_info"`
}

type GetContractInfoResponse struct {
	ContractInfo *ContractInfo `json:"contract_info"`
}

type GetCodeInfoResponse struct {
	CodeInfo *CodeInfo `json:"code_info"`
}

type GenesisRegistryContract struct {
	CodeInfos     []CodeInfo                  `json:"code_infos"`
	ContractInfos []MsgSetContractInfoRequest `json:"contract_infos"`
}

func (v *CodeOrigin) ToProto() *CodeOriginPB {
	if v == nil {
		return nil
	}
	return &CodeOriginPB{
		Address: v.Address,
		ChainId: v.ChainId,
	}
}

func (v *CodeOriginPB) ToJson() *CodeOrigin {
	if v == nil {
		return &CodeOrigin{}
	}
	return &CodeOrigin{
		Address: v.Address,
		ChainId: v.ChainId,
	}
}

func (v CodeMetadata) ToProto() CodeMetadataPB {
	return CodeMetadataPB{
		Name:       v.Name,
		Categ:      v.Categ,
		Icon:       v.Icon,
		Author:     v.Author,
		Site:       v.Site,
		Abi:        v.Abi,
		JsonSchema: v.JsonSchema,
		Origin:     v.Origin.ToProto(),
	}
}

func (v CodeMetadataPB) ToJson() CodeMetadata {
	return CodeMetadata{
		Name:       v.Name,
		Categ:      v.Categ,
		Icon:       v.Icon,
		Author:     v.Author,
		Site:       v.Site,
		Abi:        v.Abi,
		JsonSchema: v.JsonSchema,
		Origin:     v.Origin.ToJson(),
	}
}

func (v *CodeInfo) ToProto() *CodeInfoPB {
	if v == nil {
		return nil
	}
	return &CodeInfoPB{
		CodeHash:                      v.CodeHash,
		Creator:                       v.Creator,
		Deps:                          v.Deps,
		Pinned:                        v.Pinned,
		MeteringOff:                   v.MeteringOff,
		Metadata:                      v.Metadata.ToProto(),
		InterpretedBytecodeDeployment: v.InterpretedBytecodeDeployment,
		InterpretedBytecodeRuntime:    v.InterpretedBytecodeRuntime,
		RuntimeHash:                   v.RuntimeHash,
	}
}

func (v *CodeInfoPB) ToJson() *CodeInfo {
	if v == nil {
		return &CodeInfo{}
	}
	return &CodeInfo{
		CodeHash:                      v.CodeHash,
		Creator:                       v.Creator,
		Deps:                          v.Deps,
		Pinned:                        v.Pinned,
		MeteringOff:                   v.MeteringOff,
		Metadata:                      v.Metadata.ToJson(),
		InterpretedBytecodeDeployment: v.InterpretedBytecodeDeployment,
		InterpretedBytecodeRuntime:    v.InterpretedBytecodeRuntime,
		RuntimeHash:                   v.RuntimeHash,
	}
}

func (v *ContractInfo) ToProto() *ContractInfoPB {
	if v == nil {
		return nil
	}
	return &ContractInfoPB{
		CodeId:      v.CodeId,
		Creator:     v.Creator,
		Label:       v.Label,
		StorageType: v.StorageType,
		InitMessage: v.InitMessage,
		Provenance:  v.Provenance,
		IbcPortId:   v.IbcPortId,
	}
}

func (v *ContractInfoPB) ToJson() *ContractInfo {
	if v == nil {
		return &ContractInfo{}
	}
	return &ContractInfo{
		CodeId:      v.CodeId,
		Creator:     v.Creator,
		Label:       v.Label,
		StorageType: v.StorageType,
		InitMessage: v.InitMessage,
		Provenance:  v.Provenance,
		IbcPortId:   v.IbcPortId,
	}
}

func (v ContractStorageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ContractStorageType_name[int32(v)])
}

func (v *ContractStorageType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if val, ok := ContractStorageType_value[s]; ok {
		*v = ContractStorageType(val)
		return nil
	}

	return errors.New("invalid ContractStorageType: " + s)
}
