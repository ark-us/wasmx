package lib

import verifier "github.com/loredanacirstea/wasmx-kayros-verifier/lib"

type KayrosClient = verifier.KayrosClient
type KayrosConfig = verifier.KayrosConfig
type KayrosRecord = verifier.KayrosRecord
type KayrosApiResponse = verifier.KayrosApiResponse
type KayrosRecordResponse = verifier.KayrosRecordResponse
type KayrosRecordsData = verifier.KayrosRecordsData
type KayrosRecordsResponse = verifier.KayrosRecordsResponse
type KayrosRegistrationRequest = verifier.KayrosRegistrationRequest
type KayrosRegistrationResponse = verifier.KayrosRegistrationResponse
type KayrosRegistrationResponseWrap = verifier.KayrosRegistrationResponseWrap

func NewKayrosClient(config KayrosConfig) *KayrosClient {
	return verifier.NewKayrosClient(config)
}
