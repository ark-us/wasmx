package main

import (
	"encoding/json"
	"fmt"
)

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	resp := SignDKIMResponse{Error: ""}
	return resp
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	resp := VerifyDKIMResponse{Error: ""}

	// Create custom DKIM verifier with DNS-over-HTTPS
	verifier := NewCustomDKIMVerifier()

	// Verify DKIM signatures
	verifications, err := verifier.VerifyDKIMSignatures(req.EmailRaw)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	verificationsbz, err := json.Marshal(verifications)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to marshal verifications: %v", err)
		return resp
	}
	fmt.Printf("DKIM verifications: %s\n", string(verificationsbz))

	// Process verification results
	allValid := true
	for _, v := range verifications {
		if v.Valid {
			fmt.Printf("DKIM signature verified successfully for domain: %s\n", v.Domain)
		} else {
			fmt.Printf("DKIM verification failed for domain: %s: %s\n", v.Domain, v.Error)
			allValid = false
		}
		fmt.Printf("* Domain: %s, Selector: %s, Algorithm: %s\n", v.Domain, v.Selector, v.Algorithm)
		fmt.Printf("* Header Keys: %v\n", v.HeaderKeys)
	}

	resp.IsValid = allValid
	return resp
}
