package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	sdktx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenerateTxCLI(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) (client.TxBuilder, sdktx.Factory, error) {
	txf, err := sdktx.NewFactoryCLI(clientCtx, flagSet)
	if err != nil {
		return nil, txf, err
	}

	return GenerateTxWithFactory(clientCtx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateTxWithFactory(clientCtx client.Context, txf sdktx.Factory, msgs ...sdk.Msg) (client.TxBuilder, sdktx.Factory, error) {
	// Validate all msgs before generating or broadcasting the tx.
	// We were calling ValidateBasic separately in each CLI handler before.
	// Right now, we're factorizing that call inside this function.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return nil, txf, err
		}
	}

	return GenerateUnsignedTx(clientCtx, txf, msgs...)
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func GenerateUnsignedTx(clientCtx client.Context, txf sdktx.Factory, msgs ...sdk.Msg) (client.TxBuilder, sdktx.Factory, error) {
	txf, err := txf.Prepare(clientCtx)
	if err != nil {
		return nil, txf, err
	}

	tx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, txf, err
	}

	if !clientCtx.SkipConfirm {
		encoder := clientCtx.TxConfig.TxJSONEncoder()
		if encoder == nil {
			return nil, txf, errors.New("failed to encode transaction: tx json encoder is nil")
		}

		txBytes, err := encoder(tx.GetTx())
		if err != nil {
			return nil, txf, fmt.Errorf("failed to encode transaction: %w", err)
		}

		if err := clientCtx.PrintRaw(json.RawMessage(txBytes)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, txBytes)
		}

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\ncanceled transaction\n", err)
			return nil, txf, err
		}
		if !ok {
			_, _ = fmt.Fprintln(os.Stderr, "canceled transaction")
			return nil, txf, err
		}
	}

	return tx, txf, nil
}

func SignTx(clientCtx client.Context, txf sdktx.Factory, tx client.TxBuilder) (client.TxBuilder, error) {
	if err := sdktx.Sign(clientCtx.CmdContext, txf, clientCtx.FromName, tx, true); err != nil {
		return nil, err
	}
	return tx, nil
}
