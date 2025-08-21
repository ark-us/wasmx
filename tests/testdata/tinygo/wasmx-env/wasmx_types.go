package wasmx

import (
    "encoding/base64"
)

const TypeUrl_MsgExecuteContract = "/mythos.wasmx.v1.MsgExecuteContract"

// AnyWrap is a type_url + base64-encoded value
type AnyWrap struct {
    TypeURL string `json:"type_url"`
    Value   string `json:"value"`
}

func NewAnyWrap(typeUrl string, value string) AnyWrap {
    return AnyWrap{TypeURL: typeUrl, Value: base64.StdEncoding.EncodeToString([]byte(value))}
}

// MsgExecuteContract embeds a message for wasmx execution.
type MsgExecuteContract struct {
    Sender       Bech32String         `json:"sender"`
    Contract     Bech32String         `json:"contract"`
    Msg          WasmxExecutionMessage `json:"msg"`
    Funds        []Coin               `json:"funds"`
    Dependencies []string             `json:"dependencies"`
}

func (MsgExecuteContract) TypeUrl() string { return TypeUrl_MsgExecuteContract }

// PrefixedAddress holds an address with chain-specific prefix
type PrefixedAddress struct {
    Bz     string `json:"bz"`     // base64 bytes
    Prefix string `json:"prefix"`
}

