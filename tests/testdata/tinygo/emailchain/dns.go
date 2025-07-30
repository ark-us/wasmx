package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dnsMox "github.com/loredanacirstea/mailverif/dns"
	httpclient "github.com/loredanacirstea/wasmx-env-httpclient"
)

type MX struct {
	Host string `json:"host"` // e.g. "mail.example.com"
	Pref uint16 `json:"pref"` // e.g. 10
}

// DNSResolver implements DNS resolution over HTTPS using httpclient
type DNSResolver struct {
	DoHURL string // DNS-over-HTTPS URL (e.g., "https://dns.google/resolve")
}

// NewDNSResolver creates a new DNS resolver using DNS-over-HTTPS
func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		DoHURL: "https://dns.google/resolve",
	}
}

// DNSResponse represents a DNS response from DoH provider
type DNSResponse struct {
	Status int `json:"Status"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		Data string `json:"data"`
	} `json:"Answer"`
}

// LookupTXT performs a TXT record lookup using DNS-over-HTTPS
func (r *DNSResolver) LookupTXT(name string) ([]string, dnsMox.Result, error) {
	// return nil, dnsMox.Result{}, fmt.Errorf("just fail")
	// Construct DoH query URL
	url := fmt.Sprintf("%s?name=%s&type=TXT", r.DoHURL, name)

	// Create HTTP request
	req := &httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "GET",
			Url:    url,
			Header: http.Header{
				"Accept": []string{"application/dns-json"},
			},
			Data: nil,
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize: 1024 * 1024, // 1MB max response
		},
	}

	// Make HTTP request via httpclient
	resp := httpclient.Request(req)
	if resp.Error != "" {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query failed: %s", resp.Error)
	}

	if resp.Data.StatusCode != 200 {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query returned status %d", resp.Data.StatusCode)
	}

	// Parse DNS response
	var dnsResp DNSResponse
	if err := json.Unmarshal(resp.Data.Data, &dnsResp); err != nil {
		return nil, dnsMox.Result{}, fmt.Errorf("failed to parse DNS response: %v", err)
	}

	if dnsResp.Status != 0 {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query failed with status %d", dnsResp.Status)
	}

	// Extract TXT records
	var txtRecords []string
	for _, answer := range dnsResp.Answer {
		if answer.Type == 16 { // TXT record type
			// Remove quotes from TXT record data
			txtData := strings.Trim(answer.Data, "\"")
			txtRecords = append(txtRecords, txtData)
		}
	}

	// TODO do DNSSEC check - Authentic
	return txtRecords, dnsMox.Result{Authentic: true}, nil
}

// LookupMX performs an MX record lookup using DNS-over-HTTPS.
func (r *DNSResolver) LookupMX(domain string) ([]MX, dnsMox.Result, error) {
	// Construct the DoH query URL
	url := fmt.Sprintf("%s?name=%s&type=MX", r.DoHURL, domain)

	// Prepare the HTTP request
	req := &httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "GET",
			Url:    url,
			Header: http.Header{
				"Accept": []string{"application/dns-json"},
			},
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize: 1024 * 1024, // 1 MB
		},
	}

	// Execute the request
	resp := httpclient.Request(req)
	if resp.Error != "" {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query failed: %s", resp.Error)
	}
	if resp.Data.StatusCode != 200 {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query returned status %d", resp.Data.StatusCode)
	}

	// Decode the JSON DoH response
	var dnsResp DNSResponse
	if err := json.Unmarshal(resp.Data.Data, &dnsResp); err != nil {
		return nil, dnsMox.Result{}, fmt.Errorf("failed to parse DNS response: %v", err)
	}
	if dnsResp.Status != 0 {
		return nil, dnsMox.Result{}, fmt.Errorf("DNS query failed with status %d", dnsResp.Status)
	}

	// Extract MX answers
	var mxRecords []MX
	for _, answer := range dnsResp.Answer {
		if answer.Type != 15 { // 15 = MX
			continue
		}
		// MX data comes as: "10 mail.example.com."
		parts := strings.Fields(answer.Data)
		if len(parts) != 2 {
			continue
		}
		pref, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		host := strings.TrimSuffix(parts[1], ".") // strip trailing dot
		mxRecords = append(mxRecords, MX{
			Host: host,
			Pref: uint16(pref),
		})
	}

	return mxRecords, dnsMox.Result{Authentic: true}, nil
}
