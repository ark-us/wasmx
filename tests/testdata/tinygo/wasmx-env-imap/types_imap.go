package imap

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type SeqSet []SeqRange
type UIDSet []UIDRange
type Flag string
type SearchCriteriaMetadataType string
type UID uint32
type MailboxAttr string
type PartSpecifier string

type SeqRange struct {
	Start, Stop uint32
}

type UIDRange struct {
	Start, Stop UID
}

type SearchCriteriaHeaderField struct {
	Key, Value string
}

type SearchCriteriaModSeq struct {
	ModSeq       uint64
	MetadataName string
	MetadataType SearchCriteriaMetadataType
}

type SearchCriteria struct {
	SeqNum []SeqSet
	UID    []UIDSet

	// Only the date is used, the time and timezone are ignored
	Since      time.Time
	Before     time.Time
	SentSince  time.Time
	SentBefore time.Time

	Header []SearchCriteriaHeaderField
	Body   []string
	Text   []string

	Flag    []Flag
	NotFlag []Flag

	Larger  int64
	Smaller int64

	Not []SearchCriteria
	Or  [][2]SearchCriteria

	ModSeq *SearchCriteriaModSeq // requires CONDSTORE
}

type FetchItemBodyStructure struct {
	Extended bool
}

type FetchItemBinarySection struct {
	Part    []int
	Partial *SectionPartial
	Peek    bool
}

type FetchItemBinarySectionSize struct {
	Part []int
}

type SectionPartial struct {
	Offset, Size int64
}

type FetchOptions struct {
	// Fields to fetch
	BodyStructure     *FetchItemBodyStructure
	Envelope          bool
	Flags             bool
	InternalDate      bool
	RFC822Size        bool
	UID               bool
	BodySection       []*FetchItemBodySection
	BinarySection     []*FetchItemBinarySection     // requires IMAP4rev2 or BINARY
	BinarySectionSize []*FetchItemBinarySectionSize // requires IMAP4rev2 or BINARY
	ModSeq            bool                          // requires CONDSTORE

	ChangedSince uint64 // requires CONDSTORE
}

type FetchItemBodySection struct {
	Specifier       PartSpecifier
	Part            []int
	HeaderFields    []string
	HeaderFieldsNot []string
	Partial         *SectionPartial
	Peek            bool
}

type Address struct {
	Name    string
	Mailbox string
	Host    string
}

func (a Address) ToString() string {
	addr := a.ToAddress()
	if a.Name == "" {
		return addr
	}
	return fmt.Sprintf("%s <%s>", a.Name, addr)
}

func (a Address) ToAddress() string {
	return fmt.Sprintf("%s@%s", a.Mailbox, a.Host)
}

func AddressFromString(account string, name string) Address {
	parts := strings.Split(account, "@")
	return Address{name, parts[0], parts[1]}
}

func AddressesFromString(accounts []string) []Address {
	addrs := []Address{}
	for _, v := range accounts {
		addrs = append(addrs, AddressFromString(v, ""))
	}
	return addrs
}

func ParseEmailAddresses(input string) ([]Address, error) {
	addrList, err := mail.ParseAddressList(input)
	if err != nil {
		return nil, err
	}
	var result []Address
	for _, addr := range addrList {
		local, domain := SplitAddress(addr.Address)
		result = append(result, Address{
			Name:    addr.Name,
			Mailbox: local,
			Host:    domain,
		})
	}
	return result, nil
}

func SplitAddress(full string) (local, domain string) {
	parts := strings.SplitN(full, "@", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return full, ""
}

func ToAddresses(addresses []Address) []string {
	addrs := make([]string, len(addresses))
	for i, addr := range addresses {
		addrs[i] = addr.ToAddress()
	}
	return addrs
}

func SerializeAddresses(addresses []Address) string {
	addrs := make([]string, len(addresses))
	for i, addr := range addresses {
		addrs[i] = addr.ToString()
	}
	return strings.Join(addrs, ", ")
}

func SerializeMessageId(messageIds string) string {
	return fmt.Sprintf("<%s>", messageIds)
}

func SerializeMessageIds(messageIds []string) string {
	ids := make([]string, len(messageIds))
	for i, id := range messageIds {
		ids[i] = SerializeMessageId(id)
	}
	return strings.Join(ids, " ")
}

type Envelope struct {
	Date      time.Time
	Subject   string
	From      []Address
	Sender    []Address
	ReplyTo   []Address
	To        []Address
	Cc        []Address
	Bcc       []Address
	InReplyTo []string
	MessageID string
}

type CreateOptions struct {
	SpecialUse []MailboxAttr // requires CREATE-SPECIAL-USE
}
