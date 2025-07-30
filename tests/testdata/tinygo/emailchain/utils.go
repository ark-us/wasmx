package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/loredanacirstea/emailchain/imap"
	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/utils"
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

func BuildMessageID(opts SignOptions, date time.Time) (string, error) {
	return GenerateMessageID(opts.Selector+"."+opts.Domain, date)
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

func GetAttrs(folder string) []imap.MailboxAttr {
	attrs := []imap.MailboxAttr{}
	switch folder {
	case FolderInbox:
		attrs = []imap.MailboxAttr{"\\Inbox"}
	case FolderSent:
		attrs = []imap.MailboxAttr{"\\Sent"}
	case FolderArchive:
		attrs = []imap.MailboxAttr{"\\Archive"}
	case FolderDraft:
		attrs = []imap.MailboxAttr{"\\Draft"}
	case FolderJunk:
		attrs = []imap.MailboxAttr{"\\Junk"}
	case FolderSpam:
		attrs = []imap.MailboxAttr{"\\Spam"}
	case FolderTrash:
		attrs = []imap.MailboxAttr{"\\Trash"}
	}
	return attrs
}

func PtrUint32(v uint32) *uint32 { return &v }
func PtrInt64(v int64) *int64    { return &v }

func extractHeaders(raw []byte, headers []string) ([]string, error) {
	msg := strings.NewReader(string(raw))
	hdrs, _, err := utils.ParseHeaders(bufio.NewReader(&utils.AtReader{R: msg}))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dkim.ErrHeaderMalformed, err)
	}
	values := make([]string, len(headers))
	for _, h := range hdrs {
		ndx := slices.Index(headers, h.Key)
		if ndx > -1 {
			values[ndx] = h.GetValueTrimmed()
		}
	}
	return values, nil
}
