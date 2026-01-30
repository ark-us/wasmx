package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestKayrosVerifier() {
	appA := s.AppContext()
	verifierAddr, err := appA.App.WasmxKeeper.GetAddressOrRole(appA.Context(), types.ROLE_VERIFIER)
	suite.Require().NoError(err)

	sender := suite.GetRandomAccount()

	apiBaseUrl := ""
	apiUserKey := ""
	if suite.StartNodeEnv != nil {
		apiBaseUrl = suite.StartNodeEnv["kayros_base_url"]
		apiUserKey = suite.StartNodeEnv["kayros_user_key"]
	}

	if apiBaseUrl == "" {
		suite.T().Skip("SKIPPING ... set kayros_base_url and kayros_user_key for TestKayrosVerifier")
	}

	for _, proof := range KAYROS_TEST_PROOFS {
		dataType := proof.DataType
		data, err := base64.StdEncoding.DecodeString(proof.Data)
		suite.Require().NoError(err)

		verifyProofReq := map[string]any{
			"verify_proof": map[string]any{
				"data":         data,
				"data_type":    dataType,
				"hash_algo":    proof.HashAlgo,
				"api_base_url": apiBaseUrl,
				"api_user_key": apiUserKey,
			},
		}
		msg, err := json.Marshal(verifyProofReq)
		suite.Require().NoError(err)
		res := appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

		var verifyResp struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		suite.Require().NoError(json.Unmarshal(res, &verifyResp))
		suite.Require().True(verifyResp.Ok, verifyResp.Error)

		if proof.Trusted != nil {
			verifyProofReq = map[string]any{
				"verify_proof_with_inclusion": map[string]any{
					"data":              data,
					"data_type":         dataType,
					"hash_algo":         proof.HashAlgo,
					"trusted_root_hash": proof.Trusted.TopLevelHash,
					"trusted_level":     proof.Trusted.TopLevel,
					"trusted_position":  proof.Trusted.TopLevelPosition,
					"api_base_url":      apiBaseUrl,
					"api_user_key":      apiUserKey,
				},
			}
			msg, err = json.Marshal(verifyProofReq)
			suite.Require().NoError(err)
			res = appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

			verifyResp = struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{}
			suite.Require().NoError(json.Unmarshal(res, &verifyResp))
			suite.Require().True(verifyResp.Ok, verifyResp.Error)
		}

	}

	for _, proof := range KAYROS_TEST_PROOFS_HASH {
		dataType := proof.DataType
		dataHash, err := hex.DecodeString(proof.DataHash)
		suite.Require().NoError(err)

		verifyProofHashReq := map[string]any{
			"verify_proof_hash": map[string]any{
				"data_type":    dataType,
				"data_hash":    dataHash,
				"api_base_url": apiBaseUrl,
				"api_user_key": apiUserKey,
			},
		}
		msg, err := json.Marshal(verifyProofHashReq)
		suite.Require().NoError(err)
		res := appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

		var verifyResp struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		suite.Require().NoError(json.Unmarshal(res, &verifyResp))
		suite.Require().True(verifyResp.Ok, verifyResp.Error)

		if proof.Trusted != nil {
			verifyProofHashReq = map[string]any{
				"verify_proof_hash_with_inclusion": map[string]any{
					"data_type":         dataType,
					"data_hash":         dataHash,
					"trusted_root_hash": proof.Trusted.TopLevelHash,
					"trusted_level":     proof.Trusted.TopLevel,
					"trusted_position":  proof.Trusted.TopLevelPosition,
					"api_base_url":      apiBaseUrl,
					"api_user_key":      apiUserKey,
				},
			}
			msg, err = json.Marshal(verifyProofHashReq)
			suite.Require().NoError(err)
			res = appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

			verifyResp = struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{}
			suite.Require().NoError(json.Unmarshal(res, &verifyResp))
			suite.Require().True(verifyResp.Ok, verifyResp.Error)
		}
	}
}

// var KAYROS_TEST_BY_HASH = []string{
// 	"c883af9b3ffc8261f6385e33dcd0bc825de1480dad4b5e9c9425c45c66655065",
// 	"b605d60a6a531c57eba104a0e28a8be111e95d7a88a257066274de688eec1216",
// }

// var KAYROS_TEST_BY_DATA = [][]string{
// 	{
// 		"7761736d785f6d7974686f735f373030312d315f313736373039353338333930",
// 		"374f1b87a48721c2a3f4d0777a6848e23edb13d0209a69f65330ee7ae84d5e9c",
// 	},
// }

// // var KAYROS_TOPLEVEL_HASH = "be475ef98de5dcba176730b044a32239dbe5e8bb4f2172d5d8c555907eb445e0"
// // var KAYROS_TOPLEVEL_LEVEL = 1
// // var KAYROS_TOPLEVEL_POSITION = 10

// var KAYROS_TOPLEVEL_HASH = ""
// var KAYROS_TOPLEVEL_LEVEL = -1
// var KAYROS_TOPLEVEL_POSITION = -1

// var KAYROS_TEST_PROOFS = []VerifyProofRequest{
// 	{
// 		Data:     "CsoBCscBCiMvbXl0aG9zLndhc214LnYxLk1zZ0V4ZWN1dGVDb250cmFjdBKfAQotbXl0aG9zMTdzY2NsMDk1MzJod2g5eTRzeTlxNXNsd216ajdrZnBhNDJ5czY2Ei1teXRob3MxN2syMGVxanJ0czNlaDZteGtxcXNraHp2NGtqYWUwNTNrYzhkbGQaP3siZGF0YSI6ImV5SnpaWFFpT25zaWEyVjVJam9pYUdWc2JHOGlMQ0oyWVd4MVpTSTZJbk5oYlcxNUluMTkifRJvClAKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECqkwoU7+JW3gfOgr6as2ZJZD6Vo2Ob7H8sgOIIUOIrasSBAoCCAEYAxIbChMKBGFteXQSCzEwMDAwMDAwMDAwEICU69wDGkAMfcSxXine4wbRzIm1edq7DzGF089DRVHmeVI159CfqlcIS8adRZVD5ndJ4fFpUfkOjsg5DhLmQvyP+QGTynI0",
// 		DataType: "7761736d785f74785f6d7974686f735f373030312d315f313736383834393037",
// 		HashAlgo: "sha256",
// 		// txhash
// 		// 9bfbd26d8e9cc4f69eca1ffadc2e5db2891e0781efad9ec969e0cf980208eb10
// 		// kayros hash
// 		// 5dbff8d307b6576610edea33ef2df6e53c93d9849b76f98b2665f59305d8dd14
// 	},
// }

var KAYROS_TEST_BY_HASH = []string{
	"a333543860b67533eeaaa8b24f34eff0be0ef585a05ddc73c14fd50f94645ec6",
	"c27b15bd293a31374617f094c8c48283d7821618c9c5e11bd263b07940a2b234",
}

var KAYROS_TEST_PROOFS = []VerifyProofRequest{
	{
		Data:     "CsoBCscBCiMvbXl0aG9zLndhc214LnYxLk1zZ0V4ZWN1dGVDb250cmFjdBKfAQotbXl0aG9zMXM2ejZoeDdrdWdwMnZqcnIweW51ZGdwejZlOWVod2FqN2hxcHQ4Ei1teXRob3MxOXY5a2N3dTB2bXpjN2NqY2g2aGFxZXRzNXNxcGdqZjdkMG1hcmoaP3siZGF0YSI6ImV5SnpaWFFpT25zaWEyVjVJam9pYUdWc2JHOGlMQ0oyWVd4MVpTSTZJbk5oYlcxNUluMTkifRJvClAKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECTBbqvLhqttZhiEZeiyjjPe9s707sc1ZWJfN3dBrmEbMSBAoCCAEYAxIbChMKBGFteXQSCzEwMDAwMDAwMDAwEICU69wDGkDGpjSJ9mwCScxEcFCEmoRq7VUM2jJT5KMp4vWdTyjrMB22c9RRCJNW5X8tw0wW71w3MBNPZwtXcnOIu38LCq1E",
		DataType: "wasmx_t_mythos_7001-1_1769607041",
		HashAlgo: "sha256",
		// txhash
		// bb4c35eb7c081e769f1fc7c03f9e8b7b5d05f5bc8dd894ce05f7a10ed9808775
	},
}

var KAYROS_TEST_PROOFS_HASH = []VerifyProofHashRequest{
	{
		DataHash: "aab0c4c694bf8e18000000000000000000000000000000000000000000000000",
		DataType: "benchmark_s32",
		Trusted: &ProofTrusted{
			TopLevelHash:     "79687b6f0347b689ac82ac79d64eb188a7580c3cdabf04b2433bbc9e135c5f7a",
			TopLevel:         2,
			TopLevelPosition: 1,
		},
	},
	{
		DataHash: "2ce7635e8cbf8e18000000000000000000000000000000000000000000000000",
		DataType: "benchmark_s32",
		Trusted: &ProofTrusted{
			TopLevelHash:     "5825048f9e70b7f3f4d3679073815f0260f890ce1f786aeb48d97dbcd46106b8",
			TopLevel:         2,
			TopLevelPosition: 0,
		},
	},
}
