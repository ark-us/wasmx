package keeper

import (
	"encoding/hex"
	"io"
	"math"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	snapshot "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/ioutils"
	"mythos/v1/x/wasmx/types"
)

var _ snapshot.ExtensionSnapshotter = &Utf8Snapshotter{}

// SnapshotFormat format 1 is just gzipped utf8 source code for each item payload. No protobuf envelope, no metadata.
const SnapshotFormatUtf8 = 1

type Utf8Snapshotter struct {
	wasmx *Keeper
	cms   storetypes.MultiStore
}

func NewUtf8Snapshotter(cms storetypes.MultiStore, wasmx *Keeper) *Utf8Snapshotter {
	return &Utf8Snapshotter{
		wasmx: wasmx,
		cms:   cms,
	}
}

func (ws *Utf8Snapshotter) SnapshotName() string {
	return types.ModuleName + "_utf8"
}

func (ws *Utf8Snapshotter) SnapshotFormat() uint32 {
	return SnapshotFormatUtf8
}

func (ws *Utf8Snapshotter) SupportedFormats() []uint32 {
	// If we support older formats, add them here and handle them in Restore
	return []uint32{SnapshotFormatUtf8}
}

func (ws *Utf8Snapshotter) SnapshotExtension(height uint64, payloadWriter snapshot.ExtensionPayloadWriter) error {
	cacheMS, err := ws.cms.CacheMultiStoreWithVersion(int64(height))
	if err != nil {
		return err
	}

	ctx := sdk.NewContext(cacheMS, tmproto.Header{}, false, log.NewNopLogger())
	seenBefore := make(map[string]bool)
	var rerr error

	ws.wasmx.IterateCodeInfos(ctx, func(id uint64, info types.CodeInfo) bool {
		// Many code ids may point to the same code hash... only sync it once
		hexHash := hex.EncodeToString(info.CodeHash)
		// if seenBefore, just skip this one and move to the next
		if seenBefore[hexHash] {
			return false
		}
		seenBefore[hexHash] = true

		if len(info.Deps) == 0 || !types.HasUtf8Dep(info.Deps) {
			return false
		}

		// load code and skip on error
		// TODO fixme if it has utf8 dep, it should not error

		extension := GetExtensionFromDeps(info.Deps)
		fileBytes, err := ws.wasmx.wasmvm.load_utf8(extension, info.CodeHash)
		if err != nil {
			// TODO fixme now we just skip
			// rerr = err
			// return true
			return false
		}

		fileBytes = packFile(extension, fileBytes)
		compressedFile, err := ioutils.GzipIt(fileBytes)
		if err != nil {
			rerr = err
			return true
		}

		err = payloadWriter(compressedFile)
		if err != nil {
			rerr = err
			return true
		}

		return false
	})

	return rerr
}

func (ws *Utf8Snapshotter) RestoreExtension(height uint64, format uint32, payloadReader snapshot.ExtensionPayloadReader) error {
	if format == SnapshotFormatUtf8 {
		return ws.processAllItems(height, payloadReader, restoreUtf8V1, finalizeUtf8V1)
	}
	return snapshot.ErrUnknownFormat
}

func restoreUtf8V1(_ sdk.Context, k *Keeper, compressedCode []byte) error {
	if !ioutils.IsGzip(compressedCode) {
		return types.ErrInvalid.Wrap("not a gzip")
	}
	fileBz, err := ioutils.Uncompress(compressedCode, math.MaxInt64)
	if err != nil {
		return errorsmod.Wrap(types.ErrCreateFailed, err.Error())
	}
	extension, fileBytes := unpackFile(fileBz)
	_, err = k.wasmvm.CreateUtf8(fileBytes, extension)
	if err != nil {
		return errorsmod.Wrap(types.ErrCreateFailed, err.Error())
	}
	return nil
}

func finalizeUtf8V1(ctx sdk.Context, k *Keeper) error {
	// FIXME: ensure all codes have been uploaded?
	k.IterateCodeInfos(ctx, func(id uint64, info types.CodeInfo) bool {
		if !info.Pinned {
			return false
		}
		k.PinCode(ctx, id, "")
		return false
	})
	return nil
}

func (ws *Utf8Snapshotter) processAllItems(
	height uint64,
	payloadReader snapshot.ExtensionPayloadReader,
	cb func(sdk.Context, *Keeper, []byte) error,
	finalize func(sdk.Context, *Keeper) error,
) error {
	ctx := sdk.NewContext(ws.cms, tmproto.Header{Height: int64(height)}, false, log.NewNopLogger())
	for {
		payload, err := payloadReader()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := cb(ctx, ws.wasmx, payload); err != nil {
			return errorsmod.Wrap(err, "processing snapshot item")
		}
	}

	return finalize(ctx, ws.wasmx)
}

// 4b extension length + extension + file bz
func packFile(extension string, fileBz []byte) []byte {
	extensionbz := []byte(extension)
	extlenbz := sdk.Uint64ToBigEndian(uint64(len(extensionbz)))
	extensionPartBz := append(extlenbz, extensionbz...)
	return append(extensionPartBz, fileBz...)
}

func unpackFile(fileBz []byte) (string, []byte) {
	extlen := sdk.BigEndianToUint64(fileBz[0:8])
	extension := string(fileBz[8:(8 + extlen)])
	fileBytes := fileBz[(8 + extlen):]
	return extension, fileBytes
}
