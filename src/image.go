package main

import (
	"encoding/binary"
	"github.com/go-gl/gl/v2.1/gl"
	"image/color"
	"os"
	"strings"
)

func gltest() {
	vertShader := strings.Join([]string{
		"void main(void){",
		"gl_TexCoord[0] = gl_TextureMatrix[0] * gl_MultiTexCoord0;",
		"gl_Position = ftransform();",
		"}\x00"}, "")
	fragShader := strings.Join([]string{
		"uniform float a;",
		"uniform sampler2D tex;",
		"uniform sampler1D pal;",
		"uniform int msk;",
		"void main(void){",
		"float r = texture2D(tex, gl_TexCoord[0].st).r;",
		"vec4 c;",
		"gl_FragColor =",
		"int(255.0*r) == msk ? vec4(0.0)",
		": (c = texture1D(pal, r*0.9961), vec4(c.b, c.g, c.r, a));",
		"}\x00"}, "")
	fragShaderFc := strings.Join([]string{
		"uniform float a;",
		"uniform sampler2D tex;",
		"uniform bool neg;",
		"uniform float gray;",
		"uniform vec3 add;",
		"uniform vec3 mul;",
		"void main(void){",
		"vec4 c = texture2D(tex, gl_TexCoord[0].st);",
		"if(neg) c.rgb = vec3(1.0) - c.rgb;",
		"float gcol = (c.r + c.g + c.b) / 3.0;",
		"c.r += (gcol - c.r) * gray + add.r;",
		"c.g += (gcol - c.g) * gray + add.g;",
		"c.b += (gcol - c.b) * gray + add.b;",
		"c.rgb *= mul;",
		"c.a *= a;",
		"gl_FragColor = c;",
		"}\x00"}, "")
	fragShaderFcS := strings.Join([]string{
		"uniform float a;",
		"uniform sampler2D tex;",
		"uniform vec3 color;",
		"void main(void){",
		"vec4 c = texture2D(tex, gl_TexCoord[0].st);",
		"c.rgb = color * c.a;",
		"c.a *= a;",
		"gl_FragColor = c;",
		"}\x00"}, "")
	errLog := func(obl uintptr) error {
		var size int32
		gl.GetObjectParameterivARB(obl, gl.INFO_LOG_LENGTH, &size)
		if size <= 0 {
			return nil
		}
		var l int32
		str := make([]byte, size+1)
		gl.GetInfoLogARB(obl, size, &l, &str[0])
		return Error(str[:l])
	}
	compile := func(shaderType uint32, src string) (shader uintptr) {
		shader = gl.CreateShaderObjectARB(shaderType)
		s, l := gl.Str(src), int32(len(src)-1)
		gl.ShaderSourceARB(shader, 1, &s, &l)
		gl.CompileShaderARB(shader)
		var ok int32
		gl.GetObjectParameterivARB(shader, gl.OBJECT_COMPILE_STATUS_ARB, &ok)
		if ok == 0 {
			chk(errLog(shader))
			panic(Error("コンパイルエラー"))
		}
		return
	}
	link := func(v uintptr, f uintptr) (program uintptr) {
		program = gl.CreateProgramObjectARB()
		gl.AttachObjectARB(program, v)
		gl.AttachObjectARB(program, f)
		gl.LinkProgramARB(program)
		var ok int32
		gl.GetObjectParameterivARB(program, gl.OBJECT_LINK_STATUS_ARB, &ok)
		if ok == 0 {
			chk(errLog(program))
			panic(Error("リンクエラー"))
		}
		return
	}
	vertObj := compile(gl.VERTEX_SHADER, vertShader)
	fragObj := compile(gl.FRAGMENT_SHADER, fragShader)
	shader := link(vertObj, fragObj)
	gl.GetUniformLocationARB(shader, gl.Str("pal\x00"))
	gl.GetUniformLocationARB(shader, gl.Str("msk\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFc)
	shaderFc := link(vertObj, fragObj)
	gl.GetUniformLocationARB(shaderFc, gl.Str("neg\x00"))
	gl.GetUniformLocationARB(shaderFc, gl.Str("gray\x00"))
	gl.GetUniformLocationARB(shaderFc, gl.Str("add\x00"))
	gl.GetUniformLocationARB(shaderFc, gl.Str("mul\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFcS)
	shaderFcS := link(vertObj, fragObj)
	gl.GetUniformLocationARB(shaderFcS, gl.Str("color\x00"))
	gl.DeleteObjectARB(fragObj)
	gl.DeleteObjectARB(vertObj)
}

type PalleteList struct {
	palletes   [][]color.Color
	palleteMap []int
	PalTable   map[[2]int16]int
}

func (pl *PalleteList) Clear() {
	pl.palletes = nil
	pl.palleteMap = nil
	pl.PalTable = make(map[[2]int16]int)
}
func (pl *PalleteList) SetSource(i int, p []color.Color) {
	if i < len(pl.palleteMap) {
		pl.palleteMap[i] = i
	} else {
		for i > len(pl.palleteMap) {
			AppendI(&pl.palleteMap, len(pl.palleteMap))
		}
		AppendI(&pl.palleteMap, i)
	}
	if i < len(pl.palletes) {
		pl.palletes[i] = p
	} else {
		for i > len(pl.palletes) {
			AppendPal(&pl.palletes, nil)
		}
		AppendPal(&pl.palletes, p)
	}
}
func (pl *PalleteList) NewPal() (i int, p []color.Color) {
	i = len(pl.palletes)
	p = make([]color.Color, 256)
	pl.SetSource(i, p)
	return
}
func (pl *PalleteList) Get(i int) []color.Color {
	return pl.palletes[pl.palleteMap[i]]
}
func (pl *PalleteList) Remap(source int, destination int) {
	pl.palleteMap[source] = destination
}
func (pl *PalleteList) ResetRemap() {
	for i := range pl.palleteMap {
		pl.palleteMap[i] = i
	}
}
func (pl *PalleteList) GetPalMap() []int {
	pm := make([]int, len(pl.palleteMap))
	copy(pm, pl.palleteMap)
	return pm
}
func (pl *PalleteList) SwapPalMap(palMap *[]int) bool {
	if len(*palMap) != len(pl.palleteMap) {
		return false
	}
	*palMap, pl.palleteMap = pl.palleteMap, *palMap
	return true
}

type SffHeader struct {
	Ver0, Ver1, Ver2, Ver3   byte
	FirstSpriteHeaderOffset  uint32
	FirstPaletteHeaderOffset uint32
	NumberOfSprites          uint32
	NumberOfPalettes         uint32
}

func (sh *SffHeader) Read(f *os.File, lofs *uint32, tofs *uint32) error {
	buf := make([]byte, 12)
	n, err := f.Read(buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != "ElecbyteSpr\x00" {
		return Error("ElecbyteSprではありません")
	}
	read := func(x interface{}) error {
		return binary.Read(f, binary.LittleEndian, x)
	}
	if err := read(&sh.Ver3); err != nil {
		return err
	}
	if err := read(&sh.Ver2); err != nil {
		return err
	}
	if err := read(&sh.Ver1); err != nil {
		return err
	}
	if err := read(&sh.Ver0); err != nil {
		return err
	}
	var dummy uint32
	if err := read(&dummy); err != nil {
		return err
	}
	switch sh.Ver0 {
	case 1:
		sh.FirstPaletteHeaderOffset, sh.NumberOfPalettes = 0, 0
		if err := read(&sh.NumberOfSprites); err != nil {
			return err
		}
		if err := read(&sh.FirstSpriteHeaderOffset); err != nil {
			return err
		}
		if err := read(&dummy); err != nil {
			return err
		}
	case 2:
		for i := 0; i < 4; i++ {
			if err := read(&dummy); err != nil {
				return err
			}
		}
		if err := read(&sh.FirstSpriteHeaderOffset); err != nil {
			return err
		}
		if err := read(&sh.NumberOfSprites); err != nil {
			return err
		}
		if err := read(&sh.FirstPaletteHeaderOffset); err != nil {
			return err
		}
		if err := read(&sh.NumberOfPalettes); err != nil {
			return err
		}
		if err := read(lofs); err != nil {
			return err
		}
		if err := read(&dummy); err != nil {
			return err
		}
		if err := read(tofs); err != nil {
			return err
		}
	default:
		return Error("バージョンが不正です")
	}
	return nil
}
