package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	runtime "github.com/loredanacirstea/wasmx-wasmedge"

	"github.com/loredanacirstea/wasmx/app"
	cmd "github.com/loredanacirstea/wasmx/cmdutils"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(runtime.WasmEdgeVmMeta{})
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
