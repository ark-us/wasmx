package keeper

import (
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// MsgFromBz
func (k Keeper) MsgFromBz(content []byte) (sdk.Msg, error) {
	var txMsg sdk.Msg

	if err := k.cdc.UnmarshalInterface(content, &txMsg); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling sdk msg json")
	}

	return txMsg, nil
}

// RequestQueryFromBz
func (k Keeper) RequestQueryFromBz(content []byte) (abci.RequestQuery, error) {
	var queryMsg abci.RequestQuery

	err := queryMsg.Unmarshal(content)

	if err != nil {
		return queryMsg, errors.Wrap(err, "error unmarshalling abci.RequestQuery")
	}

	return queryMsg, nil
}

// // QueryFromBz
// func (k Keeper) QueryFromBz(content []byte) (types.Query, error) {
// 	var queryMsg types.Query

// 	if err := k.cdc.UnmarshalInterface(content, &queryMsg); err != nil {
// 		return queryMsg, errors.Wrap(err, "error unmarshalling query")
// 	}

// 	return queryMsg, nil
// }

// AnyFromBz
func (k Keeper) AnyFromBz(bz []byte) (cdctypes.Any, error) {
	anyMsg := &cdctypes.Any{}
	err := k.cdc.Unmarshal(bz, anyMsg)
	if err != nil {
		return *anyMsg, err
	}

	return *anyMsg, nil
}

// ConvertProtoToJSONMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshalling.
func (k Keeper) ConvertProtoToJSONMarshal(protoResponse codec.ProtoMarshaler, bz []byte) ([]byte, error) {
	// unmarshal binary into stargate response data structure
	err := k.cdc.Unmarshal(bz, protoResponse)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "to proto")
	}

	bz, err = k.cdc.MarshalJSON(protoResponse)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "to json")
	}

	return bz, nil
}
