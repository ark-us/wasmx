package lib

import (
	"encoding/json"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env"
	"github.com/loredanacirstea/wasmx-utils"
)

const MODULE_NAME = "slashing"

// MissedBlockBitmapChunkSize defines the chunk size, in number of bits, of a
// validator missed block bitmap. Chunks are used to reduce the storage and
// write overhead of IAVL nodes.
const MissedBlockBitmapChunkSize int32 = 1024 // 2^10 bits

const ValidatorUpdateDelay int64 = 1

// Infraction represents the type of infraction
type Infraction int32

const (
	// UNSPECIFIED defines an empty infraction
	INFRACTION_UNSPECIFIED Infraction = 0
	// DOUBLE_SIGN defines a validator that double-signs a block
	INFRACTION_DOUBLE_SIGN Infraction = 1
	// DOWNTIME defines a validator that missed signing too many blocks
	INFRACTION_DOWNTIME Infraction = 2
)

// String constants for infractions
const (
	Infraction_INFRACTION_UNSPECIFIED = "INFRACTION_UNSPECIFIED"
	Infraction_INFRACTION_DOUBLE_SIGN = "INFRACTION_DOUBLE_SIGN"
	Infraction_INFRACTION_DOWNTIME    = "INFRACTION_DOWNTIME"
)

// Infraction mappings
var InfractionByString = map[string]Infraction{
	Infraction_INFRACTION_UNSPECIFIED: INFRACTION_UNSPECIFIED,
	Infraction_INFRACTION_DOUBLE_SIGN: INFRACTION_DOUBLE_SIGN,
	Infraction_INFRACTION_DOWNTIME:    INFRACTION_DOWNTIME,
}

var InfractionByEnum = map[Infraction]string{
	INFRACTION_UNSPECIFIED: Infraction_INFRACTION_UNSPECIFIED,
	INFRACTION_DOUBLE_SIGN: Infraction_INFRACTION_DOUBLE_SIGN,
	INFRACTION_DOWNTIME:    Infraction_INFRACTION_DOWNTIME,
}

// GenesisState represents the slashing module's genesis state
type GenesisState struct {
	Params       Params                   `json:"params"`
	SigningInfos []SigningInfo            `json:"signing_infos"`
	MissedBlocks []ValidatorMissedBlocks  `json:"missed_blocks"`
}

// SigningInfo represents signing information for a validator
type SigningInfo struct {
	Address               wasmx.ConsensusAddressString `json:"address"` // e.g. mythosvalcons1....
	ValidatorSigningInfo  ValidatorSigningInfo         `json:"validator_signing_info"`
}

// ValidatorSigningInfo represents detailed signing information for a validator
type ValidatorSigningInfo struct {
	Address             wasmx.ConsensusAddressString `json:"address"` // e.g. mythosvalcons1....
	StartHeight         int64                        `json:"start_height"`          // Height at which validator was first a candidate OR was un-jailed
	IndexOffset         int64                        `json:"index_offset"`
	JailedUntil         time.Time                    `json:"jailed_until"`
	Tombstoned          bool                         `json:"tombstoned"`
	MissedBlocksCounter int64                        `json:"missed_blocks_counter"`
}

// ValidatorMissedBlocks represents missed blocks for a validator
type ValidatorMissedBlocks struct {
	Address      wasmx.ConsensusAddressString `json:"address"`
	MissedBlocks []MissedBlock                `json:"missed_blocks"`
}

// MissedBlock represents a single missed block entry
type MissedBlock struct {
	Index  int64 `json:"index"`
	Missed bool  `json:"missed"`
}

// ParamsExternal is used for JSON serialization with string fields
type ParamsExternal struct {
	SignedBlocksWindow      string `json:"signed_blocks_window"`
	MinSignedPerWindow      string `json:"min_signed_per_window"`
	DowntimeJailDuration    string `json:"downtime_jail_duration"`
	SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
	SlashFractionDowntime   string `json:"slash_fraction_downtime"`
}

// Params represents slashing module parameters
type Params struct {
	SignedBlocksWindow      int64  `json:"signed_blocks_window"`
	MinSignedPerWindow      string `json:"min_signed_per_window"`      // decimal string
	DowntimeJailDuration    string `json:"downtime_jail_duration"`     // duration string
	SlashFractionDoubleSign string `json:"slash_fraction_double_sign"` // decimal string
	SlashFractionDowntime   string `json:"slash_fraction_downtime"`    // decimal string
}

// MarshalJSON implements custom JSON marshaling for Params
func (p Params) MarshalJSON() ([]byte, error) {
	external := ParamsExternal{
		SignedBlocksWindow:      utils.Itoa(int(p.SignedBlocksWindow)),
		MinSignedPerWindow:      p.MinSignedPerWindow,
		DowntimeJailDuration:    p.DowntimeJailDuration,
		SlashFractionDoubleSign: p.SlashFractionDoubleSign,
		SlashFractionDowntime:   p.SlashFractionDowntime,
	}
	return json.Marshal(external)
}

// UnmarshalJSON implements custom JSON unmarshaling for Params
func (p *Params) UnmarshalJSON(data []byte) error {
	var external ParamsExternal
	if err := json.Unmarshal(data, &external); err != nil {
		return err
	}
	
	blocksWindow, err := strconv.ParseInt(external.SignedBlocksWindow, 10, 64)
	if err != nil {
		return err
	}
	
	p.SignedBlocksWindow = blocksWindow
	p.MinSignedPerWindow = external.MinSignedPerWindow
	p.DowntimeJailDuration = external.DowntimeJailDuration
	p.SlashFractionDoubleSign = external.SlashFractionDoubleSign
	p.SlashFractionDowntime = external.SlashFractionDowntime
	return nil
}

// Query types
type QuerySigningInfoRequest struct {
	ConsAddress string `json:"cons_address"` // cons_address is the address to query signing info of
}

type QuerySigningInfoResponse struct {
	ValSigningInfo ValidatorSigningInfo `json:"val_signing_info"` // val_signing_info is the signing info of requested val cons address
}

type QuerySigningInfosRequest struct {
	Pagination wasmx.PageRequest `json:"pagination"`
}

type QuerySigningInfosResponse struct {
	Info       []ValidatorSigningInfo `json:"info"`
	Pagination wasmx.PageResponse     `json:"pagination"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryMissedBlockBitmapRequest struct {
	ConsAddress string `json:"cons_address"`
}

type QueryMissedBlockBitmapResponse struct {
	Chunks []string `json:"chunks"` // base64 encoded chunks
}

// Message types
type MsgRunHook struct {
	Hook string `json:"hook"`
	Data string `json:"data"` // base64 encoded data
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
