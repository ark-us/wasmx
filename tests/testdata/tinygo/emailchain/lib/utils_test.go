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
