package lib

import verifier "github.com/loredanacirstea/wasmx-kayros-verifier/lib"

type KayrosClient = verifier.KayrosClient
type KayrosConfig = verifier.KayrosConfig
type KayrosRecord = verifier.KayrosRecord
type KayrosRecordsResponse = verifier.KayrosRecordsResponse
type KayrosRegistrationRequest = verifier.KayrosRegistrationRequest
type KayrosRegistrationResponse = verifier.KayrosRegistrationResponse

func NewKayrosClient(config KayrosConfig) *KayrosClient {
	return verifier.NewKayrosClient(config)
}
