package codec

type AddressPrefixed interface {
	Prefix() string
	Bytes() []byte
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	MarshalYAML() (interface{}, error)
	String() string
	// Unmarshal(data []byte, prefix string) error
	// UnmarshalJSON(data []byte) error
	// UnmarshalYAML(data []byte) error
}

// Codec defines an interface to convert addresses from and to string/bytes.
type Codec interface {
	// StringToBytes decodes text to bytes
	StringToBytes(text string) ([]byte, error)
	// BytesToString encodes bytes to text
	BytesToString(bz []byte) (string, error)

	StringToAddressPrefixed(text string) (AddressPrefixed, error)
	BytesToAddressPrefixed(bz []byte) AddressPrefixed
}
