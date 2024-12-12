package config

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type ExecutionMetaInfoContextKey string

const ExecutionMetaInfoKey ExecutionMetaInfoContextKey = "ExecutionMetaInfo"

type ExecutionMetaInfo struct {
	Data     map[string]interface{}
	TempData map[string]interface{}
}

func NewExecutionMetaInfo(data map[string]interface{}) *ExecutionMetaInfo {
	return &ExecutionMetaInfo{
		Data:     data,
		TempData: map[string]interface{}{},
	}
}

func WithExecutionMetaInfoEmpty(ctx context.Context) (context.Context, *ExecutionMetaInfo) {
	data := &ExecutionMetaInfo{Data: map[string]interface{}{}, TempData: map[string]interface{}{}}
	return context.WithValue(ctx, ExecutionMetaInfoKey, data), data
}

func WithExecutionMetaInfo(ctx context.Context, data *ExecutionMetaInfo) context.Context {
	return context.WithValue(ctx, ExecutionMetaInfoKey, data)
}

func SetExecutionMetaInfo(ctx context.Context, cdc codec.Codec, metainfo map[string][]byte) error {
	datai := ctx.Value(ExecutionMetaInfoKey)
	data, ok := (datai).(*ExecutionMetaInfo)
	if !ok {
		return fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	if data == nil {
		return fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	for key, value := range metainfo {
		anymsg, err := mcodec.AnyFromBzJson(cdc, value)
		if err != nil {
			return err
		}

		var msg sdk.Msg
		err = cdc.UnpackAny(&anymsg, &msg)
		if err != nil {
			return err
		}
		data.Data[key] = msg
	}
	return nil
}

func GetExecutionMetaInfo(ctx context.Context) (*ExecutionMetaInfo, error) {
	datai := ctx.Value(ExecutionMetaInfoKey)
	data, ok := (datai).(*ExecutionMetaInfo)
	if !ok {
		return nil, fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	if data == nil {
		return nil, fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	return data, nil
}

func GetExecutionMetaInfoEncoded(ctx context.Context, cdc codec.Codec) (map[string][]byte, error) {
	metainfo := map[string][]byte{}
	data, err := GetExecutionMetaInfo(ctx)
	if err != nil {
		return metainfo, err
	}
	for key, value := range data.Data {
		sdkmsg := value.(sdk.Msg)
		anymsg, err := codectypes.NewAnyWithValue(sdkmsg)
		if err != nil {
			return metainfo, err
		}
		anybz, err := cdc.MarshalJSON(anymsg)
		if err != nil {
			return metainfo, err
		}
		metainfo[key] = anybz
	}
	return metainfo, nil
}

func ResetExecutionMetaInfo(ctx context.Context) {
	datai := ctx.Value(ExecutionMetaInfoKey)
	data, ok := (datai).(*ExecutionMetaInfo)
	if ok && data != nil {
		data.Data = map[string]interface{}{}
	}
}
