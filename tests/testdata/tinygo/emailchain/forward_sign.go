package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ed25519"

	dkimS "github.com/emersion/go-msgauth/dkim"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
)

const crlf = "\r\n"

var randReader io.Reader = rand.Reader

// Signer generates a DKIM signature.
//
// The whole message header and body must be written to the Signer. Close should
// always be called (either after the whole message has been written, or after
// an error occurred and the signer won't be used anymore). Close may return an
// error in case signing fails.
//
// After a successful Close, Signature can be called to retrieve the
// DKIM-Signature header field that the caller should prepend to the message.
type SignerForward struct {
	headerName string
	sigParams  map[string]string // only valid after done received nil
}

func NewSignerForward(headerName string, options *dkimS.SignOptions, email vmimap.Email, params map[string]string) (*SignerForward, error) {
	if options == nil {
		return nil, fmt.Errorf("signer: no options specified")
	}
	if options.Signer == nil {
		return nil, fmt.Errorf("signer: no signer specified")
	}

	headerCan := options.HeaderCanonicalization
	if headerCan == "" {
		headerCan = dkimS.CanonicalizationSimple
	}
	can_ := dkimS.GetCanonicalizer(headerCan)
	if can_ == nil {
		return nil, fmt.Errorf("signer: unknown header canonicalization %q", headerCan)
	}

	bodyCan := options.BodyCanonicalization
	if bodyCan == "" {
		bodyCan = dkimS.CanonicalizationSimple
	}
	can_ = dkimS.GetCanonicalizer(bodyCan)
	if can_ == nil {
		return nil, fmt.Errorf("signer: unknown body canonicalization %q", bodyCan)
	}

	var keyAlgo string
	switch options.Signer.Public().(type) {
	case *rsa.PublicKey:
		keyAlgo = "rsa"
	case ed25519.PublicKey:
		keyAlgo = "ed25519"
	default:
		return nil, fmt.Errorf("signer: unsupported key algorithm %T", options.Signer.Public())
	}

	hash := options.Hash
	var hashAlgo string
	switch options.Hash {
	case 0: // sha256 is the default
		hash = crypto.SHA256
		fallthrough
	case crypto.SHA256:
		hashAlgo = "sha256"
	case crypto.SHA1:
		return nil, fmt.Errorf("signer: hash algorithm too weak: sha1")
	default:
		return nil, fmt.Errorf("signer: unsupported hash algorithm")
	}

	if options.HeaderKeys != nil {
		ok := false
		for _, k := range options.HeaderKeys {
			if strings.EqualFold(k, "From") {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("signer: the From header field must be signed")
		}
	}

	// br := bufio.NewReader(bytes.NewReader(msg))
	// h, err := readHeader(br)
	// if err != nil {
	// 	return nil, err
	// }

	hasher := hash.New()
	// can := dkimS.GetCanonicalizer(bodyCan).CanonicalizeBody(hasher)
	// if _, err := io.Copy(can, br); err != nil {
	// 	return nil, err
	// }
	// if err := can.Close(); err != nil {
	// 	return nil, err
	// }
	// bodyHashed := hasher.Sum(nil)

	params["a"] = keyAlgo + "-" + hashAlgo
	params["c"] = string(headerCan) + "/" + string(bodyCan)

	// params := map[string]string{
	// 	"bh": base64.StdEncoding.EncodeToString(bodyHashed),
	// 	"d":  options.Domain,
	// 	"s": options.Selector,
	// 	"t": formatTime(now()),
	// }

	var headerKeys []string
	if options.HeaderKeys != nil {
		headerKeys = options.HeaderKeys
	} else {
		for _, h := range email.Headers {
			headerKeys = append(headerKeys, h.Key)
		}
	}
	params["h"] = formatTagList(headerKeys)

	if options.Identifier != "" {
		params["i"] = options.Identifier
	}
	if options.QueryMethods != nil {
		methods := make([]string, len(options.QueryMethods))
		for i, method := range options.QueryMethods {
			methods[i] = string(method)
		}
		params["q"] = formatTagList(methods)
	}
	if !options.Expiration.IsZero() {
		params["x"] = formatTime(options.Expiration)
	}

	hasher.Reset()
	for _, k := range headerKeys {
		h := email.Headers.Get(k)
		if h.Key == "" {
			// The Signer MAY include more instances of a header field name
			// in "h=" than there are actual corresponding header fields so
			// that the signature will not verify if additional header
			// fields of that name are added.
			continue
		}
		h.Key = dkimS.GetCanonicalizer(headerCan).CanonicalizeHeader(h.Key)
		if _, err := io.WriteString(hasher, h.Key); err != nil {
			return nil, err
		}
	}

	params["b"] = ""
	sigField := dkimS.FormatHeaderParams(headerName, params)
	sigField = dkimS.GetCanonicalizer(headerCan).CanonicalizeHeader(sigField)
	sigField = strings.TrimRight(sigField, crlf)
	if _, err := io.WriteString(hasher, sigField); err != nil {
		return nil, err
	}
	hashed := hasher.Sum(nil)

	// Don't pass Hash to Sign for ed25519 as it doesn't support it
	// and will return an error ("ed25519: cannot sign hashed message").
	if keyAlgo == "ed25519" {
		hash = crypto.Hash(0)
	}

	sig, err := options.Signer.Sign(randReader, hashed, hash)
	if err != nil {
		return nil, err
	}
	params["b"] = base64.StdEncoding.EncodeToString(sig)

	return &SignerForward{sigParams: params, headerName: headerName}, nil
}

// Signature returns the whole DKIM-Signature header field.
// The returned value contains both the header field name, its value and the
// final CRLF.
func (s *SignerForward) Signature() string {
	if s.sigParams == nil {
		panic("signer: Signer.Signature must only be called after a succesful Signer.Close")
	}
	return dkimS.FormatHeaderParams(s.headerName, s.sigParams)
}

// Sign signs a message. It reads it from r and writes the signed version to w.
// SignSync reads a message from r, signs it, and writes the DKIM header + message to w.
func SignSync(w io.Writer, r io.Reader, email vmimap.Email, params map[string]string, headerName string, options *dkimS.SignOptions) error {
	msg, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	signer, err := NewSignerForward(headerName, options, email, params)
	if err != nil {
		return err
	}

	fmt.Println("--signature--", signer.Signature())
	fmt.Println("--signature b--", signer.sigParams["b"])

	if _, err := io.WriteString(w, signer.Signature()); err != nil {
		return err
	}

	_, err = w.Write(msg)
	return err
}

func formatTagList(l []string) string {
	return strings.Join(l, ":")
}

func formatTime(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
