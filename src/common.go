package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
)

const (
	IMax           = int32(^uint32(0) >> 1)
	IErr           = ^IMax
	PathDelimiters = "/\\"
)

var randseed int32

func Random() int32 {
	w := randseed / 127773
	randseed = (randseed-w*127773)*16807 - w*2836
	if randseed <= 0 {
		randseed += IMax - Btoi(randseed == 0)
	}
	return randseed
}
func Srand(s int32)             { randseed = s }
func Rand(min, max int32) int32 { return min + Random()/(IMax/(max-min+1)+1) }
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
	if len(arg) > 0 {
		min = arg[0]
		for i := 1; i < len(arg); i++ {
			if arg[i] < min {
				min = arg[i]
			}
		}
	}
	return
}
func Max(arg ...int32) (max int32) {
	if len(arg) > 0 {
		max = arg[0]
		for i := 1; i < len(arg); i++ {
			if arg[i] > max {
				max = arg[i]
			}
		}
	}
	return
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
func IsFinite(f float32) bool {
	return math.Abs(float64(f)) <= math.MaxFloat64
}
func Atoi(str string) int32 {
	n := int32(0)
	str = strings.TrimSpace(str)
	if len(str) >= 0 {
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
			n = n*10 + int32(a[i]-'0')
		}
		if str[0] == '-' {
			n *= -1
		}
	}
	return n
}
func Atof(str string) float64 {
	f := 0.0
	str = strings.TrimSpace(str)
	if len(str) >= 0 {
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
		if p > 0 {
			f *= math.Pow10(p - i)
		}
		if str[0] == '-' {
			f *= -1
		}
	}
	return f
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
func AsciiToString(ascii []byte) string {
	buf := make([]rune, len(ascii))
	for i, a := range ascii {
		buf[i] = rune(a)
	}
	return string(buf)
}
func LoadText(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	if len(bytes) >= 3 &&
		bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf {
		return string(bytes[3:]), nil
	}
	return AsciiToString(bytes), nil
}
func LoadFile(file *string, deffile string, load func(string) error) error {
	var filepath string
	if li := strings.LastIndexAny(deffile, PathDelimiters); li >= 0 {
		filepath = deffile[:li+1] + *file
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			filepath = "data/" + *file
		}
	} else {
		filepath = "data/" + *file
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		filepath = *file
	}
	if err := load(filepath); err != nil {
		return Error(fmt.Sprintf("%s:\n%s", filepath, err.Error()))
	}
	*file = filepath
	return nil
}
func SplitAndTrim(str, sep string) (ss []string) {
	ss = strings.Split(str, sep)
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}
	return
}
func SectionName(sec string) (string, string) {
	if len(sec) == 0 || sec[0] != '[' {
		return "", ""
	}
	sec = strings.TrimSpace(strings.SplitN(sec, ";", 2)[0])
	if sec[len(sec)-1] != ']' {
		return "", ""
	}
	sec = sec[1 : len(sec)-1]
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

type Error string

func (e Error) Error() string { return string(e) }

type IniSection map[string]string

func NewIniSection() IniSection { return IniSection(make(map[string]string)) }
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
			(*i)--
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
func (is IniSection) LoadFile(name, deffile string,
	load func(string) error) error {
	str := is[name]
	if len(str) == 0 {
		return nil
	}
	return LoadFile(&str, deffile, load)
}
func (is IniSection) ReadI32(name string, out ...*int32) {
	str := is[name]
	if len(str) > 0 {
		for i, s := range strings.Split(str, ",") {
			if i >= len(out) {
				break
			}
			*out[i] = Atoi(s)
		}
	}
}
func (is IniSection) ReadF32(name string, out ...*float32) {
	str := is[name]
	if len(str) > 0 {
		for i, s := range strings.Split(str, ",") {
			if i >= len(out) {
				break
			}
			*out[i] = float32(Atof(s))
		}
	}
}

type Layout struct {
	offset      [2]float32
	displaytime int32
	facing      int8
	vfacing     int8
	layerno     int16
	scale       [2]float32
}

func newLayout() *Layout {
	return &Layout{displaytime: -2, facing: 1, vfacing: 1,
		scale: [2]float32{1, 1}}
}
func readLayout(pre string, is IniSection) *Layout {
	l := newLayout()
	is.ReadF32(pre+"offset", &l.offset[0], &l.offset[1])
	is.ReadI32(pre+"displaytime", &l.displaytime)
	if str := is["facing"]; len(str) > 0 {
		if Atoi(str) < 0 {
			l.facing = -1
		} else {
			l.facing = 1
		}
	}
	if str := is["vfacing"]; len(str) > 0 {
		if Atoi(str) < 0 {
			l.vfacing = -1
		} else {
			l.vfacing = 1
		}
	}
	is.ReadF32(pre+"scale", &l.scale[0], &l.scale[1])
	return l
}
func (l *Layout) setup() {
	if l.facing < 0 {
		l.offset[0] += lifebarFontScale
	}
	if l.vfacing < 0 {
		l.offset[1] += lifebarFontScale
	}
}
