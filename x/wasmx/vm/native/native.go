package native

var NativeMap = map[string]func([]byte) []byte{
	"0x0000000000000000000000000000000000000001": Secp256k1RecoverNative,
	"0x0000000000000000000000000000000000000022": SecretSharing,
}
