package runtime

import (
	"errors"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type Gas = uint64
type GasMeter struct {
	gasLimit uint64
	gasUsed  uint64
	gasMeter memc.GasMeter
}

func NewGasMeter(gasLimit uint64, gasUsed uint64, gasMeter memc.GasMeter) *GasMeter {
	return &GasMeter{
		gasLimit: gasLimit,
		gasUsed:  gasUsed,
		gasMeter: gasMeter,
	}
}

// ErrOutOfGas is returned when execution exceeds gas limit
var ErrOutOfGas = errors.New("out of gas")

// ConsumeGas attempts to consume the given amount of gas and returns error if not enough gas remains
func (g *GasMeter) ConsumeGas(amount uint64, descriptor string) {
	g.gasUsed += amount
	if g.gasUsed > g.gasLimit {
		// we expect this to error with out of gas
		if g.gasMeter != nil {
			g.gasMeter.ConsumeGas(g.gasUsed, descriptor)
		} else {
			panic(ErrOutOfGas)
		}
	}
}

func (g *GasMeter) GasConsumed() Gas {
	return g.gasUsed
}
func (g *GasMeter) GasLimit() Gas {
	return g.gasLimit
}
func (g *GasMeter) GasRemaining() Gas {
	return g.gasLimit - g.gasUsed
}
