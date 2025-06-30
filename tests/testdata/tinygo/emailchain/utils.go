package main

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
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
