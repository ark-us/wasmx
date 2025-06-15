package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dkim "github.com/redsift/dkim"

	dkimS "github.com/emersion/go-msgauth/dkim"
)

const DKIM_HEADER = "DKIM-Signature"

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	fmt.Println("--SignDKIM--")
	resp := SignDKIMResponse{Error: ""}

	r := strings.NewReader(req.EmailRaw)
	var b bytes.Buffer

	now := func() time.Time {
		return req.Timestamp
	}

	options := req.Options.toLib()
	fmt.Println("--SignSync options--")
	err := dkimS.SignSync(&b, r, options, now)
	fmt.Println("--SignSync--")
	fmt.Println("--SignSync err--", err)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.SignedEmail = b.String()
	return resp
}

func SignARC(req *SignARCRequest) SignARCResponse {
	resp := SignARCResponse{Error: ""}

	r := strings.NewReader(req.EmailRaw)
	var b bytes.Buffer

	now := func() time.Time {
		return req.Timestamp
	}

	dnsResolver := NewDNSResolver()
	lookupTxt := func(name string) ([]string, error) {
		return dnsResolver.LookupTXT(name)
	}

	options := req.Options.toLib()
	options.LookupTXT = lookupTxt
	fmt.Println("--SignSync options--")
	err := dkimS.SignARCSync(&b, r, options, now)
	fmt.Println("--SignSync--")
	fmt.Println("--SignSync err--", err)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.SignedEmail = b.String()
	return resp
}

func VerifyARC(req *VerifyDKIMRequest) VerifyARCResponse {
	fmt.Println("--VerifyARC--")
	resp := VerifyARCResponse{Error: ""}
	msg, err := dkim.ParseMessage(req.EmailRaw)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	dnsResolver := NewDNSResolver()
	lookupTxt := func(name string) ([]string, error) {
		return dnsResolver.LookupTXT(name)
	}

	res, err := dkim.VerifyArc(lookupTxt, req.PublicKey, msg)
	fmt.Println("--VerifyArc err, res--", err, res)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	resbz, err := json.Marshal(res)
	fmt.Println("--dkim.VerifyArc resbz--", err, string(resbz))

	resp.Response = &ArcResult{}
	fmt.Println("--VerifyArc FromLib pre--")
	resp.Response.FromLib(res)
	fmt.Println("--VerifyArc FromLib--")

	resbz, err = json.Marshal(resp.Response)
	fmt.Println("--dkim.VerifyArc resp.Response--", err, string(resbz))
	return resp
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	fmt.Println("--VerifyDKIM--" + req.EmailRaw)
	resp := VerifyDKIMResponse{Error: ""}

	msg, err := dkim.ParseMessage(req.EmailRaw)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	dnsResolver := NewDNSResolver()
	lookupTxt := func(name string) ([]string, error) {
		return dnsResolver.LookupTXT(name)
	}

	// if we want to exclude domains
	// InvalidSigningEntityOption("com", "org", "net"),

	// if we want to fail if expiration date failed
	// SignatureTimingOption(5*time.Minute)
	res, err := dkim.Verify(DKIM_HEADER, msg, lookupTxt, req.PublicKey)
	fmt.Println("--dkim.Verify--", err, res)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	resbz, err := json.Marshal(res)
	fmt.Println("--dkim.Verify resbz--", err, string(resbz))

	resp.Response = ResultArrFromLib(res)

	fmt.Println("--dkim.Verify post ResultArrFromLib--")

	resbz, err = json.Marshal(resp.Response)
	fmt.Println("--dkim.Verify resp.Response--", err, string(resbz))

	return resp
}
