package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emersion/go-msgauth/dkim"
)

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	resp := SignDKIMResponse{Error: ""}
	return resp
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	resp := VerifyDKIMResponse{Error: ""}
	reader := strings.NewReader(req.EmailRaw)

	// Verify DKIM signatures
	options := &dkim.VerifyOptions{}
	verifications, err := dkim.VerifyWithOptions(reader, options)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	verificationsbz, err := json.Marshal(verifications)
	fmt.Printf("DKIM verifications: %s\n", string(verificationsbz))

	// Process verification results
	for _, v := range verifications {
		if v.Err == nil {
			fmt.Printf("DKIM signature verified successfully for domain: %s\n", v.Domain)
		} else {
			fmt.Printf("DKIM verification failed for domain: %s: %v\n", v.Domain, v.Err)
			// return v.Err
		}
		fmt.Println("* ", v.Domain, v.Expiration, v.Identifier, v.Time)
		fmt.Println("* ", v.HeaderKeys)
	}
	resp.IsValid = true
	return resp
}
