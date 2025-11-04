package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	IMax = int32(^uint32(0) >> 1)
	IErr = ^IMax
)

func Random() int32 {
	w := sys.randseed / 127773
	sys.randseed = (sys.randseed-w*127773)*16807 - w*2836
	if sys.randseed <= 0 {
		sys.randseed += IMax - Btoi(sys.randseed == 0)
	}
	return sys.randseed
}

func Srand(s int32) {
	sys.randseed = s
}

func Rand(min, max int32) int32 {
	return min + Random()/(IMax/(max-min+1)+1)
}

func RandF32(min, max float32) float32 {
	return min + float32(Random())/(float32(IMax)/(max-min+1.0)+1.0)
}

func RandI(x, y int32) int32 {
	if y < x {
		if uint32(x-y) > uint32(IMax) {
			return int32(int64(y) + int64(Random())*(int64(x)-int64(y))/int64(IMax))
		}
		return Rand(y, x)
	}
	if uint32(y-x) > uint32(IMax) {
		return int32(int64(x) + int64(Random())*(int64(y)-int64(x))/int64(IMax))
	}
	return Rand(x, y)
}

func RandF(x, y float32) float32 {
	return x + float32(Random())*(y-x)/float32(IMax)
}

func Min(arg ...int32) (min int32) {
	for i, x := range arg {
		if i == 0 || x < min {
			min = x
		}
	}
	return
}

func Max(arg ...int32) (max int32) {
	for i, x := range arg {
		if i == 0 || x > max {
			max = x
		}
	}
	return
}

func MinF(arg ...float32) (min float32) {
	for i, x := range arg {
		if i == 0 || x < min {
			min = x
		}
	}
	return
}

func MaxF(arg ...float32) (max float32) {
	for i, x := range arg {
		if i == 0 || x > max {
			max = x
		}
	}
	return
}

func MinI(arg ...int) (min int) {
	for i, x := range arg {
		if i == 0 || x < min {
			min = x
		}
	}
	return
}

func MaxI(arg ...int) (max int) {
	for i, x := range arg {
		if i == 0 || x > max {
			max = x
		}
	}
	return
}

func Clamp(x, a, b int32) int32 {
	return Max(a, Min(x, b))
}

func ClampF(x, a, b float32) float32 {
	return MaxF(a, MinF(x, b))
}

func Rad(f float32) float32 {
	return float32(f * math.Pi / 180)
}

func Deg(f float32) float32 {
	return float32(f * 180 / math.Pi)
}

func Cos(f float32) float32 {
	return float32(math.Cos(float64(f)))
}

func Sin(f float32) float32 {
	return float32(math.Sin(float64(f)))
}

func Sign(i int32) int32 {
	if i < 0 {
		return -1
	} else if i > 0 {
		return 1
	} else {
		return 0
	}
}

func SignF(f float32) float32 {
	if f < 0 {
		return -1
	} else if f > 0 {
		return 1
	} else {
		return 0
	}
}

func Abs(i int32) int32 {
	if i < 0 {
		return -i
	}
	return i
}

func AbsF(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

func Pow(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}

func Lerp(x, y, a float32) float32 {
	//return float32(x + (y - x) * ClampF(a, 0, 1))
	return float32((1-a)*x + a*y)
}

func Ceil(x float32) int32 {
	return int32(math.Ceil(float64(x)))
}

func Floor(x float32) int32 {
	return int32(math.Floor(float64(x)))
}

func IsFinite(f float32) bool {
	return math.Abs(float64(f)) <= math.MaxFloat64
}

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil
}

func Atoi(str string) int32 {
	var n int64
	str = strings.TrimSpace(str)
	if len(str) > 0 {
		var a string
		if str[0] == '-' || str[0] == '+' {
			a = str[1:]
		} else {
			a = str
		}
		for i := range a {
			if a[i] < '0' || '9' < a[i] {
				break
			}
			n = n*10 + int64(a[i]-'0')
			if n > 2147483647 {
				sys.appendToConsole(fmt.Sprintf("WARNING: Atoi conversion outside int32 range: %v", a[:i+1]))
				sys.errLog.Printf("Atoi conversion outside int32 range: %v\n", a[:i+1])
				if str[0] == '-' {
					return IErr
				}
				return IMax
			}
		}
		if str[0] == '-' {
			n *= -1
		}
	}
	return int32(n)
}

func Atof(str string) float64 {
	f := 0.0
	str = strings.TrimSpace(str)
	if len(str) > 0 {
		var a string
		if str[0] == '-' || str[0] == '+' {
			a = str[1:]
		} else {
			a = str
		}
		i, p := 0, 0
		for ; i < len(a); i++ {
			if a[i] == '.' {
				if p != 0 {
					break
				}
				p = i + 1
				continue
			}
			if a[i] < '0' || '9' < a[i] {
				break
			}
			f = f*10 + float64(a[i]-'0')
		}
		e := 0.0
		if i+1 < len(a) && (a[i] == 'e' || a[i] == 'E') {
			j := i + 1
			if a[j] == '-' || a[j] == '+' {
				j++
			}
			for ; j < len(a) && '0' <= a[j] && a[j] <= '9'; j++ {
				e = e*10 + float64(a[j]-'0')
			}
			if e != 0 {
				if str[i+1] == '-' {
					e *= -1
				}
				if p == 0 {
					p = i
				}
			}
		}
		if p > 0 {
			f *= math.Pow10(p - i + int(e))
		}
		if str[0] == '-' {
			f *= -1
		}
	}
	return f
}

func RectRotate(x, y, w, h, cx, cy, angle float32) [][2]float32 {
	rp := make([][2]float32, 4)
	corners := [][2]float32{
		{x, y},
		{x + w, y},
		{x + w, y + h},
		{x, y + h},
	}
	for i := 0; i < 4; i++ {
		p := corners[i]
		tx := p[0] - cx
		ty := p[1] - cy
		newX := Cos(angle)*tx - Sin(angle)*ty + cx
		newY := Sin(angle)*tx + Cos(angle)*ty + cy
		rp[i] = [2]float32{newX, newY}
	}
	return rp
}

// Check if two rotated rectangles intersect.
func RectIntersect(x1, y1, w1, h1, x2, y2, w2, h2, cx1, cy1, cx2, cy2, angle1, angle2 float32) bool {
	rect1 := RectRotate(x1, y1, w1, h1, cx1, cy1, angle1)
	rect2 := RectRotate(x2, y2, w2, h2, cx2, cy2, angle2)
	projectionRange := func(rect [][2]float32, normal [2]float32) (min float32, max float32) {
		min = math.MaxFloat32
		max = -math.MaxFloat32
		for _, p := range rect {
			projected := normal[0]*p[0] + normal[1]*p[1]
			if projected < min {
				min = projected
			}
			if projected > max {
				max = projected
			}
		}
		return
	}
	for i1 := 0; i1 < len(rect1); i1++ {
		i2 := (i1 + 1) % len(rect1)
		p1 := rect1[i1]
		p2 := rect1[i2]
		nx := p2[1] - p1[1]
		ny := p1[0] - p2[0]
		minA, maxA := projectionRange(rect1, [2]float32{nx, ny})
		minB, maxB := projectionRange(rect2, [2]float32{nx, ny})
		if maxA < minB || maxB < minA {
			return false
		}
	}
	for i1 := 0; i1 < len(rect2); i1++ {
		i2 := (i1 + 1) % len(rect2)
		p1 := rect2[i1]
		p2 := rect2[i2]
		nx := p2[1] - p1[1]
		ny := p1[0] - p2[0]
		minA, maxA := projectionRange(rect1, [2]float32{nx, ny})
		minB, maxB := projectionRange(rect2, [2]float32{nx, ny})
		if maxA < minB || maxB < minA {
			return false
		}
	}
	return true
}

// Prevent overflow errors when converting float64 to int32
func F64toI32(f float64) int32 {
	if f >= float64(math.MaxInt32) {
		return math.MaxInt32
	}
	if f <= float64(math.MinInt32) {
		return math.MinInt32
	}
	return int32(f)
}

func readDigit(d string) (int32, bool) {
	if len(d) == 0 || (len(d) >= 2 && d[0] == '0') {
		return 0, false
	}
	for _, c := range d {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	return int32(Atof(d)), true
}

func Btoi(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func I32ToI16(i32 int32) int16 {
	if i32 < ^int32(^uint16(0)>>1) {
		return ^int16(^uint16(0) >> 1)
	}
	if i32 > int32(^uint16(0)>>1) {
		return int16(^uint16(0) >> 1)
	}
	return int16(i32)
}

func I32ToU16(i32 int32) uint16 {
	if i32 < 0 {
		return 0
	}
	if i32 > int32(^uint16(0)) {
		return ^uint16(0)
	}
	return uint16(i32)
}

func RoundFloat(val float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(val*factor) / factor
}

func NormalizeNewlines(input string) string {
	// Replace CRLF (\r\n) with LF (\n)
	input = strings.ReplaceAll(input, "\r\n", "\n")
	// Replace any remaining CR (\r) with LF (\n)
	input = strings.ReplaceAll(input, "\r", "\n")
	return input
}

func LoadText(filename string) (string, error) {
	rc, err := OpenFile(filename)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	bytes, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	if len(bytes) >= 3 && bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf { // UTF-8 BOM
		return string(bytes[3:]), nil
	}

	return string(bytes), nil
}

func decodeShiftJIS(input string) string {
	bytes := []byte(input)

	if utf8.Valid(bytes) {
		return input
	}

	decodedBytes, _, err := transform.Bytes(japanese.ShiftJIS.NewDecoder(), bytes)
	if err != nil {
		sys.errLog.Printf("Warning: Failed to decode string as Shift_JIS, falling back to original. String: %s, Error: %v\n", input, err)
		return input
	}
	return string(decodedBytes)
}

func FileExist(filename string) string {
	filename = filepath.ToSlash(filename)
	isZip, zipFilePath, pathInZip := IsZipPath(filename)
	if isZip {
		var actualZipFilePathOnDisk string
		info, err := os.Stat(zipFilePath)
		if err == nil && !info.IsDir() {

			actualZipFilePathOnDisk = zipFilePath
		} else if os.IsNotExist(err) || (info != nil && info.IsDir()) {
			var pattern string
			tempPattern := ""
			for _, r := range zipFilePath {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					tempPattern += "[" + string(unicode.ToLower(r)) + string(unicode.ToUpper(r)) + "]"
				} else {
					tempPattern += string(r)
				}
			}
			pattern = tempPattern

			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				// Ensure the found match is not a directory
				infoMatch, errMatch := os.Stat(matches[0])
				if errMatch == nil && !infoMatch.IsDir() {
					actualZipFilePathOnDisk = matches[0]
				} else {
					return ""
				}
			} else {
				return "" // Zip file not found by direct stat or glob
			}
		} else if err != nil {
			return "" // Some other error stating the zip file
		}
		// At this point, actualZipFilePathOnDisk should be the correct path to the zip file.
		// Now check if pathInZip is empty (meaning we only check for the zip file itself)
		if pathInZip == "" {
			return filepath.ToSlash(actualZipFilePathOnDisk)
		}

		// If pathInZip is specified, check for its existence within the archive
		zr, err := zip.OpenReader(actualZipFilePathOnDisk)
		if err != nil {
			return "" // Cannot open zip
		}
		defer zr.Close()
		pathInZipLower := strings.ToLower(pathInZip)
		foundInZip := false
		for _, f := range zr.File {
			if strings.ToLower(filepath.ToSlash(f.Name)) == pathInZipLower {
				foundInZip = true
				break
			}
		}
		if foundInZip {
			// Return the logical path (original filename) if the file exists in the zip
			return filename
		}
		return "" // File not found in zip
	}
	if info, err := os.Stat(filename); !os.IsNotExist(err) {
		if info == nil || info.IsDir() {
			return ""
		}
		return filepath.ToSlash(filename)
	}
	var pattern string
	for _, r := range filename {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			pattern += "[" + string(unicode.ToLower(r)) +
				string(unicode.ToLower(r)+'A'-'a') + "]"
		} else if r == '*' || r == '?' || r == '[' {
			pattern += "\\" + string(r)
		} else {
			pattern += string(r)
		}
	}
	if m, _ := filepath.Glob(pattern); len(m) > 0 {
		info, _ := os.Stat(m[0])
		if info != nil && !info.IsDir() {
			return filepath.ToSlash(m[0])
		}
	}
	return ""
}

// SearchFile searches for 'file' in 'dirs'.
// 'dirs' elements can be plain directory paths or logical paths to .def files (which might be inside zips).
// 'file' is the filename to search (e.g., "kfm.sff").
func SearchFile(file string, dirs []string) string {
	file = strings.Replace(file, "\\", "/", -1)
	if file == "" {
		return ""
	}
	if isZipFull, _, _ := IsZipPath(file); isZipFull {
		if found := FileExist(file); found != "" {
			return found
		}
	}

	if filepath.IsAbs(file) {
		if found := FileExist(file); found != "" {
			return found
		}
	}

	for _, dirCtx := range dirs {
		dirCtx = filepath.ToSlash(dirCtx)
		isZipCtx, zipFileCtx, pathInZipCtx := IsZipPath(dirCtx)

		var candidate string
		if isZipCtx {
			baseDirInZip := filepath.Dir(pathInZipCtx)
			if baseDirInZip == "." {
				baseDirInZip = ""
			}
			candidate = filepath.ToSlash(filepath.Join(zipFileCtx, baseDirInZip, file))
			if found := FileExist(candidate); found != "" {
				return found
			}
		} else {
			candidate = filepath.ToSlash(filepath.Join(filepath.Dir(dirCtx), file))
			if found := FileExist(candidate); found != "" {
				return found
			}
			candidate2 := filepath.ToSlash(filepath.Join(dirCtx, file))
			if found := FileExist(candidate2); found != "" {
				return found
			}
		}
	}

	if found := FileExist(file); found != "" {
		return found
	}
	return file
}

func LoadFile(file *string, dirs []string, load func(string) error) error {
	fp := SearchFile(*file, dirs)
	if err := load(fp); err != nil {
		if len(dirs) > 0 {
			return Error(dirs[0] + ":\n" + fp + "\n" + err.Error())
		}
		return Error(fp + "\n" + err.Error())
	}
	*file = fp
	return nil
}

// Split string on separator, and remove all
// leading and trailing white space from each line
func SplitAndTrim(str, sep string) (ss []string) {
	ss = strings.Split(str, sep)
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}
	return
}

func OldSprintf(f string, a ...interface{}) (s string) {
	iIdx, lIdx, numVerbs := []int{}, []int{}, 0
	for i := 0; i < len(f); i++ {
		if f[i] == '%' {
			i++
			if i >= len(f) {
				break
			}
			for ; i < len(f) && (f[i] == ' ' || f[i] == '0' ||
				f[i] == '-' || f[i] == '+' || f[i] == '#'); i++ {
			}
			if i >= len(f) {
				break
			}
			for ; i < len(f) && f[i] >= '0' && f[i] <= '9'; i++ {
			}
			if i >= len(f) {
				break
			}
			if f[i] == '.' {
				for i++; i < len(f) && f[i] >= '0' && f[i] <= '9'; i++ {
				}
				if i >= len(f) {
					break
				}
			}
			if f[i] == 'h' || f[i] == 'l' || f[i] == 'L' {
				lIdx = append(lIdx, i)
				i++
			}
			if f[i] == '%' {
				continue
			}
			numVerbs++
			if f[i] == 'i' || f[i] == 'u' {
				iIdx = append(iIdx, i)
			}
		}
	}
	if len(iIdx) > 0 || len(lIdx) > 0 {
		b := []byte(f)
		for _, i := range iIdx {
			b[i] = 'd'
		}
		for i := len(lIdx) - 1; i >= 0; i-- {
			b = append(b[:lIdx[i]], b[lIdx[i]+1:]...)
		}
		f = string(b)
	}
	if len(a) > numVerbs {
		a = a[:numVerbs]
	}
	return fmt.Sprintf(f, a...)
}

func SectionName(sec string) (string, string) {
	if len(sec) == 0 || sec[0] != '[' {
		return "", ""
	}
	sec = strings.TrimSpace(strings.SplitN(sec, ";", 2)[0])
	if sec[len(sec)-1] != ']' {
		return "", ""
	}
	sec = sec[1:strings.Index(sec, "]")]
	var name string
	i := strings.Index(sec, " ")
	if i >= 0 {
		name = sec[:i+1]
		sec = sec[i+1:]
	} else {
		name = sec
		sec = ""
	}
	return strings.ToLower(name), sec
}

func HasExtension(file, ext string) bool {
	match, _ := regexp.MatchString(ext, filepath.Ext(strings.ToLower(file)))
	return match
}

func sliceContains(s []string, str string, lower bool) bool {
	if lower {
		strings.ToLower(str)
	}
	for _, v := range s {
		if lower {
			strings.ToLower(v)
		}
		if v == str {
			return true
		}
	}
	return false
}

func sliceInsertInt(array []int, value int, index int) []int {
	return append(array[:index], append([]int{value}, array[index:]...)...)
}

func sliceRemoveInt(array []int, index int) []int {
	return append(array[:index], array[index+1:]...)
}

func sliceMoveInt(array []int, srcIndex int, dstIndex int) []int {
	value := array[srcIndex]
	return sliceInsertInt(sliceRemoveInt(array, srcIndex), value, dstIndex)
}

// We save an array for precise checking, and a float for triggers
func parseIkemenVersion(versionStr string) ([3]uint16, float32) {
	var ver [3]uint16
	parts := SplitAndTrim(versionStr, ".")
	for i, s := range parts {
		if i >= len(ver) {
			break
		}
		if v, err := strconv.ParseUint(s, 10, 16); err == nil {
			ver[i] = uint16(v)
		} else {
			break
		}
	}

	// Convert into a float for triggers
	re := regexp.MustCompile(`[^0-9.]`)
	cleanStr := re.ReplaceAllString(versionStr, "")
	// Keep only the first decimal point
	strParts := strings.Split(cleanStr, ".")
	if len(strParts) > 1 {
		cleanStr = strParts[0] + "." + strings.Join(strParts[1:], "")
	}
	var verF float32
	if result, err := strconv.ParseFloat(cleanStr, 32); err == nil {
		verF = float32(result)
	}

	return ver, verF
}

func parseMugenVersion(versionStr string) ([2]uint16, float32) {
	var ver [2]uint16
	var verF float32

	// Parse the string into the array
	parts := SplitAndTrim(versionStr, ".")
	for i, s := range parts {
		if i >= len(ver) {
			break
		}
		if v, err := strconv.ParseUint(s, 10, 16); err == nil {
			ver[i] = uint16(v)
		} else {
			ver = [2]uint16{}
			break
		}
	}

	// Turn the array into the versions we know
	if ver[0] == 1 && ver[1] == 1 {
		verF = 1.1
	} else if ver[0] == 1 && ver[1] == 0 {
		verF = 1.0
	} else if ver[0] != 0 {
		verF = 0.5 // Arbitrary value
	}

	return ver, verF
}

type Error string

func (e Error) Error() string {
	return string(e)
}

type IniSection map[string]string

func NewIniSection() IniSection {
	return IniSection(make(map[string]string))
}

func ReadIniSection(lines []string, i *int) (
	is IniSection, name string, subname string) {
	for ; *i < len(lines); (*i)++ {
		name, subname = SectionName(lines[*i])
		if len(name) > 0 {
			(*i)++
			break
		}
	}
	if len(name) == 0 {
		return
	}
	is = NewIniSection()
	is.Parse(lines, i)
	return
}

func (is IniSection) Parse(lines []string, i *int) {
	for ; *i < len(lines); (*i)++ {
		if len(lines[*i]) > 0 && lines[*i][0] == '[' {
			break
		}
		line := strings.TrimSpace(strings.SplitN(lines[*i], ";", 2)[0])
		ia := strings.IndexAny(line, "= \t")
		if ia > 0 {
			name := strings.ToLower(line[:ia])
			var data string
			ia = strings.Index(line, "=")
			if ia >= 0 {
				data = strings.TrimSpace(line[ia+1:])
			}
			_, ok := is[name]
			if !ok {
				is[name] = data
			}
		}
	}
}

func (is IniSection) LoadFile(name string, dirs []string,
	load func(string) error) error {
	str := is[name]
	if len(str) == 0 {
		return nil
	}
	return LoadFile(&str, dirs, load)
}

func (is IniSection) ReadI32(name string, out ...*int32) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		if s = strings.TrimSpace(s); len(s) > 0 {
			*out[i] = Atoi(s)
		}
	}
	return true
}

func (is IniSection) ReadF32(name string, out ...*float32) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		if s = strings.TrimSpace(s); len(s) > 0 {
			*out[i] = float32(Atof(s))
		}
	}
	return true
}

func (is IniSection) ReadBool(name string, out ...*bool) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		if s = strings.TrimSpace(s); len(s) > 0 {
			*out[i] = Atoi(s) != 0
		}
	}
	return true
}

func (is IniSection) readI32ForStage(name string, out ...*int32) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		if s = strings.TrimLeftFunc(s, unicode.IsSpace); len(s) > 0 {
			*out[i] = Atoi(s)
		}
		if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
			break
		}
	}
	return true
}

func (is IniSection) readF32ForStage(name string, out ...*float32) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		if s = strings.TrimLeftFunc(s, unicode.IsSpace); len(s) > 0 {
			*out[i] = float32(Atof(s))
		}
		if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
			break
		}
	}
	return true
}

func (is IniSection) readI32CsvForStage(name string) (ary []int32) {
	if str := is[name]; len(str) > 0 {
		for _, s := range strings.Split(str, ",") {
			if s = strings.TrimLeftFunc(s, unicode.IsSpace); len(s) > 0 {
				ary = append(ary, Atoi(s))
			}
			if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
				break
			}
		}
	}
	return
}

func (is IniSection) readF32CsvForStage(name string) (ary []float32) {
	if str := is[name]; len(str) > 0 {
		for _, s := range strings.Split(str, ",") {
			if s = strings.TrimLeftFunc(s, unicode.IsSpace); len(s) > 0 {
				ary = append(ary, float32(Atof(s)))
			}
			if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
				break
			}
		}
	}
	return
}

func (is IniSection) getText(name string) (str string, ok bool, err error) {
	str, ok = is[name]
	if !ok {
		return
	}
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	} else {
		err = Error("String not enclosed in \"")
	}
	return
}

type Layout struct {
	offset  [2]float32
	facing  int8
	vfacing int8
	layerno int16
	scale   [2]float32
	angle   float32
	xshear  float32
	window  [4]int32
}

func newLayout(ln int16) *Layout {
	return &Layout{facing: 1, vfacing: 1, layerno: ln, scale: [...]float32{1, 1}}
}

func ReadLayout(pre string, is IniSection, ln int16) *Layout {
	l := newLayout(ln)
	l.Read(pre, is)
	return l
}

func (l *Layout) Read(pre string, is IniSection) {
	is.ReadF32(pre+"offset", &l.offset[0], &l.offset[1])
	if str := is[pre+"facing"]; len(str) > 0 {
		if Atoi(str) < 0 {
			l.facing = -1
		} else {
			l.facing = 1
		}
	}
	if str := is[pre+"vfacing"]; len(str) > 0 {
		if Atoi(str) < 0 {
			l.vfacing = -1
		} else {
			l.vfacing = 1
		}
	}
	ln := int32(l.layerno)
	is.ReadI32(pre+"layerno", &ln)
	l.layerno = I32ToI16(Min(2, ln))
	is.ReadF32(pre+"scale", &l.scale[0], &l.scale[1])
	is.ReadF32(pre+"angle", &l.angle)
	is.ReadF32(pre+"xshear", &l.xshear)
	if is.ReadI32(pre+"window", &l.window[0], &l.window[1], &l.window[2], &l.window[3]) {
		l.window[0] = int32(float32(l.window[0]) * float32(sys.scrrect[2]/sys.lifebarLocalcoord[0]))
		l.window[1] = int32(float32(l.window[1]) * float32(sys.scrrect[3]/sys.lifebarLocalcoord[1]))
		l.window[2] = int32(float32(l.window[2]) * float32(sys.scrrect[2]/sys.lifebarLocalcoord[0]))
		l.window[3] = int32(float32(l.window[3]) * float32(sys.scrrect[3]/sys.lifebarLocalcoord[1]))
		window := l.window
		if window[2] < window[0] {
			l.window[2] = window[0]
			l.window[0] = window[2]
		}
		if window[3] < window[1] {
			l.window[3] = window[1]
			l.window[1] = window[3]
		}
		l.window[2] -= l.window[0]
		l.window[3] -= l.window[1]
	} else {
		l.window = sys.scrrect
	}
}

func (l *Layout) DrawFaceSprite(x, y float32, ln int16, s *Sprite, fx *PalFX, fscale float32, window *[4]int32) {
	if l.layerno == ln && s != nil {
		// TODO: test "phantom pixel"
		if l.facing < 0 {
			x += sys.lifebar.fnt_scale * sys.lifebarScale
		}
		if l.vfacing < 0 {
			y += sys.lifebar.fnt_scale * sys.lifebarScale
		}
		if s.coldepth <= 8 && s.PalTex == nil {
			s.PalTex = s.CachePalette(s.Pal)
		}
		// Xshear offset correction
		xshear := -l.xshear
		xsoffset := xshear * (float32(s.Offset[1]) * l.scale[1] * fscale)

		s.Draw(x+l.offset[0]*sys.lifebarScale-xsoffset, y+l.offset[1]*sys.lifebarScale,
			l.scale[0]*float32(l.facing)*fscale, l.scale[1]*float32(l.vfacing)*fscale,
			xshear, Rotation{l.angle, 0, 0}, fx, window)
	}
}

func (l *Layout) DrawAnim(r *[4]int32, x, y, scl, xscl, yscl float32, ln int16, a *Animation, palfx *PalFX) {
	// Skip blank animations
	if a == nil || a.isBlank() {
		return
	}
	if l.layerno == ln {
		// TODO: test "phantom pixel"
		if l.facing < 0 {
			x += sys.lifebar.fnt_scale
		}
		if l.vfacing < 0 {
			y += sys.lifebar.fnt_scale
		}
		// Xshear offset correction
		xshear := -l.xshear
		xsoffset := xshear * (float32(a.spr.Offset[1]) * l.scale[1] * scl)

		a.Draw(r, x+l.offset[0]-xsoffset, y+l.offset[1]+float32(sys.gameHeight-240),
			scl, scl, (l.scale[0]*xscl)*float32(l.facing), (l.scale[0]*xscl)*float32(l.facing),
			(l.scale[1]*yscl)*float32(l.vfacing), xshear, Rotation{l.angle, 0, 0},
			float32(sys.gameWidth-320)/2, palfx, 1, [2]float32{1, 1}, 0, 0, 0, false)
	}
}

func (l *Layout) DrawText(x, y, scl float32, ln int16,
	text string, f *Fnt, b, a int32, palfx *PalFX, frgba [4]float32) {
	if l.layerno == ln {
		// TODO: test "phantom pixel"
		if l.facing < 0 {
			x += sys.lifebar.fnt_scale
		}
		if l.vfacing < 0 {
			y += sys.lifebar.fnt_scale
		}
		// Xshear offset correction
		xshear := -l.xshear
		xsoffset := xshear * (float32(f.offset[1]) * l.scale[1] * scl)

		f.Print(text, (x+l.offset[0]-xsoffset)*scl, (y+l.offset[1])*scl,
			l.scale[0]*sys.lifebar.fnt_scale*float32(l.facing)*scl,
			l.scale[1]*sys.lifebar.fnt_scale*float32(l.vfacing)*scl, xshear, Rotation{l.angle, 0, 0},
			b, a, &l.window, palfx, frgba)
	}
}

type AnimLayout struct {
	anim  *Animation
	lay   Layout
	palfx *PalFX
}

func newAnimLayout(sff *Sff, ln int16) AnimLayout {
	return AnimLayout{
		anim:  newAnimation(sff, &sff.palList),
		lay:   *newLayout(ln),
		palfx: newPalFX(),
	}
}

func ReadAnimLayout(pre string, is IniSection, sff *Sff, at AnimationTable, ln int16) AnimLayout {
	al := newAnimLayout(sff, ln)
	al.Read(pre, is, at, ln)
	return al
}

func (al *AnimLayout) Read(pre string, is IniSection, at AnimationTable, ln int16) {
	var g, n int32
	if is.ReadI32(pre+"spr", &g, &n) {
		al.anim.frames = []AnimFrame{*newAnimFrame()}
		al.anim.frames[0].Group, al.anim.frames[0].Number = g, n
		al.anim.mask = 0
		al.lay = *newLayout(ln)
	}
	if is.ReadI32(pre+"anim", &n) {
		if ani := at.get(n); ani != nil {
			al.anim = ani
			al.lay = *newLayout(ln)
		}
	}
	ReadPalFX(pre+"palfx.", is, al.palfx)
	al.lay.Read(pre, is)
}

func (al *AnimLayout) Reset() {
	if al.anim != nil {
		al.anim.Reset()
	}
}

func (al *AnimLayout) Action() {
	if al.palfx != nil {
		al.palfx.step()
	}
	if al.anim != nil {
		al.anim.Action()
	}
}

func (al *AnimLayout) HasAnim() bool {
	return al != nil && al.anim != nil && len(al.anim.frames) > 0
}

func (al *AnimLayout) Draw(x, y float32, layerno int16, scale float32) {
	al.lay.DrawAnim(&al.lay.window, x, y, scale, 1, 1, layerno, al.anim, al.palfx)
}

func ReadPalFX(pre string, is IniSection, pfx *PalFX) int32 {
	pfx.clear()
	pfx.time = -1
	tInit := int32(-1)
	if is.ReadI32(pre+"time", &pfx.time) {
		tInit = pfx.time
	}
	is.ReadI32(pre+"add", &pfx.add[0], &pfx.add[1], &pfx.add[2])
	is.ReadI32(pre+"mul", &pfx.mul[0], &pfx.mul[1], &pfx.mul[2])
	var s [4]int32
	if is.ReadI32(pre+"sinadd", &s[0], &s[1], &s[2], &s[3]) {
		if s[3] < 0 {
			pfx.sinadd[0] = -s[0]
			pfx.sinadd[1] = -s[1]
			pfx.sinadd[2] = -s[2]
			pfx.cycletime[0] = -s[3]
		} else {
			pfx.sinadd[0] = s[0]
			pfx.sinadd[1] = s[1]
			pfx.sinadd[2] = s[2]
			pfx.cycletime[0] = s[3]
		}
	}
	if is.ReadI32(pre+"sinmul", &s[0], &s[1], &s[2], &s[3]) {
		if s[3] < 0 {
			pfx.sinmul[0] = -s[0]
			pfx.sinmul[1] = -s[1]
			pfx.sinmul[2] = -s[2]
			pfx.cycletime[1] = -s[3]
		} else {
			pfx.sinmul[0] = s[0]
			pfx.sinmul[1] = s[1]
			pfx.sinmul[2] = s[2]
			pfx.cycletime[1] = s[3]
		}
	}
	var s2 [2]int32
	if is.ReadI32(pre+"sincolor", &s2[0], &s2[1]) {
		if s2[1] < 0 {
			pfx.sincolor = -s2[0]
			pfx.cycletime[2] = -s2[1]
		} else {
			pfx.sincolor = s2[0]
			pfx.cycletime[2] = s2[1]
		}
	}
	if is.ReadI32(pre+"sinhue", &s2[0], &s2[1]) {
		if s2[1] < 0 {
			pfx.sinhue = -s2[0]
			pfx.cycletime[3] = -s2[1]
		} else {
			pfx.sinhue = s2[0]
			pfx.cycletime[3] = s2[1]
		}
	}
	is.ReadBool(pre+"invertall", &pfx.invertall)
	is.ReadI32(pre+"invertblend", &pfx.invertblend)
	var n float32
	if is.ReadF32(pre+"color", &n) {
		pfx.color = n / 256
	}
	if is.ReadF32(pre+"hue", &n) {
		pfx.hue = n / 256
	}
	return tInit
}

type AnimTextSnd struct {
	snd         [2]int32
	text        LbText
	animLayout  AnimLayout
	displaytime int32
	cnt         int32
}

func newAnimTextSnd(sff *Sff, ln int16) *AnimTextSnd {
	return &AnimTextSnd{
		snd:         [2]int32{-1},
		animLayout:  newAnimLayout(sff, ln),
		displaytime: -2,
	}
}

func ReadAnimTextSnd(pre string, is IniSection,
	sff *Sff, at AnimationTable, ln int16, f []*Fnt) *AnimTextSnd {

	ats := newAnimTextSnd(sff, ln)
	ats.Read(pre, is, at, ln, f)
	return ats
}

func (ats *AnimTextSnd) Read(pre string, is IniSection, at AnimationTable,
	ln int16, f []*Fnt) {
	is.ReadI32(pre+"snd", &ats.snd[0], &ats.snd[1])
	ats.text = *readLbText(pre, is, "", ln, f, 0)
	ats.animLayout.lay = *newLayout(ln)
	ats.animLayout.Read(pre, is, at, ln)
	is.ReadI32(pre+"displaytime", &ats.displaytime)
}

func (ats *AnimTextSnd) Reset() {
	ats.animLayout.Reset()
	ats.cnt = 0
	ats.text.resetTxtPfx()
}

func (ats *AnimTextSnd) Action() {
	ats.animLayout.Action()
	ats.cnt++
	ats.text.step()
}

func (ats *AnimTextSnd) Draw(x, y float32, layerno int16, f []*Fnt, scale float32) {
	if ats.displaytime > 0 && ats.cnt > ats.displaytime {
		return
	}

	// Draw animation if available
	if ats.animLayout.anim != nil && len(ats.animLayout.anim.frames) > 0 {
		ats.animLayout.Draw(x, y, layerno, scale)
		return // Skip text
	}

	// Otherwise draw text
	if ats.text.font[0] >= 0 && int(ats.text.font[0]) < len(f) && len(ats.text.text) > 0 {
		for k, v := range strings.Split(ats.text.text, "\\n") {
			ats.text.lay.DrawText(x, y+
				float32(k)*(float32(f[ats.text.font[0]].Size[1])*ats.text.lay.scale[1]*sys.lifebar.fnt_scale+
					float32(f[ats.text.font[0]].Spacing[1])*ats.text.lay.scale[1]*sys.lifebar.fnt_scale),
				scale, layerno, v, f[ats.text.font[0]], ats.text.font[1], ats.text.font[2], ats.text.palfx,
				ats.text.frgba)
		}
	}
}

func (ats *AnimTextSnd) NoSound() bool {
	return ats.snd[0] < 0
}

// Check if Draw() function is worth calling
func (ats *AnimTextSnd) HasDrawable() bool {
	hasAnim := ats.animLayout.anim != nil && len(ats.animLayout.anim.frames) > 0
	hasText := ats.text.font[0] >= 0 && len(ats.text.text) > 0

	return hasAnim || hasText
}

// In Mugen this returns true if the animation ends before "displaytime" is over
// It seems like the current Ikemen behavior makes more sense however
// https://github.com/ikemen-engine/Ikemen-GO/issues/1150
func (ats *AnimTextSnd) End(dt int32, inf bool) bool {
	anim := ats.animLayout.anim

	// If displaytime is negative, rely on animation current state
	if ats.displaytime < 0 {
		if anim == nil || len(anim.frames) == 0 {
			return true
		}
		if anim.loopend {
			return true
		}
		if inf && anim.curelem == int32(len(anim.frames)-1) && anim.frames[anim.curelem].Time == -1 {
			return true
		}
		return false
	}

	// Otherwise, use displaytime
	return dt >= ats.displaytime
}

// compareNatural compares two strings using natural order
func compareNatural(a, b string) bool {
	re := regexp.MustCompile(`^([a-zA-Z]+)(\d*)$`)

	// Extract prefix and numeric parts for both strings
	aMatches := re.FindStringSubmatch(a)
	bMatches := re.FindStringSubmatch(b)

	// Index range check
	if len(aMatches) < 3 || len(bMatches) < 3 {
		return false
	}

	// Extract prefix and numeric part
	aPrefix, aNumStr := aMatches[1], aMatches[2]
	bPrefix, bNumStr := bMatches[1], bMatches[2]

	// Compare prefixes lexicographically
	if aPrefix != bPrefix {
		return aPrefix < bPrefix
	}

	// Parse numeric part (default to 0 if empty)
	return Atoi(aNumStr) < Atoi(bNumStr)
}

// SortedKeys returns the keys of the map sorted in natural order.
// It works with any map where the key is a string.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	// Sort using natural order
	sort.Slice(keys, func(i, j int) bool {
		return compareNatural(keys[i], keys[j])
	})

	return keys
}

func IsZipPath(path string) (isZip bool, zipFilePath string, pathInZip string) {
	path = filepath.ToSlash(path) // Normalize to forward slashes

	// Try to find ".zip/" as a separator.
	// This handles cases like "C:/path/to/archive.zip/internal/file.txt"
	// or "chars/archive.zip/internal/file.txt"
	idx := strings.Index(strings.ToLower(path), ".zip/")
	if idx != -1 {
		potentialZipPath := path[:idx+4]
		zipFilePath = potentialZipPath
		pathInZip = path[idx+5:]
		isZip = true
		return
	}

	// If no ".zip/" separator, check if the path itself ends with .zip
	if strings.HasSuffix(strings.ToLower(path), ".zip") {
		isZip = true
		zipFilePath = path
		pathInZip = ""
		return
	}

	return false, "", ""
}

type zipMemFileReader struct {
	reader     *bytes.Reader   // Reader for the in-memory content of the file in zip
	zipArchive *zip.ReadCloser // The main zip archive reader
}

func (zmfr *zipMemFileReader) Read(p []byte) (n int, err error) {
	return zmfr.reader.Read(p)
}

func (zmfr *zipMemFileReader) Seek(offset int64, whence int) (int64, error) {
	return zmfr.reader.Seek(offset, whence)
}

func (zmfr *zipMemFileReader) Close() error {
	// The bytes.Reader itself doesn't need closing, but we must close the main zip archive.
	return zmfr.zipArchive.Close()
}

// OpenFile opens a regular file or a file within a zip archive.
// For zip files, it reads the entire entry into memory to ensure full io.Seeker compatibility.
// It returns an io.ReadSeekCloser that must be closed by the caller.
func OpenFile(filename string) (io.ReadSeekCloser, error) {
	filename = filepath.ToSlash(filename)
	isZip, zipFilePath, pathInZip := IsZipPath(filename)
	if isZip {
		zr, err := zip.OpenReader(zipFilePath)
		if err != nil {
			f, err2 := os.Open(filename)
			if err2 != nil {
				return nil, fmt.Errorf("opening zip archive %s: %w", zipFilePath, err)
			}
			return f, nil
		}

		if pathInZip == "" {
			zr.Close() // Close the main archive if we're not reading a specific file from it.
			return nil, fmt.Errorf("path inside zip archive not specified for %s", filename)
		}

		var targetFile *zip.File
		pathInZipLower := strings.ToLower(pathInZip)
		for _, f := range zr.File {
			if strings.ToLower(filepath.ToSlash(f.Name)) == pathInZipLower {
				targetFile = f
				break
			}
		}

		if targetFile == nil {
			zr.Close()
			return nil, fmt.Errorf("file '%s' not found in zip archive '%s'", pathInZip, zipFilePath)
		}

		rc, err := targetFile.Open()
		if err != nil {
			zr.Close()
			return nil, fmt.Errorf("opening file '%s' in zip archive '%s': %w", pathInZip, zipFilePath, err)
		}

		// Read the entire content of the zip file entry into memory
		fileData, err := io.ReadAll(rc)
		rc.Close() // Close the individual file reader from the zip entry
		if err != nil {
			zr.Close() // Close the main zip archive on error
			return nil, fmt.Errorf("reading file '%s' from zip archive '%s': %w", pathInZip, zipFilePath, err)
		}

		// Create a bytes.Reader from the in-memory data
		// *bytes.Reader implements io.ReadSeeker
		bytesReader := bytes.NewReader(fileData)

		// Return our custom wrapper that closes the main zip archive
		return &zipMemFileReader{reader: bytesReader, zipArchive: zr}, nil
	}

	// Not a zip path, open as a normal file
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return f, nil // *os.File implements io.ReadSeekCloser
}

func LowercaseNoExtension(filename string) string {
	basename := filepath.Base(filename)
	ext := filepath.Ext(basename)
	nameOnly := basename
	if len(ext) > 0 {
		nameOnly = basename[0 : len(basename)-len(ext)]
	}
	return strings.ToLower(nameOnly)
}
