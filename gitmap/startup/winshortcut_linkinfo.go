package startup

// LinkInfo sub-section of a Shell Link. Per [MS-SHLLINK] §2.3 this
// block tells Windows where the target file lives. We populate the
// "LocalBasePath" variant only — sufficient for any local .exe and
// the form WScript.Shell.CreateShortcut produces when only the
// TargetPath property is set.
//
// Layout (all little-endian):
//
//   uint32  LinkInfoSize             // total bytes of this block
//   uint32  LinkInfoHeaderSize       // 0x1C (no optional Unicode offsets)
//   uint32  LinkInfoFlags            // 0x01 = VolumeIDAndLocalBasePath
//   uint32  VolumeIDOffset           // 0x1C (right after header)
//   uint32  LocalBasePathOffset      // 0x1C + sizeof(VolumeID)
//   uint32  CommonNetworkRelativeLinkOffset  // 0 (none)
//   uint32  CommonPathSuffixOffset   // points to a single 0x00 terminator
//   <VolumeID block>
//   <LocalBasePath ASCII, NUL-terminated>
//   <CommonPathSuffix = single 0x00>
//
// We embed a minimal VolumeID (drive type Fixed, no label) because
// some shell versions reject LinkInfo without one even when the
// LocalBasePath flag is set.

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// LinkInfo + VolumeID constants from [MS-SHLLINK] §2.3.
const (
	linkInfoHeaderSize       uint32 = 0x0000001C
	linkInfoFlagVolumeID     uint32 = 0x00000001
	volumeIDHeaderSize       uint32 = 0x00000010
	driveTypeFixed           uint32 = 0x00000003
	driveSerialDefault       uint32 = 0x00000000
	volumeLabelOffsetMissing uint32 = 0x00000014 // "no label" sentinel
)

// buildLinkInfo assembles the LinkInfo block as a single byte slice.
// Returns an error only when the target path contains characters
// LocalBasePath cannot represent (NUL bytes — would break the
// NUL-terminated ASCII encoding).
func buildLinkInfo(target string) ([]byte, error) {
	if bytes.IndexByte([]byte(target), 0) >= 0 {

		return nil, fmt.Errorf("shortcut target contains NUL byte")
	}
	volumeID := buildVolumeID()
	pathBytes := append([]byte(target), 0x00) // NUL-terminated ASCII
	suffixBytes := []byte{0x00}               // empty CommonPathSuffix

	volOff := linkInfoHeaderSize
	pathOff := volOff + uint32(len(volumeID))
	suffixOff := pathOff + uint32(len(pathBytes))
	totalSize := suffixOff + uint32(len(suffixBytes))

	le := binary.LittleEndian
	out := make([]byte, totalSize)
	le.PutUint32(out[0:4], totalSize)
	le.PutUint32(out[4:8], linkInfoHeaderSize)
	le.PutUint32(out[8:12], linkInfoFlagVolumeID)
	le.PutUint32(out[12:16], volOff)
	le.PutUint32(out[16:20], pathOff)
	le.PutUint32(out[20:24], 0) // no CommonNetworkRelativeLink
	le.PutUint32(out[24:28], suffixOff)
	copy(out[volOff:], volumeID)
	copy(out[pathOff:], pathBytes)
	copy(out[suffixOff:], suffixBytes)

	return out, nil
}

// buildVolumeID emits a minimal VolumeID block: 16-byte header,
// fixed drive type, zero serial, "label missing" sentinel. No
// VolumeLabel string follows — the sentinel offset 0x14 is the
// spec-defined "no label" marker and equals VolumeIDSize, which
// Windows treats as end-of-block.
func buildVolumeID() []byte {
	le := binary.LittleEndian
	vol := make([]byte, volumeIDHeaderSize)
	le.PutUint32(vol[0:4], volumeIDHeaderSize)  // VolumeIDSize
	le.PutUint32(vol[4:8], driveTypeFixed)      // DriveType
	le.PutUint32(vol[8:12], driveSerialDefault) // DriveSerialNumber
	le.PutUint32(vol[12:16], volumeLabelOffsetMissing)

	return vol
}
