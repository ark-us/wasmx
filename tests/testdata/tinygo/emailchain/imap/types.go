package imap

import (
	"time"
	"unsafe"

	"github.com/loredanacirstea/emailchain/imap/imapnum"
)

type MailboxAttr string

type CreateOptions struct {
	SpecialUse []MailboxAttr // requires CREATE-SPECIAL-USE
}

type ListOptions struct {
	SelectSubscribed     bool
	SelectRemote         bool
	SelectRecursiveMatch bool // requires SelectSubscribed to be set
	SelectSpecialUse     bool // requires SPECIAL-USE

	ReturnSubscribed bool
	ReturnChildren   bool
	ReturnStatus     *StatusOptions // requires IMAP4rev2 or LIST-STATUS
	ReturnSpecialUse bool           // requires SPECIAL-USE
}

type StatusOptions struct {
	NumMessages bool
	UIDNext     bool
	UIDValidity bool
	NumUnseen   bool
	NumDeleted  bool // requires IMAP4rev2 or QUOTA
	Size        bool // requires IMAP4rev2 or STATUS=SIZE

	AppendLimit    bool // requires APPENDLIMIT
	DeletedStorage bool // requires QUOTA=RES-STORAGE
	HighestModSeq  bool // requires CONDSTORE
}

type Flag string
type AppendOptions struct {
	Flags []Flag
	Time  time.Time
}

type UID uint32
type UIDRange struct {
	Start, Stop UID
}
type UIDSet []UIDRange
type SeqRange struct {
	Start, Stop uint32
}

type SeqSet []SeqRange

// SeqSetNum returns a new SeqSet containing the specified sequence numbers.
func SeqSetNum(nums ...uint32) SeqSet {
	var s SeqSet
	s.AddNum(nums...)
	return s
}

func (s *SeqSet) numSetPtr() *imapnum.Set {
	return (*imapnum.Set)(unsafe.Pointer(s))
}

func (s SeqSet) numSet() imapnum.Set {
	return *s.numSetPtr()
}

func (s SeqSet) String() string {
	return s.numSet().String()
}

func (s SeqSet) Dynamic() bool {
	return s.numSet().Dynamic()
}

// Contains returns true if the non-zero sequence number num is contained in
// the set.
func (s *SeqSet) Contains(num uint32) bool {
	return s.numSet().Contains(num)
}

// Nums returns a slice of all sequence numbers contained in the set.
func (s *SeqSet) Nums() ([]uint32, bool) {
	return s.numSet().Nums()
}

// AddNum inserts new sequence numbers into the set. The value 0 represents "*".
func (s *SeqSet) AddNum(nums ...uint32) {
	s.numSetPtr().AddNum(nums...)
}

// AddRange inserts a new range into the set.
func (s *SeqSet) AddRange(start, stop uint32) {
	s.numSetPtr().AddRange(start, stop)
}

// AddSet inserts all sequence numbers from other into s.
func (s *SeqSet) AddSet(other SeqSet) {
	s.numSetPtr().AddSet(other.numSet())
}

type SearchCriteriaHeaderField struct {
	Key, Value string
}

type SearchCriteriaMetadataType string

const (
	SearchCriteriaMetadataAll     SearchCriteriaMetadataType = "all"
	SearchCriteriaMetadataPrivate SearchCriteriaMetadataType = "priv"
	SearchCriteriaMetadataShared  SearchCriteriaMetadataType = "shared"
)

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

type SearchOptions struct {
	// Requires IMAP4rev2 or ESEARCH
	ReturnMin   bool
	ReturnMax   bool
	ReturnAll   bool
	ReturnCount bool
	// Requires IMAP4rev2 or SEARCHRES
	ReturnSave bool
}
type FetchItemBodyStructure struct {
	Extended bool
}
type PartSpecifier string
type SectionPartial struct {
	Offset, Size int64
}
type FetchItemBodySection struct {
	Specifier       PartSpecifier
	Part            []int
	HeaderFields    []string
	HeaderFieldsNot []string
	Partial         *SectionPartial
	Peek            bool
}
type FetchItemBinarySection struct {
	Part    []int
	Partial *SectionPartial
	Peek    bool
}
type FetchItemBinarySectionSize struct {
	Part []int
}

// FetchOptions contains options for the FETCH command.
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
type StoreFlagsOp int
type StoreFlags struct {
	Op     StoreFlagsOp
	Silent bool
	Flags  []Flag
}

type StoreOptions struct {
	UnchangedSince uint64 // requires CONDSTORE
}
type ListDataChildInfo struct {
	Subscribed bool
}
type StatusData struct {
	Mailbox string

	NumMessages *uint32
	UIDNext     UID
	UIDValidity uint32
	NumUnseen   *uint32
	NumDeleted  *uint32
	Size        *int64

	AppendLimit    *uint32
	DeletedStorage *int64
	// Last modification sequence number (change tracker)
	HighestModSeq uint64
}
type ListData struct {
	Attrs   []MailboxAttr
	Delim   rune
	Mailbox string

	// Extended data
	ChildInfo *ListDataChildInfo
	OldName   string
	Status    *StatusData
}
type SelectData struct {
	// Flags defined for this mailbox
	Flags []Flag
	// Flags that the client can change permanently
	PermanentFlags []Flag
	// Number of messages in this mailbox (aka. "EXISTS")
	NumMessages uint32
	UIDNext     UID
	UIDValidity uint32

	List *ListData // requires IMAP4rev2

	HighestModSeq uint64 // requires CONDSTORE
}

type Range struct {
	Start, Stop uint32
}

type Set []Range

type NumSet interface {
	// String returns the IMAP representation of the message number set.
	String() string
	// Dynamic returns true if the set contains "*" or "n:*" ranges or if the
	// set represents the special SEARCHRES marker.
	Dynamic() bool

	numSet() Set
}

type SearchData struct {
	All UIDSet

	// requires IMAP4rev2 or ESEARCH
	UID   bool
	Min   uint32
	Max   uint32
	Count uint32

	// requires CONDSTORE
	ModSeq uint64
}

type AppendData struct {
	// requires UIDPLUS or IMAP4rev2
	UID         UID
	UIDValidity uint32
}

type CopyData struct {
	// requires UIDPLUS or IMAP4rev2
	UIDValidity uint32
	SourceUIDs  UIDSet
	DestUIDs    UIDSet
}
