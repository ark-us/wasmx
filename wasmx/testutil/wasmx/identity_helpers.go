package wasmx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// RegisterIdentityUser registers the sender's address with wasmx-identity and returns user_id.
func RegisterIdentityUser(app *AppContext, sender simulation.Account, address string, pubKeyBytes []byte) string {
	identityAddr, found := app.App.WasmxKeeper.GetContractAddressByRoleInitial(app.Context(), types.ROLE_ACCOUNT_IDENTITY)
	app.S.Require().True(found, "identity contract not found")

	if address == "" {
		address = app.BytesToAccAddressPrefixed(sender.Address).String()
	}

	registerMsg := types.WasmxExecutionMessage{
		Data: []byte(fmt.Sprintf(`{"register_user":{"address":"%s","public_key":"%s"}}`, address, base64.StdEncoding.EncodeToString(pubKeyBytes))),
	}
	res := app.ExecuteContract(sender, identityAddr, registerMsg, nil, nil)

	var out struct {
		UserID string `json:"user_id"`
	}
	err := app.DecodeExecuteResponse(res, &out)
	app.S.Require().NoError(err)
	app.S.Require().NotEmpty(out.UserID)
	return out.UserID
}

// BuildIdentityRegisterTx builds a signed register_user transaction for wasmx-identity.
func BuildIdentityRegisterTx(app *AppContext, privKey *secp256k1.PrivKey, address string, pubKeyBytes []byte) []byte {
	identityAddr := app.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ACCOUNT_IDENTITY))

	registerUserMsg := map[string]interface{}{
		"register_user": map[string]interface{}{
			"address":        address,
			"public_key":     base64.StdEncoding.EncodeToString(pubKeyBytes),
			"service_domain": "",
			"permissions":    []string{},
			"expires_at":     0,
		},
	}
	innerMsgBytes, err := json.Marshal(registerUserMsg)
	app.S.Require().NoError(err)
	wrappedMsg := map[string]string{
		"data": base64.StdEncoding.EncodeToString(innerMsgBytes),
	}
	msgBytes, err := json.Marshal(wrappedMsg)
	app.S.Require().NoError(err)

	executeMsg := &types.MsgExecuteContract{
		Sender:       address,
		Contract:     identityAddr.String(),
		Msg:          nil,
		Funds:        sdk.NewCoins(),
		Dependencies: []string{},
	}
	wasmxMsg := types.WasmxExecutionMessage{Data: msgBytes}
	executeMsg.Msg, err = json.Marshal(wasmxMsg)
	app.S.Require().NoError(err)

	account := simulation.Account{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Address: sdk.AccAddress(privKey.PubKey().Address()),
	}
	accountAddr := app.BytesToAccAddressPrefixed(account.Address).String()
	if address != "" {
		app.S.Require().Equal(address, accountAddr, "address must match private key")
	}

	txBytes := app.PrepareCosmosTx(account, []sdk.Msg{executeMsg}, nil, nil, "")
	app.S.Require().NotEmpty(txBytes)
	return txBytes
}

// QueryIdentityUserID queries wasmx-identity by address and returns user_id.
func QueryIdentityUserID(app *AppContext, sender simulation.Account, address string) string {
	identityAddr, found := app.App.WasmxKeeper.GetContractAddressByRoleInitial(app.Context(), types.ROLE_ACCOUNT_IDENTITY)
	app.S.Require().True(found, "identity contract not found")

	queryMsg := []byte(fmt.Sprintf(`{"query_user_by_address":{"address":"%s"}}`, address))
	resp, err := app.WasmxQueryRawNoCheck(sender, identityAddr, types.WasmxExecutionMessage{Data: queryMsg}, nil, nil)
	app.S.Require().NoError(err)
	decoded := app.QueryDecode(resp.Value)

	var out struct {
		UserID string `json:"user_id"`
	}
	err = json.Unmarshal(decoded, &out)
	app.S.Require().NoError(err)
	return out.UserID
}
