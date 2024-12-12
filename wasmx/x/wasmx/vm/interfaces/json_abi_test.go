package interfaces

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	interfacesTestdata "github.com/loredanacirstea/wasmx/v1/x/wasmx/keeper/testdata/interfaces"
)

func TestInterfaces_JsonToAbi(t *testing.T) {
	Erc20Abi := interfacesTestdata.Erc20Abi

	transferParamsJson := `{"amount":"100","recipient":"0x3defca2d10c7540621fd8ad553e7f987571b712d"}`

	transferParams := "0000000000000000000000003defca2d10c7540621fd8ad553e7f987571b712d0000000000000000000000000000000000000000000000000000000000000064"

	var val map[string]interface{}
	err := json.Unmarshal([]byte(transferParamsJson), &val)
	require.NoError(t, err)

	fabi := Erc20Abi.Methods["transfer"]

	args := make([]interface{}, len(fabi.Inputs))
	for i, inp := range fabi.Inputs {
		_val := val[inp.Name]

		switch inp.Type.GetType() {
		case reflect.TypeOf(common.Address{}):
			inner := _val.(string)
			_val = common.HexToAddress(inner)
		case BigIntType:
			inner := _val.(string)
			var ok bool
			_val, ok = big.NewInt(1).SetString(inner, 10)
			if !ok {
				panic("err")
			}
		}
		args[i] = _val
	}

	packed, err := fabi.Inputs.Pack(args...)
	require.NoError(t, err)

	require.Equal(t, transferParams, hex.EncodeToString(packed))
}
