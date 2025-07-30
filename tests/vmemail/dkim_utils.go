package keeper_test

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

const testPrivateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXwIBAAKBgQDwIRP/UC3SBsEmGqZ9ZJW3/DkMoGeLnQg1fWn7/zYtIxN2SnFC
jxOCKG9v3b4jYfcTNh5ijSsq631uBItLa7od+v/RtdC2UzJ1lWT947qR+Rcac2gb
to/NMqJ0fzfVjH4OuKhitdY9tf6mcwGjaNBcWToIMmPSPDdQPNUYckcQ2QIDAQAB
AoGBALmn+XwWk7akvkUlqb+dOxyLB9i5VBVfje89Teolwc9YJT36BGN/l4e0l6QX
/1//6DWUTB3KI6wFcm7TWJcxbS0tcKZX7FsJvUz1SbQnkS54DJck1EZO/BLa5ckJ
gAYIaqlA9C0ZwM6i58lLlPadX/rtHb7pWzeNcZHjKrjM461ZAkEA+itss2nRlmyO
n1/5yDyCluST4dQfO8kAB3toSEVc7DeFeDhnC1mZdjASZNvdHS4gbLIA1hUGEF9m
3hKsGUMMPwJBAPW5v/U+AWTADFCS22t72NUurgzeAbzb1HWMqO4y4+9Hpjk5wvL/
eVYizyuce3/fGke7aRYw/ADKygMJdW8H/OcCQQDz5OQb4j2QDpPZc0Nc4QlbvMsj
7p7otWRO5xRa6SzXqqV3+F0VpqvDmshEBkoCydaYwc2o6WQ5EBmExeV8124XAkEA
qZzGsIxVP+sEVRWZmW6KNFSdVUpk3qzK0Tz/WjQMe5z0UunY9Ax9/4PVhp/j61bf
eAYXunajbBSOLlx4D+TunwJBANkPI5S9iylsbLs6NkaMHV6k5ioHBBmgCak95JGX
GMot/L2x0IYyMLAz6oLWh2hm7zwtb0CgOrPo1ke44hFYnfc=
-----END RSA PRIVATE KEY-----
`

const testEd25519SeedBase64 = "nWGxne/9WmC6hEr0kuwsxERJxWl7MmkZcDusAxyuf2A="

const mailHeaderString = "From: Joe SixPack <joe@football.example.com>\r\n" +
	"To: Suzie Q <suzie@shopping.example.net>\r\n" +
	"Subject: Is dinner ready?\r\n" +
	"Date: Fri, 11 Jul 2003 21:00:37 -0700 (PDT)\r\n" +
	"Message-ID: <20030712040037.46341.5F8J@football.example.com>\r\n"

const mailBodyString = "Hi.\r\n" +
	"\r\n" +
	"We lost the game. Are you hungry yet?\r\n" +
	"\r\n" +
	"Joe."

const mailString = mailHeaderString + "\r\n" + mailBodyString

// const signedMailString = "DKIM-Signature: a=rsa-sha256; bh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;" + "\r\n" +
// 	" " + "c=simple/simple; d=example.org; h=From:To:Subject:Date:Message-ID;" + "\r\n" +
// 	" " + "s=brisbane; t=424242; v=1;" + "\r\n" +
// 	" " + "b=MobyyDTeHhMhNJCEI6ATNK63ZQ7deSXK9umyzAvYwFqE6oGGvlQBQwqr1aC11hWpktjMLP1/" + "\r\n" +
// 	" " + "m0PBi9v7cRLKMXXBIv2O0B1mIWdZPqd9jveRJqKzCb7SpqH2u5kK6i2vZI639ENTQzRQdxSAGXc" + "\r\n" +
// 	" " + "PcPYjrgkqj7xklnrNBs0aIUA=" + "\r\n" +
// 	mailHeaderString +
// 	"\r\n" +
// 	mailBodyString

const mailStringDkim = "DKIM-Signature: v=1; d=example.org; s=football; i=joe@example.org; a=rsa-sha256; t=424242; h=From:To:Cc:Bcc:Reply-To:References:In-Reply-To:" + "\r\n" +
	" " + "Subject:Date:Message-ID:Content-Type; bh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=; b=dOh/+3OaXjU2VCHG2odzVYq4" + "\r\n" +
	" " + "AUIKzgkyKCxl1KVBF1fBtJo24KHKU4+rOndhsp+htFWVU1G5pFxpYIacVOedMBAdt6lPSHZBe1p" + "\r\n" +
	" " + "Rr6T404Dpc7UIiZypfRAnoEsC6pZeuCmnpX/H0olAkSk4Pdel2e+iud92JQMwGlgHy/a3JcQ="

const signedMailString = mailStringDkim + "\r\n" +
	mailHeaderString +
	"\r\n" +
	mailBodyString

var (
	testPrivateKey        *rsa.PrivateKey
	testEd25519PrivateKey ed25519.PrivateKey
)

func init() {
	block, _ := pem.Decode([]byte(testPrivateKeyPEM))
	var err error
	testPrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	ed25519Seed, err := base64.StdEncoding.DecodeString(testEd25519SeedBase64)
	if err != nil {
		panic(err)
	}
	testEd25519PrivateKey = ed25519.NewKeyFromSeed(ed25519Seed)
}

func appendDKIM(mailStringDkim string, restofemail string) string {
	return mailStringDkim + "\r\n" + restofemail
}
