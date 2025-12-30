package lib

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func LoggerInfo(msg string, parts []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, parts)
}

func LoggerError(msg string, parts []string) {
	wasmx.LoggerError(MODULE_NAME, msg, parts)
}

func LoggerDebug(msg string, parts []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, parts)
}

func LoggerDebugExtended(msg string, parts []string) {
	wasmx.LoggerDebugExtended(MODULE_NAME, msg, parts)
}

func Revert(message string) {
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// HexHashToBase64 converts a hex-encoded hash to base64 format
// This is needed because mempool stores hashes in base64, but Kayros uses hex
func HexHashToBase64(hexHash string) (string, error) {
	hashBytes, err := hex.DecodeString(hexHash)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hashBytes), nil
}

// TimeuuidHexToTimestamp extracts timestamp from a Kayros TimeUUID hex string
// TimeUUID format (version 1): time-low(4) time-mid(2) time-high-and-version(2) clock-seq(2) node(6)
// Returns RFC3339Nano formatted timestamp string
func TimeuuidHexToTimestamp(timeuuidHex string) (string, error) {
	// Decode hex string to bytes
	uuidBytes, err := hex.DecodeString(timeuuidHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex: %w", err)
	}

	if len(uuidBytes) != 16 {
		return "", fmt.Errorf("UUIDs must be exactly 16 bytes long")
	}

	// Extract timestamp from TimeUUID (version 1)
	// UUIDs are stored in big-endian (network byte order)
	// The timestamp is split across time_low (bytes 0-3), time_mid (bytes 4-5),
	// and time_hi_and_version (bytes 6-7)

	// Read time_low (32 bits, big-endian)
	timeLow := uint64(uuidBytes[0])<<24 | uint64(uuidBytes[1])<<16 |
		uint64(uuidBytes[2])<<8 | uint64(uuidBytes[3])

	// Read time_mid (16 bits, big-endian)
	timeMid := uint64(uuidBytes[4])<<8 | uint64(uuidBytes[5])

	// Read time_hi_and_version (16 bits, big-endian), mask out version bits (top 4 bits)
	timeHi := (uint64(uuidBytes[6])<<8 | uint64(uuidBytes[7])) & 0x0FFF

	// Reconstruct the 60-bit timestamp
	// timestamp = time_low + (time_mid << 32) + (time_hi << 48)
	timestamp := timeLow | (timeMid << 32) | (timeHi << 48)

	// UUID timestamp is in 100-nanosecond intervals since UUID epoch (1582-10-15)
	// Convert to Unix time (nanoseconds since 1970-01-01)
	const gregorianEpoch = 122192928000000000 // 100-ns intervals between UUID epoch and Unix epoch

	unixNanos := int64((timestamp - gregorianEpoch) * 100)

	// Convert to time.Time
	t := time.Unix(0, unixNanos)

	// Format as RFC3339Nano
	return t.UTC().Format(time.RFC3339Nano), nil
}
