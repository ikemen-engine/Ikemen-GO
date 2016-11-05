package main

// #cgo pkg-config: libpng
// #include <png.h>
import "C"
import (
	"encoding/binary"
	"github.com/go-gl/gl/v2.1/gl"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"runtime"
	"strings"
	"unsafe"
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

type Texture uint32

func textureFinalizer(t *Texture) {
	if *t != 0 {
		gl.DeleteTextures(1, (*uint32)(t))
	}
}
func NewTexture() (t *Texture) {
	t = new(Texture)
	runtime.SetFinalizer(t, textureFinalizer)
	return
}
func (t *Texture) FromImageRGBA(rgba *image.RGBA) {
	gl.BindTexture(gl.TEXTURE_2D, uint32(*t))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()),
		0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&rgba.Pix[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP)
}
func (t *Texture) FromImage(im image.Image) {
	switch trueim := im.(type) {
	case *image.RGBA:
		t.FromImageRGBA(trueim)
	default:
		copy := image.NewRGBA(trueim.Bounds())
		draw.Draw(copy, trueim.Bounds(), trueim, image.Pt(0, 0), draw.Src)
		t.FromImageRGBA(copy)
	}
}

type PalleteList struct {
	palletes   [][]uint32
	palleteMap []int
	PalTable   map[[2]int16]int
}

func (pl *PalleteList) Clear() {
	pl.palletes = nil
	pl.palleteMap = nil
	pl.PalTable = make(map[[2]int16]int)
}
func (pl *PalleteList) SetSource(i int, p []uint32) {
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
func (pl *PalleteList) NewPal() (i int, p []uint32) {
	i = len(pl.palletes)
	p = make([]uint32, 256)
	pl.SetSource(i, p)
	return
}
func (pl *PalleteList) Get(i int) []uint32 {
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

func (sh *SffHeader) Read(r io.Reader, lofs *uint32, tofs *uint32) error {
	buf := make([]byte, 12)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != "ElecbyteSpr\x00" {
		return Error("ElecbyteSprではありません")
	}
	read := func(x interface{}) error {
		return binary.Read(r, binary.LittleEndian, x)
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

type Sprite struct {
	Pal           []uint32
	Tex           *Texture
	Group, Number int16
	Size          [2]uint16
	Offset        [2]int16
	palidx, link  int
	rle           int
}

func NewSprite() *Sprite {
	return &Sprite{palidx: -1, link: -1}
}
func (s *Sprite) shareCopy(src *Sprite) {
	s.Pal = src.Pal
	s.Tex = src.Tex
	s.Size = src.Size
	s.palidx = src.palidx
}
func (s *Sprite) GetPal(pl *PalleteList) []uint32 {
	if s.Pal != nil || s.rle == -12 {
		return s.Pal
	}
	return pl.Get(int(s.palidx))
}
func (s *Sprite) SetPxl(px []byte) {
	if int64(len(px)) != int64(s.Size[0])*int64(s.Size[1]) {
		return
	}
	s.Tex = NewTexture()
	gl.BindTexture(gl.TEXTURE_2D, uint32(*s.Tex))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE,
		int32(s.Size[0]), int32(s.Size[1]),
		0, gl.LUMINANCE, gl.UNSIGNED_BYTE, unsafe.Pointer(&px[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP)
}
func (s *Sprite) readHeader(r io.Reader, ofs *uint32, size *uint32,
	link *uint16) error {
	read := func(x interface{}) error {
		return binary.Read(r, binary.LittleEndian, x)
	}
	if err := read(ofs); err != nil {
		return err
	}
	if err := read(size); err != nil {
		return err
	}
	if err := read(s.Offset[:]); err != nil {
		return err
	}
	if err := read(&s.Group); err != nil {
		return err
	}
	if err := read(&s.Number); err != nil {
		return err
	}
	if err := read(link); err != nil {
		return err
	}
	return nil
}
func (s *Sprite) readPcxHeader(f *os.File, offset int64) error {
	f.Seek(offset, 0)
	read := func(x interface{}) error {
		return binary.Read(f, binary.LittleEndian, x)
	}
	var dummy uint16
	if err := read(&dummy); err != nil {
		return err
	}
	var encoding, bpp byte
	if err := read(&encoding); err != nil {
		return err
	}
	if err := read(&bpp); err != nil {
		return err
	}
	if bpp != 8 {
		return Error("256色でありません")
	}
	var rect [4]uint16
	if err := read(rect[:]); err != nil {
		return err
	}
	f.Seek(offset+66, 0)
	var bpl uint16
	if err := read(&bpl); err != nil {
		return err
	}
	s.Size[0] = rect[2] - rect[0] + 1
	s.Size[1] = rect[3] - rect[1] + 1
	if encoding == 1 {
		s.rle = int(bpl)
	} else {
		s.rle = 0
	}
	return nil
}
func (s *Sprite) RlePcxDecode(rle []byte) (p []byte) {
	if len(rle) == 0 || s.rle <= 0 {
		return rle
	}
	p = make([]byte, int(s.Size[0])*int(s.Size[1]))
	i, j, k, w := 0, 0, 0, int(s.Size[0])
	for j < len(p) {
		n, d := 1, rle[i]
		if i < len(rle)-1 {
			i++
		}
		if d >= 0xc0 {
			n = int(d & 0x3f)
			d = rle[i]
			if i < len(rle)-1 {
				i++
			}
		}
		for ; n > 0; n-- {
			if k < w && j < len(p) {
				p[j] = d
				j++
			}
			k++
			if k == s.rle {
				k = 0
				n = 1
			}
		}
	}
	s.rle = 0
	return
}
func (s *Sprite) read(f *os.File, sh *SffHeader, offset int64,
	datasize uint32, nextSubheader uint32, prev *Sprite,
	palletSame *bool, pl *PalleteList, c00 bool) error {
	if int64(nextSubheader) > offset {
		// 最後以外datasizeを無視
		datasize = nextSubheader - uint32(offset)
	}
	read := func(x interface{}) error {
		return binary.Read(f, binary.LittleEndian, x)
	}
	var ps byte
	if err := read(&ps); err != nil {
		return err
	}
	*palletSame = ps != 0 && prev != nil
	if err := s.readPcxHeader(f, offset); err != nil {
		return err
	}
	f.Seek(offset+128, 0)
	var palSize int
	if c00 || *palletSame {
		palSize = 0
	} else {
		palSize = 768
	}
	px := make([]byte, int(datasize)-(128+palSize))
	if err := read(px); err != nil {
		return err
	}
	if *palletSame {
		if prev != nil {
			s.palidx = prev.palidx
		}
		if s.palidx < 0 {
			s.palidx, _ = pl.NewPal()
		}
	} else {
		var pal []uint32
		s.palidx, pal = pl.NewPal()
		if c00 {
			f.Seek(offset+int64(datasize)-768, 0)
		}
		var rgb [3]byte
		for i := range pal {
			if err := read(rgb[:]); err != nil {
				return err
			}
			pal[i] = uint32(rgb[0])<<16 | uint32(rgb[1])<<16 | uint32(rgb[2])
		}
	}
	s.SetPxl(s.RlePcxDecode(px))
	return nil
}
func (s *Sprite) readHeaderV2(r io.Reader, ofs *uint32, size *uint32,
	lofs uint32, tofs uint32, link *uint16) error {
	read := func(x interface{}) error {
		return binary.Read(r, binary.LittleEndian, x)
	}
	if err := read(&s.Group); err != nil {
		return err
	}
	if err := read(&s.Number); err != nil {
		return err
	}
	if err := read(s.Size[:]); err != nil {
		return err
	}
	if err := read(s.Offset[:]); err != nil {
		return err
	}
	if err := read(link); err != nil {
		return err
	}
	var format byte
	if err := read(&format); err != nil {
		return err
	}
	s.rle = -int(format)
	var dummy byte
	if err := read(&dummy); err != nil {
		return err
	}
	if err := read(ofs); err != nil {
		return err
	}
	if err := read(size); err != nil {
		return err
	}
	var tmp uint16
	if err := read(&tmp); err != nil {
		return err
	}
	s.palidx = int(tmp)
	if err := read(&tmp); err != nil {
		return err
	}
	if tmp&1 == 0 {
		*ofs += lofs
	} else {
		*ofs += tofs
	}
	return nil
}
func (s *Sprite) Rle8Decode(rle []byte) (p []byte) {
	if len(rle) == 0 {
		return rle
	}
	p = make([]byte, int(s.Size[0])*int(s.Size[1]))
	i, j := 0, 0
	for j < len(p) {
		n, d := 1, rle[i]
		if i < len(rle)-1 {
			i++
		}
		if d&0xc0 == 0x40 {
			n = int(d & 0x3f)
			d = rle[i]
			if i < len(rle)-1 {
				i++
			}
		}
		for ; n > 0; n-- {
			if j < len(p) {
				p[j] = d
				j++
			}
		}
	}
	return
}
func (s *Sprite) Rle5Decode(rle []byte) (p []byte) {
	if len(rle) == 0 {
		return rle
	}
	p = make([]byte, int(s.Size[0])*int(s.Size[1]))
	i, j := 0, 0
	for j < len(p) {
		rl := int(rle[i])
		if i < len(rle)-1 {
			i++
		}
		dl := int(rle[i] & 0x7f)
		c := byte(0)
		if rle[i]>>7 != 0 {
			if i < len(rle)-1 {
				i++
			}
			c = rle[i]
		}
		if i < len(rle)-1 {
			i++
		}
		for {
			if j < len(p) {
				p[j] = c
				j++
			}
			rl--
			if rl < 0 {
				dl--
				if dl < 0 {
					break
				}
				c = rle[i] & 0x1f
				rl = int(rle[i] >> 5)
				if i < len(rle)-1 {
					i++
				}
			}
		}
	}
	return
}
func (s *Sprite) Lz5Decode(rle []byte) (p []byte) {
	if len(rle) == 0 {
		return rle
	}
	p = make([]byte, int(s.Size[0])*int(s.Size[1]))
	i, j, n := 0, 0, 0
	ct, cts, rb, rbc := rle[i], uint(0), byte(0), uint(0)
	if i < len(rle)-1 {
		i++
	}
	for j < len(p) {
		d := int(rle[i])
		if i < len(rle)-1 {
			i++
		}
		if ct&byte(1<<cts) != 0 {
			if d&0x3f == 0 {
				d = (d<<2 | int(rle[i])) + 1
				if i < len(rle)-1 {
					i++
				}
				n = int(rle[i]) + 2
				if i < len(rle)-1 {
					i++
				}
			} else {
				rb |= byte(d & 0xc0 >> rbc)
				rbc += 2
				n = int(d & 0x3f)
				if rbc < 8 {
					d = int(rle[i]) + 1
					if i < len(rle)-1 {
						i++
					}
				} else {
					d = int(rb) + 1
					rb, rbc = 0, 0
				}
			}
			for {
				if j < len(p) {
					p[j] = p[j-d]
					j++
				}
				n--
				if n < 0 {
					break
				}
			}
		} else {
			if d&0xe0 == 0 {
				n = int(rle[i]) + 8
				if i < len(rle)-1 {
					i++
				}
			} else {
				n = d >> 5
				d &= 0x1f
			}
			for ; n > 0; n-- {
				if j < len(p) {
					p[j] = byte(d)
					j++
				}
			}
		}
		cts++
		if cts >= 8 {
			ct, cts = rle[i], 0
			if i < len(rle)-1 {
				i++
			}
		}
	}
	return
}
func (s *Sprite) readV2(f *os.File, sh *SffHeader, offset int64,
	datasize uint32) error {
	f.Seek(offset+4, 0)
	if s.rle < 0 {
		format := -s.rle
		var px []byte
		if 2 <= format && format <= 4 {
			px = make([]byte, datasize)
			if err := binary.Read(f, binary.LittleEndian, px); err != nil {
				return err
			}
		}
		switch format {
		case 2:
			px = s.Rle8Decode(px)
		case 3:
			px = s.Rle5Decode(px)
		case 4:
			px = s.Lz5Decode(px)
		case 10:
		case 11, 12:
			s.rle = -12
			img, err := png.Decode(f)
			if err != nil {
				return err
			}
			s.Tex = NewTexture()
			s.Tex.FromImage(img)
			return nil
		default:
			return Error("不明な形式です")
		}
		s.SetPxl(px)
	}
	return nil
}
