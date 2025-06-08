package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emersion/go-msgauth/dkim"

	dkim2 "github.com/redsift/dkim"
)

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	resp := SignDKIMResponse{Error: ""}
	return resp
}

func VerifyARC(req *VerifyDKIMRequest) VerifyARCResponse {
	resp := VerifyARCResponse{Error: ""}
	msg, err := dkim2.ParseMessage(req.EmailRaw)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	dnsResolver := NewDNSResolver()
	lookupTxt := func(name string) ([]string, error) {
		return dnsResolver.LookupTXT(name)
	}

	res, err := dkim2.VerifyArc(lookupTxt, msg)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Response = res
	return resp
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	fmt.Println("--VerifyDKIM--" + req.EmailRaw)
	resp := VerifyDKIMResponse{Error: ""}

	// Create custom DKIM verifier with DNS-over-HTTPS
	// verifier := NewCustomDKIMVerifier()

	// Verify DKIM signatures
	// verifications, err := verifier.VerifyDKIMSignatures(req.EmailRaw)

	// Create DNS resolver for DKIM verification
	dnsResolver := NewDNSResolver()

	reader := strings.NewReader(req.EmailRaw)
	// Create custom DKIM verification options with our DNS resolver
	options := &dkim.VerifyOptions{
		LookupTXT: func(name string) ([]string, error) {
			return dnsResolver.LookupTXT(name)
		},
	}

	verifications, err := dkim.VerifyWithOptions(reader, options)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Verifications = verifications

	verificationsbz, err := json.Marshal(verifications)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to marshal verifications: %v", err)
		return resp
	}
	fmt.Printf("DKIM verifications: %s\n", string(verificationsbz))

	allValid := true
	for _, v := range verifications {
		if v.Err != nil {
			allValid = false
			break
		}
	}
	resp.IsValid = allValid
	return resp
}
