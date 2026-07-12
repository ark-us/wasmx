package imap

import "testing"

func TestParseEmailAddressesUnfoldsHeaderLines(t *testing.T) {
	input := "name@example.com, name2@example.com, Person Three <name3@example.com>,\r\n Person Four <name4@example.com>, name5@example.com,\n\tname6@example.com, name7@example.com"

	addrs, err := ParseEmailAddresses(input)
	if err != nil {
		t.Fatalf("ParseEmailAddresses returned error: %v", err)
	}
	if len(addrs) != 7 {
		t.Fatalf("ParseEmailAddresses returned %d addresses, want 7", len(addrs))
	}
	if got := addrs[3].ToAddress(); got != "name4@example.com" {
		t.Fatalf("address 3 = %q, want name4@example.com", got)
	}
	if got := addrs[5].ToAddress(); got != "name6@example.com" {
		t.Fatalf("address 5 = %q, want name6@example.com", got)
	}
}
