package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestKayrosVerifier() {
	type VerifyProofWithInclusionResp struct {
		Ok               bool   `json:"ok"`
		Error            string `json:"error"`
		Pending          bool   `json:"pending"`
		MaxLevel         int    `json:"max_level"`
		MaxLevelPosition int64  `json:"max_level_position"`
		MaxLevelHash     string `json:"max_level_hash"`
	}

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

	var verifyRespInc VerifyProofWithInclusionResp

	for _, proof := range KAYROS_TEST_PROOFS {
		dataType := proof.DataType
		if proof.DataHash == "" && proof.Data == "" {
			suite.T().Fatal("missing proof data and data hash")
		}

		if proof.DataHash != "" {
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

			if proof.VerifyInclusion {
				trustedRootHash := ""
				trustedLevel := -1
				trustedPosition := -1
				if proof.Trusted != nil {
					trustedRootHash = proof.Trusted.TopLevelHash
					trustedLevel = proof.Trusted.TopLevel
					trustedPosition = proof.Trusted.TopLevelPosition
				}
				verifyProofHashReq = map[string]any{
					"verify_proof_hash_with_inclusion": map[string]any{
						"data_type":           dataType,
						"data_hash":           dataHash,
						"trusted_root_hash":   trustedRootHash,
						"trusted_level":       trustedLevel,
						"trusted_position":    trustedPosition,
						"verify_db_existence": proof.VerifyDbExistence,
						"api_base_url":        apiBaseUrl,
						"api_user_key":        apiUserKey,
					},
				}
				msg, err = json.Marshal(verifyProofHashReq)
				suite.Require().NoError(err)
				res = appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

				verifyRespInc = VerifyProofWithInclusionResp{}
				suite.Require().NoError(json.Unmarshal(res, &verifyRespInc))
				suite.Require().True(verifyRespInc.Ok, verifyRespInc.Error, proof.DataHash, proof.DataType)
				fmt.Println("--verifyRespInc--", string(res))
			}
			continue
		}

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

		if proof.VerifyInclusion {
			trustedRootHash := ""
			trustedLevel := -1
			trustedPosition := -1
			if proof.Trusted != nil {
				trustedRootHash = proof.Trusted.TopLevelHash
				trustedLevel = proof.Trusted.TopLevel
				trustedPosition = proof.Trusted.TopLevelPosition
			}
			verifyProofReq = map[string]any{
				"verify_proof_with_inclusion": map[string]any{
					"data":              data,
					"data_type":         dataType,
					"hash_algo":         proof.HashAlgo,
					"trusted_root_hash": trustedRootHash,
					"trusted_level":     trustedLevel,
					"trusted_position":  trustedPosition,
					"api_base_url":      apiBaseUrl,
					"api_user_key":      apiUserKey,
				},
			}
			msg, err = json.Marshal(verifyProofReq)
			suite.Require().NoError(err)
			res = appA.WasmxQueryRaw(sender, verifierAddr, types.WasmxExecutionMessage{Data: msg}, nil, nil)

			verifyRespInc = VerifyProofWithInclusionResp{}
			suite.Require().NoError(json.Unmarshal(res, &verifyRespInc))
			suite.Require().True(verifyRespInc.Ok, verifyRespInc.Error)
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

var KAYROS_TEST_PROOFS = []VerifyProofTest{
	// {
	// 	KayrosHash:      "0bd650857aedc4b1edbcb9319be3f9b7676589062795439b10e0761140478f2a",
	// 	DataType:        "benchmark_s32",
	// 	VerifyInclusion: false,
	// },
	// {
	// 	KayrosHash:      "336b2d19406f947646b294e108a5228c6da8bbbc7474b8eafe8e48409a9d3cd8",
	// 	DataType:        "benchmark_s32",
	// 	VerifyInclusion: false,
	// },
	{
		DataHash: "bbce1c1109d29018000000000000000000000000000000000000000000000000",
		DataType: "benchmark_s32",
		Trusted: &ProofTrusted{
			TopLevelHash:     "a930abc733ab8b16355c34682d5e1998da559655800d6c3c192808ef14bbab61",
			TopLevel:         2,
			TopLevelPosition: 31,
		},
		// 0bd650857aedc4b1edbcb9319be3f9b7676589062795439b10e0761140478f2a
		VerifyInclusion: true,
	},
	{
		DataHash:          "88188c338fd09018000000000000000000000000000000000000000000000000",
		DataType:          "benchmark_s32",
		VerifyDbExistence: true,
		Trusted: &ProofTrusted{
			TopLevelHash:     "150743eeee4a9410462a3f1a24afefe43b48e8ec3617c9f54e4d5bd695428e06",
			TopLevel:         2,
			TopLevelPosition: 0,
		},
		// hash: 336b2d19406f947646b294e108a5228c6da8bbbc7474b8eafe8e48409a9d3cd8
		VerifyInclusion: true,
	},
	{
		DataHash:          "88188c338fd09018000000000000000000000000000000000000000000000000",
		DataType:          "benchmark_s32",
		VerifyDbExistence: true,
		// hash: 336b2d19406f947646b294e108a5228c6da8bbbc7474b8eafe8e48409a9d3cd8
		VerifyInclusion: true,
	},
	{
		DataHash:          "5f43421e09d29018000000000000000000000000000000000000000000000000",
		DataType:          "benchmark_s32",
		VerifyDbExistence: true,
		// hash: 5840eed227115b8712ac1d1e3434c92c6161735853c6af7e70ea206f367ea089
		VerifyInclusion: true,
	},
}
