package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	runtime "github.com/loredanacirstea/wasmx-wazero"

	"wasmx/v1/app"
	cmd "wasmx/v1/cmdutils"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(runtime.WazeroVmMeta{})
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
