// IKEMEN:VOID dual-mode ZIP loading — HyperNull / WinMUGEN unzip 0.15 + zlib 1.1.3 parity.
//
// WinMUGEN char ZIP path (unzip 0.15 / zlib 1.1.3) had no uncompressed-size caps, deferred CRC,
// and inflate that could overflow / soft-corrupt. HyperNull cheapies rely on that permissiveness.
//
// Dual behavior:
//   High-tier / HyperNull — soft ceilings, ignore CRC, soft-fail inflate → partial data, continue load
//   Normal                — modern safe caps + hard fail on inflate/CRC/oversize
//
// We do NOT reintroduce CVE-2005-2096 heap overflows (would kill the host). We emulate the
// *observable* WinMUGEN outcomes HyperNull needs: oversized declares accepted, corrupt streams
// still yield bytes, load continues (Zero-Crash).
package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

// Safe ceiling for untagged / normal characters (zip-bomb guard).
const voidZipSafeMaxUncompressed int64 = 64 << 20 // 64 MiB

// Soft ceiling for HyperNull / high-tier — WinMUGEN malloc'd attacker-controlled size with no
// upper bound; we soft-cap to keep the host alive while still allowing large payload buffers.
const voidZipHyperNullMaxUncompressed int64 = 512 << 20 // 512 MiB

// Declared size above safe max (but used for HyperNull detection / escalation).
const voidZipHyperNullDeclareThreshold int64 = voidZipSafeMaxUncompressed + 1

var (
	voidZipPermissiveHintMu sync.Mutex
	// Logical zip path (or basename) → force permissive extract for this archive.
	voidZipPermissiveHint = map[string]bool{}
)

func voidZipNormalizeKey(zipPath string) string {
	return strings.ToLower(filepath.ToSlash(zipPath))
}

// voidZipMarkPermissive flags an archive for WinMUGEN-style extract (HyperNull).
func voidZipMarkPermissive(zipPath string) {
	if zipPath == "" {
		return
	}
	voidZipPermissiveHintMu.Lock()
	voidZipPermissiveHint[voidZipNormalizeKey(zipPath)] = true
	voidZipPermissiveHintMu.Unlock()
}

func voidZipHintPermissive(zipPath string) bool {
	voidZipPermissiveHintMu.Lock()
	defer voidZipPermissiveHintMu.Unlock()
	return voidZipPermissiveHint[voidZipNormalizeKey(zipPath)]
}

// voidZipHyperNullNameSignal — folder/zip naming without character-identity checks.
func voidZipHyperNullNameSignal(path string) bool {
	low := strings.ToLower(filepath.ToSlash(path))
	for _, tok := range []string{"hypernull", "hyper_null", "hyper-null", "_null.zip", "null.zip"} {
		if strings.Contains(low, tok) {
			return true
		}
	}
	return false
}

// voidZipPermissiveActive decides WinMUGEN-faithful ZIP extract for this archive/entry.
func voidZipPermissiveActive(zipPath string) bool {
	if voidZipHintPermissive(zipPath) || voidZipHyperNullNameSignal(zipPath) {
		return true
	}
	// During character load, workingChar / budget char may already be high-tier.
	if c := sys.workingChar; c != nil && voidHighTier(c) {
		return true
	}
	if c := sys.voidBudgetChar; c != nil && voidHighTier(c) {
		return true
	}
	// Any roster slot whose DEF lives in this zip and is already void-tagged.
	key := voidZipNormalizeKey(zipPath)
	for pn := range sys.cgi {
		def := voidZipNormalizeKey(sys.cgi[pn].def)
		if def == "" {
			continue
		}
		if strings.HasPrefix(def, key+"/") || strings.HasPrefix(def, key) {
			if sys.cgi[pn].voidTagged() || sys.cgi[pn].voidTier > VoidTierNull {
				return true
			}
		}
	}
	return false
}

// voidZipScanAndEscalate — inspect central directory; escalate / mark permissive on HyperNull signals.
func voidZipScanAndEscalate(pn int, zipPath string) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return
	}
	defer zr.Close()
	hyper := voidZipHyperNullNameSignal(zipPath)
	var reasons []string
	for _, f := range zr.File {
		sz := int64(f.UncompressedSize64)
		if sz >= voidZipHyperNullDeclareThreshold {
			hyper = true
			reasons = append(reasons, fmt.Sprintf("huge_uncomp:%s:%d", filepath.Base(f.Name), sz))
		}
		// Extra-field / filename lengths that unzip 0.15 would mishandle in position calc.
		if len(f.Name) >= 256 {
			hyper = true
			reasons = append(reasons, "filename_256:"+filepath.Base(f.Name))
		}
		if len(f.Extra) >= 0x7FFF {
			hyper = true
			reasons = append(reasons, "extra_overflow:"+filepath.Base(f.Name))
		}
	}
	if !hyper {
		return
	}
	voidZipMarkPermissive(zipPath)
	if pn >= 0 && pn < len(sys.cgi) {
		detail := strings.Join(reasons, "; ")
		if detail == "" {
			detail = "hypernull_zip_signal"
		}
		voidEscalatePlayerTier(pn, VoidTierUltranull, "hypernull_zip:"+detail)
		sys.cgi[pn].supernullChar = true
		sys.voidRefreshMatchExtender()
		LogMessage("IKEMEN:VOID: HyperNull ZIP signals on P%v — %s", pn+1, detail)
	}
}

// voidZipMaxUncompressed returns the per-entry ceiling for this mode.
func voidZipMaxUncompressed(permissive bool) int64 {
	if permissive {
		return voidZipHyperNullMaxUncompressed
	}
	return voidZipSafeMaxUncompressed
}

// voidZipReadEntry extracts one zip.File with dual-mode policy.
// Normal: hard size cap + CRC via File.Open + full ReadAll; fail on error.
// HyperNull: soft cap, OpenRaw+flate (no CRC), soft-fail → partial bytes.
func voidZipReadEntry(f *zip.File, permissive bool) ([]byte, error) {
	declared := int64(f.UncompressedSize64)
	max := voidZipMaxUncompressed(permissive)

	if !permissive {
		if declared > max {
			return nil, fmt.Errorf("zip entry %q uncompressed size %d exceeds safe limit %d", f.Name, declared, max)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		// LimitReader as defense-in-depth even when header lies small.
		data, err := io.ReadAll(io.LimitReader(rc, max+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > max {
			return nil, fmt.Errorf("zip entry %q decompressed past safe limit %d", f.Name, max)
		}
		return data, nil
	}

	// --- HyperNull / WinMUGEN-permissive ---
	if declared > max {
		// WinMUGEN would malloc(declared) and often OOM/crash; we soft-cap and continue.
		voidLogPermissiveLoad("zip_hypernull_size", f.Name,
			fmt.Sprintf("declared=%d soft_cap=%d", declared, max))
	}
	raw, err := f.OpenRaw()
	if err != nil {
		// Fall back to standard Open; still soft-fail on read errors.
		rc, err2 := f.Open()
		if err2 != nil {
			voidLogPermissiveLoad("zip_hypernull_open", f.Name, err2.Error())
			return []byte{}, nil // Zero-Crash: empty payload, keep char slot alive
		}
		defer rc.Close()
		data, err3 := io.ReadAll(io.LimitReader(rc, max))
		if err3 != nil {
			voidLogPermissiveLoad("zip_hypernull_read", f.Name, err3.Error()+" — partial/empty")
		}
		return data, nil
	}

	var reader io.Reader = raw
	var closer io.Closer
	switch f.Method {
	case zip.Deflate:
		fr := flate.NewReader(raw)
		closer = fr
		reader = fr
	case zip.Store:
		// raw is already stored payload
	default:
		// Unknown method — try flate anyway (WinMUGEN sometimes still fed inflate).
		fr := flate.NewReader(raw)
		closer = fr
		reader = fr
	}
	if closer != nil {
		defer closer.Close()
	}

	buf := &bytes.Buffer{}
	_, copyErr := io.Copy(buf, io.LimitReader(reader, max))
	if copyErr != nil && copyErr != io.EOF && copyErr != io.ErrUnexpectedEOF {
		// Inflate errors (corrupt Huffman / truncated stream): keep partial — Zero-Crash.
		voidLogPermissiveLoad("zip_hypernull_inflate", f.Name,
			fmt.Sprintf("%v — returning %d partial bytes (WinMUGEN soft-corrupt parity)", copyErr, buf.Len()))
	}
	return buf.Bytes(), nil
}

// voidZipOpenFile is the dual-mode replacement body for zip entries in OpenFile.
func voidZipOpenFile(zipFilePath, pathInZip string) (io.ReadSeekCloser, error) {
	zr, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("opening zip archive %s: %w", zipFilePath, err)
	}
	if pathInZip == "" {
		zr.Close()
		return nil, fmt.Errorf("path inside zip archive not specified for %s", zipFilePath)
	}

	permissive := voidZipPermissiveActive(zipFilePath)
	// Opportunistic HyperNull scan (escalates + marks) when working player known.
	if pn := -1; sys.workingChar != nil {
		pn = sys.workingChar.playerNo
		voidZipScanAndEscalate(pn, zipFilePath)
		permissive = voidZipPermissiveActive(zipFilePath)
	}

	pathInZipLower := strings.ToLower(pathInZip)
	var targetFile *zip.File
	for _, f := range zr.File {
		if strings.ToLower(filepath.ToSlash(f.Name)) == pathInZipLower {
			targetFile = f
			break
		}
	}
	// HyperNull: basename-only fallback (WinMUGEN locate was looser).
	if targetFile == nil && permissive {
		baseWant := strings.ToLower(filepath.Base(pathInZip))
		for _, f := range zr.File {
			if strings.ToLower(filepath.Base(f.Name)) == baseWant {
				targetFile = f
				voidLogPermissiveLoad("zip_hypernull_basename", zipFilePath,
					pathInZip+" → "+f.Name)
				break
			}
		}
	}
	if targetFile == nil {
		zr.Close()
		return nil, fmt.Errorf("file '%s' not found in zip archive '%s'", pathInZip, zipFilePath)
	}

	fileData, err := voidZipReadEntry(targetFile, permissive)
	if err != nil {
		zr.Close()
		return nil, fmt.Errorf("reading file '%s' from zip archive '%s': %w", pathInZip, zipFilePath, err)
	}
	return &zipMemFileReader{reader: bytes.NewReader(fileData), zipArchive: zr}, nil
}
