package vm

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"time"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// nondeterministic
func asDateNow_NonD(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO move
	returns := make([]interface{}, 1)
	returns[0] = float64(time.Now().UTC().UnixMilli())
	return returns, nil
}

// nondeterministic
func asSeed_NonD(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO move
	returns := make([]interface{}, 1)
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return nil, err
	}
	returns[0] = float64(binary.LittleEndian.Uint64(b[:]))
	return returns, nil
}
