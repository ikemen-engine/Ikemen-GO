package main

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode"
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
func Srand(s int32)             { sys.randseed = s }
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
func FileExist(filename string) string {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return filename
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
		return m[0]
	}
	return ""
}
func LoadFile(file *string, deffile string, load func(string) error) error {
	var fp string
	*file = strings.Replace(*file, "\\", "/", -1)
	defdir := filepath.Dir(strings.Replace(deffile, "\\", "/", -1))
	if defdir == "." {
		fp = *file
	} else if defdir == "/" {
		fp = "/" + *file
	} else {
		fp = defdir + "/" + *file
	}
	if fp = FileExist(fp); len(fp) == 0 {
		_else := false
		if defdir != "data" {
			fp = "data/" + *file
			if fp = FileExist(fp); len(fp) == 0 {
				_else = true
			}
		} else {
			_else = true
		}
		if _else {
			fp = *file
			if fp = FileExist(fp); len(fp) == 0 {
				fp = *file
			}
		}
	}
	if err := load(fp); err != nil {
		return Error(fp + "\n" + err.Error())
	}
	*file = fp
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
func (is IniSection) ReadI32(name string, out ...*int32) bool {
	str := is[name]
	if len(str) == 0 {
		return false
	}
	for i, s := range strings.Split(str, ",") {
		if i >= len(out) {
			break
		}
		*out[i] = Atoi(s)
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
		*out[i] = float32(Atof(s))
	}
	return true
}

type Layout struct {
	offset  [2]float32
	facing  int8
	vfacing int8
	layerno int16
	scale   [2]float32
}

func newLayout() *Layout {
	return &Layout{facing: 1, vfacing: 1, scale: [2]float32{1, 1}}
}
func ReadLayout(pre string, is IniSection) *Layout {
	l := newLayout()
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
	var ln int32
	is.ReadI32(pre+"layerno", &ln)
	l.layerno = I32ToI16(Min(2, ln))
	is.ReadF32(pre+"scale", &l.scale[0], &l.scale[1])
	return l
}
func (l *Layout) DrawAnim(r *[4]int32, x, y, scl float32, ln int16,
	a *Animation) {
	if l.layerno == ln {
		if l.facing < 0 {
			x += sys.lifebarFontScale
		}
		if l.vfacing < 0 {
			y += sys.lifebarFontScale
		}
		a.Draw(r, x+l.offset[0], y+l.offset[1]+float32(sys.gameHeight-240),
			scl, scl, l.scale[0]*float32(l.facing), l.scale[0]*float32(l.facing),
			l.scale[1]*float32(l.vfacing),
			0, 0, float32(sys.gameWidth-320)/2, nil, false)
	}
}
func (l *Layout) DrawText(x, y, scl float32, ln int16,
	text string, f *Fnt, b, a int32) {
	if l.layerno == ln {
		if l.facing < 0 {
			x += sys.lifebarFontScale
		}
		if l.vfacing < 0 {
			y += sys.lifebarFontScale
		}
		f.DrawText(text, (x+l.offset[0])*scl, (y+l.offset[1])*scl,
			l.scale[0]*sys.lifebarFontScale*float32(l.facing)*scl,
			l.scale[1]*sys.lifebarFontScale*float32(l.vfacing)*scl, a, b)
	}
}

type AnimLayout struct {
	anim Animation
	lay  Layout
}

func newAnimLayout(sff *Sff) *AnimLayout {
	return &AnimLayout{anim: *newAnimation(sff)}
}
func ReadAnimLayout(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *AnimLayout {
	al := newAnimLayout(sff)
	var g, n int32
	if is.ReadI32(pre+"spr", &g, &n) {
		al.anim.frames = make([]AnimFrame, 1)
		al.anim.frames[0].Group, al.anim.frames[0].Number = I32ToI16(g), I32ToI16(n)
		al.anim.mask = 0
	}
	if is.ReadI32(pre+"anim", &n) {
		ani := at.get(n)
		if ani != nil {
			al.anim = *ani
		}
	}
	al.lay = *ReadLayout(pre, is)
	return al
}
func (al *AnimLayout) Reset() {
	al.anim.Reset()
}
func (al *AnimLayout) Action() {
	al.anim.Action()
}
func (al *AnimLayout) Draw(x, y float32, layerno int16) {
	al.lay.DrawAnim(&sys.scrrect, x, y, 1, layerno, &al.anim)
}

type AnimTextSnd struct {
	snd         [2]int32
	font        [3]int32
	text        string
	anim        AnimLayout
	displaytime int32
}

func newAnimTextSnd(sff *Sff) *AnimTextSnd {
	return &AnimTextSnd{snd: [2]int32{-1}, font: [3]int32{-1}, displaytime: -2}
}
func ReadAnimTextSnd(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *AnimTextSnd {
	ats := newAnimTextSnd(sff)
	is.ReadI32(pre+"snd", &ats.snd[0], &ats.snd[1])
	is.ReadI32(pre+"font", &ats.font[0], &ats.font[1], &ats.font[2])
	ats.text = is[pre+"text"]
	ats.anim = *ReadAnimLayout(pre, is, sff, at)
	is.ReadI32(pre+"displaytime", &ats.displaytime)
	return ats
}
func (ats *AnimTextSnd) Reset()  { ats.anim.Reset() }
func (ats *AnimTextSnd) Action() { ats.anim.Action() }
func (ats *AnimTextSnd) Draw(x, y float32, layerno int16, f []*Fnt) {
	if len(ats.anim.anim.frames) > 0 {
		ats.anim.Draw(x, y, layerno)
	} else if ats.font[0] >= 0 && int(ats.font[0]) < len(f) &&
		len(ats.text) > 0 {
		ats.anim.lay.DrawText(x, y, 1, layerno, ats.text,
			f[ats.font[0]], ats.font[1], ats.font[2])
	}
}
func (ats *AnimTextSnd) NoSound() bool { return ats.snd[0] < 0 }
func (ats *AnimTextSnd) NoDisplay() bool {
	return len(ats.anim.anim.frames) == 0 &&
		(ats.font[0] < 0 || len(ats.text) == 0)
}
func (ats *AnimTextSnd) End(dt int32) bool {
	if ats.displaytime < 0 {
		return len(ats.anim.anim.frames) == 0 || ats.anim.anim.loopend
	}
	return dt >= ats.displaytime
}
