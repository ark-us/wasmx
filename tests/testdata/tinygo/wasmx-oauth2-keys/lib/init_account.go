package lib

import (
	"encoding/hex"
	"fmt"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InitAccount initializes a new account by sending native coins to it
// This can only be called by contracts with internal roles
func InitAccount(msg *MsgInitAccount) []byte {
	fmt.Println("InitAccount called", []string{"address", msg.Address})

	// Check OnlyInternal - verify caller is a contract with a role
	// TODO: Implement proper access control check
	// For now, we rely on the fact that only contracts can call this via CallSimple
	wasmx.OnlyInternal(MODULE_NAME, "InitAccount")

	// Load funder private key
	privKeyHex := LoadFunderPrivateKey()
	if privKeyHex == "" {
		LoggerError("InitAccount: funder private key not configured", nil)
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "funder private key not configured",
		})
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		LoggerError("InitAccount: failed to decode private key", []string{"error", err.Error()})
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "failed to decode funder private key",
		})
	}

	// Get coin to send
	coin := LoadInitAccountAmount()
	if coin == nil {
		LoggerError("InitAccount: empty init account amount", []string{})
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "empty init account amount",
		})
	}

	// Load gas price
	gasPrice := LoadGasPrice()
	if gasPrice == nil {
		gasPrice = &wasmx.Coin{}
	}

	// Prepare transaction: bank send (empty data, just native coin transfer)
	// The fromAddress will be derived from the privateKey by the host
	txBytes, err := wasmxcore.PrepareTx(
		msg.Address,
		[]byte{}, // empty data for native coin transfer
		[]wasmx.Coin{*coin},
		nil,
		30000000,  // gas
		*gasPrice, // gas price from storage
		privKeyBytes,
	)
	if err != nil {
		LoggerError("InitAccount: failed to prepare transaction", []string{"error", err.Error()})
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "failed to prepare transaction: " + err.Error(),
		})
	}

	fmt.Println("InitAccount: transaction prepared", []string{"tx_bytes_len", fmt.Sprintf("%d", len(txBytes))})

	// Broadcast transaction
	broadcastResp, err := wasmxcore.BroadcastTxAsync(txBytes)
	fmt.Println("InitAccount: tx broadcasted", err, broadcastResp)
	if err != nil {
		LoggerError("InitAccount: failed to broadcast transaction", []string{"error", err.Error()})
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "failed to broadcast transaction: " + err.Error(),
		})
	}

	if broadcastResp.Error != "" {
		LoggerError("InitAccount: transaction broadcast error", []string{"error", broadcastResp.Error})
		return MarshalJSON(MsgInitAccountResponse{
			Success: false,
			Error:   "transaction broadcast error: " + broadcastResp.Error,
		})
	}

	// Get transaction hash
	txHash := ""
	if broadcastResp.Response != nil {
		txHash = broadcastResp.Response.Hash.String()
	}

	fmt.Println("InitAccount: transaction broadcasted", []string{"tx_hash", txHash, "to_address", msg.Address})

	return MarshalJSON(MsgInitAccountResponse{
		Success: true,
		TxHash:  txHash,
	})
}
