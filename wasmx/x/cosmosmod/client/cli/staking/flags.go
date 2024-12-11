package staking

import (
	flag "github.com/spf13/pflag"

	cli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// common flagsets to add to various functions
var (
	fsShares       = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator    = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsShares.String(cli.FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(cli.FlagSharesFraction, "", "Fraction of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsValidator.String(cli.FlagAddressValidator, "", "The address of the validator")
	fsRedelegation.String(cli.FlagAddressValidatorSrc, "", "The address of the source validator")
	fsRedelegation.String(cli.FlagAddressValidatorDst, "", "The address of the destination validator")
}

func flagSetDescriptionEdit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(cli.FlagEditMoniker, types.DoNotModifyDesc, "The validator's name")
	fs.String(cli.FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(cli.FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fs.String(cli.FlagSecurityContact, types.DoNotModifyDesc, "The validator's (optional) security contact email")
	fs.String(cli.FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")

	return fs
}

func flagSetCommissionUpdate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(cli.FlagCommissionRate, "", "The new commission rate percentage")

	return fs
}
