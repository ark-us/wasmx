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

	// Create DNS resolver for DKIM verification
	dnsResolver := NewDNSResolver()

	// Create custom DKIM verification options with our DNS resolver
	options := &dkim.VerifyOptions{
		LookupTXT: func(name string) ([]string, error) {
			return dnsResolver.LookupTXT(name)
		},
	}

	// Verify DKIM signatures
	verifications, err := dkim.VerifyWithOptions(reader, options)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	verificationsbz, err := json.Marshal(verifications)
	fmt.Printf("DKIM verifications: %s\n", string(verificationsbz))

	// Process verification results
	allValid := true
	for _, v := range verifications {
		if v.Err == nil {
			fmt.Printf("DKIM signature verified successfully for domain: %s\n", v.Domain)
		} else {
			fmt.Printf("DKIM verification failed for domain: %s: %v\n", v.Domain, v.Err)
			allValid = false
		}
		fmt.Println("* ", v.Domain, v.Expiration, v.Identifier, v.Time)
		fmt.Println("* ", v.HeaderKeys)
	}

	resp.IsValid = allValid
	return resp
}
