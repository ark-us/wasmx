package interfaces

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

var BigIntElem = &big.Int{}
var BigIntType = reflect.TypeOf(BigIntElem)

func AbiMarshalJSON(iface interface{}) (string, error) {
	ifv := reflect.ValueOf(iface)
	// ift := reflect.TypeOf(iface)
	// fmt.Println("-AbiMarshalJSON-ifv-", ifv)
	// fmt.Println("-AbiMarshalJSON-ift-", ift)
	// fmt.Println("-AbiMarshalJSON-iface-", iface)
	// fmt.Println("-AbiMarshalJSON-ift.NumField()-", ift.NumField())
	// fmt.Println("-AbiMarshalJSON-ift.Name()-", ift.Name())
	// fmt.Println("-AbiMarshalJSON-ift.Elem().Name()()-", ift.Elem().Name())
	// fmt.Println("-AbiMarshalJSON-ift.Key().Name()()-", ift.Key().Name())

	typ := ifv.Type()
	// fmt.Println("-AbiMarshalJSON-typ-", typ)
	// fmt.Println("-AbiMarshalJSON-typ-", typ.NumField())

	var fields []string

	for i := 0; i < typ.NumField(); i++ {
		structFieldName := typ.Field(i).Name
		fieldName := strings.Join(strings.Split(strings.ToLower(sdkclient.CamelCaseToString(structFieldName)), " "), "_")
		structFieldValue := ifv.Field(i)
		encoded := fmt.Sprintf(`"%s":`, fieldName)

		// fmt.Println("-structFieldName-", structFieldName, fieldName, structFieldValue.Kind())
		if structFieldValue.Kind() == reflect.Struct {
			encoding, err := AbiMarshalJSON(structFieldValue.Interface())
			if err != nil {
				return "", err
			}
			encoded = encoded + string(encoding)
		} else {
			// fmt.Println("----***--structFieldValue.Type()", structFieldValue.Type())
			switch structFieldValue.Type() {
			case reflect.TypeOf(common.Address{}):
				// fmt.Println("----common.Address{}")
				val := structFieldValue.Interface().(common.Address)
				bech32, err := sdk.AccAddressFromHexUnsafe(val.Hex()[2:])
				if err != nil {
					return "", err
				}
				// TODO special structs
				if strings.Contains(fieldName, "validator") {
					validator := sdk.ValAddress(bech32.Bytes())
					encoded = fmt.Sprintf(`%s"%s"`, encoded, validator.String())
				} else {
					encoded = fmt.Sprintf(`%s"%s"`, encoded, bech32.String())
				}
			case BigIntType:
				// fmt.Println("----big.Int{}")
				val := structFieldValue.Interface().(*big.Int)
				encoded = fmt.Sprintf(`%s"%s"`, encoded, sdkmath.NewIntFromBigInt(val).String())
			default:
				// fmt.Println("--default", structFieldValue.Kind(), structFieldValue.Type())
				// common.Address is an array, so the following needs to be last
				if structFieldValue.Kind() == reflect.Slice || structFieldValue.Kind() == reflect.Array {
					// fmt.Println("--array", structFieldValue.Len())
					arrayItems := make([]string, 0)
					for j := 0; j < structFieldValue.Len(); j++ {
						item := structFieldValue.Index(j)
						encoding, err := AbiMarshalJSON(item.Interface())
						if err != nil {
							return "", err
						}
						arrayItems = append(arrayItems, encoding)
					}
					encoded = fmt.Sprintf(`%s[%s]`, encoded, strings.Join(arrayItems, ","))
					// fmt.Println("----array", strings.Join(arrayItems, ","))
				} else {
					v, err := json.Marshal(structFieldValue.Interface())
					if err != nil {
						return "", err
					}
					encoded = fmt.Sprintf(`%s%s`, encoded, string(v))
				}
			}
		}
		fields = append(fields, encoded)
	}

	encoded := fmt.Sprintf("{%s}", strings.Join(fields, ","))
	return encoded, nil
}
