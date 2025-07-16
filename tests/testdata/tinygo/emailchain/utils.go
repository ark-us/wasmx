package main

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

func ToPrivateKey(keyType string, pk []byte) crypto.Signer {
	var signer crypto.Signer
	var err error
	if keyType == "rsa" {
		// we expect privatekey in PEM format
		block, _ := pem.Decode(pk)
		var err error
		signer, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}
	} else {
		signer, err = loadPrivateKey(pk)
		if err != nil {
			panic(err)
		}
	}
	return signer
}

// GenerateMessageID generates a unique RFC 5322-compliant Message-ID.
// Example: e4cfd38a7bce4fda9a2a4cc21f24a3b2@yourdomain.com
func GenerateMessageID(domain string, date time.Time) (string, error) {
	// 16 bytes of randomness
	buf := make([]byte, 16)
	// TODO we dont have randomness yet
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	timestamp := date.UnixNano()
	localPart := fmt.Sprintf("22%x.%d", buf, timestamp)

	return fmt.Sprintf("%s@%s", localPart, domain), nil
}

func ParseEmailDate(value string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,  // "02 Jan 06 15:04 -0700"
		"02 Jan 2006 15:04:05 -0700",
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %q", value)
}
