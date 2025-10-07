package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

type ExecutionMetaInfoContextKey string

const ExecutionMetaInfoKey ExecutionMetaInfoContextKey = "ExecutionMetaInfo"

type ExecutionMetaInfo struct {
	data        map[string]interface{}
	mtxData     sync.Mutex
	tempData    map[string]interface{}
	mtxTempData sync.Mutex
}

func (m *ExecutionMetaInfo) GetData(key string) (interface{}, bool) {
	m.mtxData.Lock()
	defer m.mtxData.Unlock()
	v, found := m.data[key]
	return v, found
}

func (m *ExecutionMetaInfo) SetData(key string, value interface{}) {
	m.mtxData.Lock()
	defer m.mtxData.Unlock()
	m.data[key] = value
}

func (m *ExecutionMetaInfo) GetTempData(key string) (interface{}, bool) {
	m.mtxTempData.Lock()
	defer m.mtxTempData.Unlock()
	v, found := m.tempData[key]
	return v, found
}

func (m *ExecutionMetaInfo) SetTempData(key string, value interface{}) {
	m.mtxTempData.Lock()
	defer m.mtxTempData.Unlock()
	m.tempData[key] = value
}

func NewExecutionMetaInfo(data map[string]interface{}) *ExecutionMetaInfo {
	return &ExecutionMetaInfo{
		data:     data,
		tempData: map[string]interface{}{},
	}
}

func WithExecutionMetaInfoEmpty(ctx context.Context) (context.Context, *ExecutionMetaInfo) {
	data := &ExecutionMetaInfo{data: map[string]interface{}{}, tempData: map[string]interface{}{}}
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
		data.SetData(key, msg)
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
	data.mtxData.Lock()
	defer data.mtxData.Unlock()
	for key, value := range data.data {
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
		data.mtxData.Lock()
		defer data.mtxData.Unlock()
		data.data = map[string]interface{}{}
	}
}
