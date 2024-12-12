package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
)

// MsgFromBz
func (k *Keeper) MsgFromBz(content []byte) (sdk.Msg, error) {
	msg, err := mcodec.MsgFromBz(k.cdc, content)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// RequestQueryFromBz
func (k *Keeper) RequestQueryFromBz(content []byte) (abci.RequestQuery, error) {
	msg, err := mcodec.RequestQueryFromBz(content)
	if err != nil {
		return abci.RequestQuery{}, err
	}
	return msg, nil
}

// // QueryFromBz
// func (k *Keeper) QueryFromBz(content []byte) (types.Query, error) {
// 	var queryMsg types.Query

// 	if err := k.cdc.UnmarshalInterface(content, &queryMsg); err != nil {
// 		return queryMsg, errors.Wrap(err, "error unmarshalling query")
// 	}

// 	return queryMsg, nil
// }

// AnyFromBz
func (k *Keeper) AnyFromBz(bz []byte) (cdctypes.Any, error) {
	msg, err := mcodec.AnyFromBz(k.cdc, bz)
	if err != nil {
		return cdctypes.Any{}, err
	}
	return msg, nil
}

// ConvertProtoToJSONMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshalling.
func (k *Keeper) ConvertProtoToJSONMarshal(protoResponse codec.ProtoMarshaler, bz []byte) ([]byte, error) {
	msg, err := mcodec.ConvertProtoToJSONMarshal(k.cdc, protoResponse, bz)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
