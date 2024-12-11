package codec

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/cosmos/gogoproto/proto"

	sdkerr "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// typeUrl:   /cosmos.auth.v1beta1.QueryAccountRequest
// queryPath: /cosmos.auth.v1beta1.Query/Account
func GetPathFromType(typeUrl string) (string, error) {
	parts := strings.Split(typeUrl, ".")
	lastIndex := len(parts) - 1
	lastParts := strings.Split(parts[lastIndex], "Query")
	if len(lastParts) == 0 {
		return "", fmt.Errorf("wrong query format: %s", typeUrl)
	}
	queryName := strings.Split(lastParts[1], "Request")[0]

	path := strings.Join(parts[:lastIndex], ".") + ".Query/" + queryName
	return path, nil
}

// MsgFromBz
func MsgFromBz(cdc codec.Codec, content []byte) (sdk.Msg, error) {
	var txMsg sdk.Msg
	if err := cdc.UnmarshalInterface(content, &txMsg); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling sdk msg json")
	}
	return txMsg, nil
}

// RequestQueryFromBz
func RequestQueryFromBz(content []byte) (abci.RequestQuery, error) {
	var queryMsg abci.RequestQuery
	err := queryMsg.Unmarshal(content)
	if err != nil {
		return queryMsg, errors.Wrap(err, "error unmarshalling abci.RequestQuery")
	}
	return queryMsg, nil
}

// // QueryFromBz
// func QueryFromBz(cdc codec.Codec, content []byte) (types.Query, error) {
// 	var queryMsg types.Query
// 	if err := cdc.UnmarshalInterface(content, &queryMsg); err != nil {
// 		return queryMsg, errors.Wrap(err, "error unmarshalling query")
// 	}
// 	return queryMsg, nil
// }

// AnyFromBz
func AnyFromBz(cdc codec.Codec, bz []byte) (cdctypes.Any, error) {
	anyMsg := &cdctypes.Any{}
	err := cdc.Unmarshal(bz, anyMsg)
	if err != nil {
		return *anyMsg, err
	}
	return *anyMsg, nil
}

func AnyFromBzJson(cdc codec.Codec, bz []byte) (cdctypes.Any, error) {
	anyMsg := &cdctypes.Any{}
	err := cdc.UnmarshalJSON(bz, anyMsg)
	if err != nil {
		return *anyMsg, err
	}
	return *anyMsg, nil
}

// ConvertProtoToJSONMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshalling.
func ConvertProtoToJSONMarshal(cdc codec.Codec, protoResponse codec.ProtoMarshaler, bz []byte) ([]byte, error) {
	// unmarshal binary into stargate response data structure
	err := cdc.Unmarshal(bz, protoResponse)
	if err != nil {
		return nil, sdkerr.Wrap(err, "to proto")
	}
	bz, err = cdc.MarshalJSON(protoResponse)
	if err != nil {
		return nil, sdkerr.Wrap(err, "to json")
	}
	return bz, nil
}

func AnyToSdkMsg(cdc codec.Codec, anymsg *cdctypes.Any) (sdk.Msg, error) {
	sdkmsg, err := cdc.InterfaceRegistry().Resolve(anymsg.TypeUrl)
	if err != nil {
		return nil, err
	}
	err = cdc.Unmarshal(anymsg.Value, sdkmsg)
	if err != nil {
		return nil, err
	}
	return sdkmsg, nil
}

func TxMsgDataFromBz(bz []byte) (*sdk.TxMsgData, error) {
	var txmsgdata sdk.TxMsgData
	err := proto.Unmarshal(bz, &txmsgdata)
	if err != nil {
		return nil, err
	}
	return &txmsgdata, nil
}
