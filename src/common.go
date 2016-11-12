package main

import (
	"math"
	"strings"
)

func Abs(f float32) float32 {
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
func AsciiToString(ascii []byte) string {
	buf := make([]rune, len(ascii))
	for i, a := range ascii {
		buf[i] = rune(a)
	}
	return string(buf)
}
func SplitAndTrim(str, sep string) (ss []string) {
	ss = strings.Split(str, sep)
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}
	return
}
func SectionName(sec *string) string {
	if len(*sec) == 0 || (*sec)[0] != '[' {
		return ""
	}
	*sec = strings.TrimSpace(strings.SplitN(*sec, ";", 2)[0])
	if (*sec)[len(*sec)-1] != ']' {
		return ""
	}
	*sec = (*sec)[1 : len(*sec)-1]
	var name string
	i := strings.Index(*sec, " ")
	if i >= 0 {
		name = (*sec)[:i]
		*sec = (*sec)[i+1:]
	} else {
		name = *sec
		*sec = ""
	}
	return strings.ToLower(name)
}

type Error string

func (e Error) Error() string { return string(e) }

type IniSection map[string]string

func NewIniSection() IniSection { return IniSection(make(map[string]string)) }
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
func (is IniSection) Benri(prefix, name string) string {
	return is[strings.Join([]string{prefix, name}, "")]
}
