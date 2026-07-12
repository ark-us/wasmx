package lib

import (
	"testing"
	"time"
)

func TestParseEmailDateSingleDigitDay(t *testing.T) {
	got, err := ParseEmailDate("Fri, 5 Jun 2026 00:24:30 +0200")
	if err != nil {
		t.Fatalf("ParseEmailDate returned error: %v", err)
	}

	want := time.Date(2026, time.June, 5, 0, 24, 30, 0, time.FixedZone("", 2*60*60))
	if !got.Equal(want) {
		t.Fatalf("ParseEmailDate = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestExtractEmailUnfoldsLongCcHeader(t *testing.T) {
	raw := []byte("MIME-Version: 1.0\r\n" +
		"Date: Sun, 12 Jul 2026 14:54:08 +0200\r\n" +
		"Message-ID: <message-id@example.com>\r\n" +
		"Subject: Testing long Cc on multiple lines\r\n" +
		"From: Person One <name@example.com>\r\n" +
		"To: name2@example.com\r\n" +
		"Cc: name3@example.com, name4@example.com, Person Five <name5@example.com>,\r\n" +
		" Person Six <name6@example.com>, name7@example.com, name8@example.com, name9@example.com\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		"Testing long Cc on multiple lines\r\n")

	email, err := extractEmail(raw)
	if err != nil {
		t.Fatalf("extractEmail returned error: %v", err)
	}
	if len(email.Envelope.Cc) != 7 {
		t.Fatalf("Cc length = %d, want 7", len(email.Envelope.Cc))
	}
	if got := email.Envelope.Cc[3].ToAddress(); got != "name6@example.com" {
		t.Fatalf("Cc[3] = %q, want name6@example.com", got)
	}
}
