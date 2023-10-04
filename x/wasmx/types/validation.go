package types

import (
	"fmt"
	"math/big"
	"net/url"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func validateWasmCode(s []byte, maxSize int) error {
	if len(s) == 0 {
		return sdkerr.Wrap(ErrEmpty, "is required")
	}
	if len(s) > maxSize {
		return sdkerr.Wrapf(ErrLimit, "cannot be longer than %d bytes", maxSize)
	}
	return nil
}

func validateCode(s []byte, maxSize int) error {
	if len(s) == 0 {
		return sdkerr.Wrap(ErrEmpty, "is required")
	}
	if len(s) > maxSize {
		return sdkerr.Wrapf(ErrLimit, "cannot be longer than %d bytes", maxSize)
	}
	return nil
}

// ValidateLabel ensure label constraints
func ValidateLabel(label string) error {
	if label == "" {
		return sdkerr.Wrap(ErrEmpty, "is required")
	}
	if len(label) > MaxLabelSize {
		return ErrLimit.Wrapf("cannot be longer than %d characters", MaxLabelSize)
	}
	return nil
}

// ValidateSalt ensure salt constraints
func ValidateSalt(salt []byte) error {
	switch n := len(salt); {
	case n == 0:
		return sdkerr.Wrap(ErrEmpty, "is required")
	case n > MaxSaltSize:
		return ErrLimit.Wrapf("cannot be longer than %d characters", MaxSaltSize)
	}
	return nil
}

// ValidateVerificationInfo ensure source, builder and checksum constraints
func ValidateVerificationInfo(source, builder string, codeHash []byte) error {
	// if any set require others to be set
	if len(source) != 0 || len(builder) != 0 || codeHash != nil {
		if source == "" {
			return fmt.Errorf("source is required")
		}
		if _, err := url.ParseRequestURI(source); err != nil {
			return fmt.Errorf("source: %s", err)
		}
		if builder == "" {
			return fmt.Errorf("builder is required")
		}
		if codeHash == nil {
			return fmt.Errorf("code hash is required")
		}
		// code hash checksum match validation is done in the keeper, ungzipping consumes gas
	}
	return nil
}

func IsSystemAddress(addr sdk.AccAddress) bool {
	return big.NewInt(0).SetBytes(addr.Bytes()).Cmp(TRUSTED_ADDRESS_LIMIT) >= 0
}
