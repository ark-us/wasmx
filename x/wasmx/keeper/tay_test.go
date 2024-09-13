package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestInterpreterTaySimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(SimpleStorageTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "SimpleContractTay", nil)

	key := []byte("hello")
	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("sammy"), value)

	data = []byte(`{"get":{"key":"hello"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("sammy"), resp)
}

var SimpleStorageTay = `(do
(def! replaceAll (fn* (sourcestr oldstr newstr)
  (apply str
    (map (fn* (s) (if (= s oldstr) newstr s)) (seq sourcestr))
  )
))
(def! json-to-hm (fn* (jsonstr)
  (eval (read-string (replaceAll (replaceAll jsonstr ":" " ") "," " ")))
))
(def! store (fn* (key val) (do (println "store" key val) (wasmx-storageStore (string-encode key) (string-encode val)))))
(def! load (fn* (key) (wasmx-storageLoad (string-encode key))))
(def! instantiate (fn* (obj)
	(println "instantiated")
))
(def! main (fn* (obj) (do
   (println "main" obj)
   (println "main set?" (contains? obj "set"))
   (println "main get?" (contains? obj "get"))
  (if (contains? obj "set")
    (let* (
        params (get obj "set")
        ff (println "set params" params)
      )
      (do
        (println "set" params)
        (store (get params "key") (get params "value"))
        (wasmx-finish b[])
      )
    )
    (if (contains? obj "get")
      (let* (
          params (get obj "get")
          ff (println "get params" params)
        )
        (wasmx-finish (load (get params "key")))
      )
      (throw "function not found")
    )
  )
)))
(println "before main calld")
(let* (
  calld (string-decode (wasmx-getCallData))
  dfd (println "calld" calld)
)
  (main (json-to-hm calld))
)
)`
