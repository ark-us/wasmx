package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	runtime "github.com/loredanacirstea/wasmx-wazero"

	"github.com/loredanacirstea/wasmx/app"
	cmd "github.com/loredanacirstea/wasmx/cmdutils"

	vmpostgresql "github.com/loredanacirstea/wasmx-vmpostgresql"
	"github.com/loredanacirstea/wasmx/x/network/vmcrosschain"
	"github.com/loredanacirstea/wasmx/x/network/vmmc"
	"github.com/loredanacirstea/wasmx/x/network/vmp2p"
	"github.com/loredanacirstea/wasmx/x/vmhttpclient"
	"github.com/loredanacirstea/wasmx/x/vmhttpserver"
	"github.com/loredanacirstea/wasmx/x/vmkv"
	"github.com/loredanacirstea/wasmx/x/vmoauth2client"
	"github.com/loredanacirstea/wasmx/x/vmsql"
)

func init() {
	// enabled VM extensions for contracts
	vmp2p.Setup()
	vmmc.Setup()

	// experimental WIP, do not enable in production:
	vmcrosschain.Setup()
	vmsql.Setup()
	vmpostgresql.Setup()
	vmkv.Setup()
	vmimap.Setup()
	vmsmtp.Setup()
	vmhttpclient.Setup()
	vmhttpserver.Setup()
	vmoauth2client.Setup()
}

func main() {
	rootCmd, _ := cmd.NewRootCmd(runtime.NewWazeroVmMeta(), app.DefaultNodeHome, nil)
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
