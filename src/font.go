package main

import (
	"encoding/binary"
	"github.com/go-gl/gl/v2.1/gl"
	"os"
	"regexp"
	"strings"
	"unsafe"
)

type FntCharImage struct {
	ofs, w uint16
	img    []Sprite
}
type Fnt struct {
	images    map[rune]*FntCharImage
	palettes  [][256]uint32
	ver, ver2 uint16
	Size      [2]uint16
	Spacing   [2]int32
	colors    int32
	offset    [2]int32
}

func newFnt() *Fnt { return &Fnt{images: make(map[rune]*FntCharImage)} }
func LoadFnt(filename string) (*Fnt, error) {
	f := newFnt()
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { chk(fp.Close()) }()
	buf := make([]byte, 12)
	var n int
	if n, err = fp.Read(buf); err != nil {
		return nil, err
	}
	if string(buf[:n]) != "ElecbyteFnt\x00" {
		return nil, Error("ElecbyteFntではありません")
	}
	read := func(x interface{}) error {
		return binary.Read(fp, binary.LittleEndian, x)
	}
	if err := read(&f.ver); err != nil {
		return nil, err
	}
	if err := read(&f.ver2); err != nil {
		return nil, err
	}
	var pcxDataOffset, pcxDataLenght, txtDataOffset, txtDataLenght uint32
	if err := read(&pcxDataOffset); err != nil {
		return nil, err
	}
	if err := read(&pcxDataLenght); err != nil {
		return nil, err
	}
	if err := read(&txtDataOffset); err != nil {
		return nil, err
	}
	if err := read(&txtDataLenght); err != nil {
		return nil, err
	}
	spr := newSprite()
	if err := spr.readPcxHeader(fp, int64(pcxDataOffset)); err != nil {
		return nil, err
	}
	fp.Seek(int64(pcxDataOffset)+128, 0)
	px := make([]byte, pcxDataLenght-128-768)
	if err := read(px); err != nil {
		return nil, err
	}
	spr.Pal = make([]uint32, 256)
	var rgb [3]byte
	for i := range spr.Pal {
		if err := read(rgb[:]); err != nil {
			return nil, err
		}
		spr.Pal[i] = uint32(rgb[2])<<16 | uint32(rgb[1])<<8 | uint32(rgb[0])
	}
	px = spr.RlePcxDecode(px)
	fp.Seek(int64(txtDataOffset), 0)
	buf = make([]byte, txtDataLenght)
	if err := read(buf); err != nil {
		return nil, err
	}
	lines := SplitAndTrim(string(buf), "\n")
	i := 0
	mapflg, defflg := true, true
	for {
		var name string
		for ; i < len(lines); i++ {
			name, _ = SectionName(lines[i])
			if len(name) > 0 {
				i++
				break
			}
		}
		if len(name) == 0 {
			break
		}
		switch name {
		case "map":
			if mapflg {
				mapflg = false
				re := regexp.MustCompile("(\\S+)(?:\\s+(\\S+)(?:\\s+(\\S+))?)?")
				ofs := uint16(0)
				for ; i < len(lines); i++ {
					if len(lines[i]) > 0 && lines[i][0] == '[' {
						i--
						break
					}
					cap := re.FindStringSubmatch(strings.SplitN(lines[i], ";", 2)[0])
					if len(cap) > 0 {
						var c rune
						if len(cap[1]) >= 2 && cap[1][0] == '0' &&
							(cap[1][1] == 'X' || cap[1][1] == 'x') {
							hex := strings.ToLower(cap[1][2:])
							for _, r := range hex {
								if '0' <= r && r <= '9' {
									c = c<<4 | (r - '0')
								} else if 'a' <= r && r <= 'f' {
									c = c<<4 | (r - 'a' + 10)
								} else {
									break
								}
							}
						} else {
							c = rune(cap[1][0])
						}
						if len(cap[2]) > 0 {
							ofs = I32ToU16(Atoi(cap[2]))
						}
						fci := &FntCharImage{ofs: ofs}
						f.images[c] = fci
						if len(cap[3]) > 0 {
							fci.w = I32ToU16(Atoi(cap[3]))
							ofs += fci.w - f.Size[0]
						} else {
							fci.w = f.Size[0]
						}
					}
					ofs += f.Size[0]
				}
			}
		case "def":
			if defflg {
				defflg = false
				is := NewIniSection()
				is.Parse(lines, &i)
				ary := SplitAndTrim(is["size"], ",")
				if len(ary[0]) > 0 {
					f.Size[0] = I32ToU16(Atoi(ary[0]))
				}
				if len(ary) > 1 && len(ary[1]) > 0 {
					f.Size[1] = I32ToU16(Atoi(ary[1]))
				}
				ary = SplitAndTrim(is["spacing"], ",")
				if len(ary[0]) > 0 {
					f.Spacing[0] = Atoi(ary[0])
				}
				if len(ary) > 1 && len(ary[1]) > 0 {
					f.Spacing[1] = Atoi(ary[1])
				}
				f.colors = Atoi(is["colors"])
				if f.colors > 255 {
					f.colors = 255
				} else if f.colors < 1 {
					f.colors = 1
				}
				ary = SplitAndTrim(is["offset"], ",")
				if len(ary[0]) > 0 {
					f.offset[0] = Atoi(ary[0])
				}
				if len(ary) > 1 && len(ary[1]) > 0 {
					f.offset[1] = Atoi(ary[1])
				}
			}
		}
	}
	f.palettes = make([][256]uint32, 255/f.colors)
	for i := int32(0); int(i) < len(f.palettes); i++ {
		copy(f.palettes[i][:256-f.colors], spr.Pal[:256-f.colors])
		copy(f.palettes[i][256-f.colors:],
			spr.Pal[256-f.colors*(i+1):256-f.colors*i])
	}
	copyCharRect := func(dst []byte, dw int, src []byte, x, w, h int) {
		for i := 0; i < h; i++ {
			copy(dst[dw*i:dw*(i+1)], src[w*i+x:w*i+x+dw])
		}
	}
	for _, fci := range f.images {
		fci.img = make([]Sprite, len(f.palettes))
		for i, p := range f.palettes {
			if i == 0 {
				fci.img[0].shareCopy(spr)
				fci.img[0].Size[0] = fci.w
				px2 := make([]byte, int(fci.w)*int(fci.img[0].Size[1]))
				copyCharRect(px2, int(fci.w), px, int(fci.ofs),
					int(spr.Size[0]), int(spr.Size[1]))
				fci.img[0].SetPxl(px2)
			} else {
				fci.img[i].shareCopy(&fci.img[0])
				fci.img[i].Size[0] = fci.w
			}
			fci.img[i].Offset[0], fci.img[i].Offset[1], fci.img[i].Pal = 0, 0, p[:]
		}
	}
	return f, nil
}
func (f *Fnt) CharWidth(c rune) int32 {
	if c == ' ' {
		return int32(f.Size[0])
	}
	fci := f.images[c]
	if fci == nil {
		return 0
	}
	return int32(fci.w)
}
func (f *Fnt) TextWidth(txt string) (w int32) {
	for _, c := range txt {
		w += f.CharWidth(c) + f.Spacing[0]
	}
	return
}
func (f *Fnt) getCharSpr(c rune, bank int32) *Sprite {
	fci := f.images[c]
	if fci == nil {
		return nil
	}
	return &fci.img[bank]
}
func (f *Fnt) drawChar(x, y, xscl, yscl float32, bank int32, c rune,
	paltex uint32) float32 {
	if c == ' ' {
		return float32(f.Size[0]) * xscl
	}
	spr := f.getCharSpr(c, bank)
	if spr == nil || spr.Tex == nil {
		return 0
	}
	RenderMugenPal(*spr.Tex, paltex, 0, spr.Size, -x*sys.widthScale,
		-y*sys.heightScale, &notiling, xscl*sys.widthScale, xscl*sys.widthScale,
		yscl*sys.heightScale, 1, 0, 0, sys.brightness*255>>8|1<<9, &sys.scrrect,
		0, 0)
	return float32(spr.Size[0]) * xscl
}
func (f *Fnt) DrawText(txt string, x, y, xscl, yscl float32,
	bank, align int32) {
	if len(txt) == 0 {
		return
	}
	x += float32(f.offset[0])*xscl + float32(sys.gameWidth-320)/2
	y += float32(f.offset[1]-int32(f.Size[1])+1)*yscl +
		float32(sys.gameHeight-240)
	if align == 0 {
		x -= float32(f.TextWidth(txt)) * xscl * 0.5
	} else if align < 0 {
		x -= float32(f.TextWidth(txt)) * xscl
	}
	if bank < 0 || len(f.palettes) <= int(bank) {
		bank = 0
	}
	// nil から呼ぶと allPalFX が適用される
	pal := (*PalFX)(nil).getFxPal(f.palettes[bank][:], false)
	gl.Enable(gl.TEXTURE_1D)
	gl.ActiveTexture(gl.TEXTURE1)
	var paltex uint32
	gl.GenTextures(1, &paltex)
	gl.BindTexture(gl.TEXTURE_1D, paltex)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage1D(gl.TEXTURE_1D, 0, gl.RGBA, 256, 0, gl.RGBA, gl.UNSIGNED_BYTE,
		unsafe.Pointer(&pal[0]))
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	for _, c := range txt {
		x += f.drawChar(x, y, xscl, yscl, bank, c, paltex) +
			xscl*float32(f.Spacing[0])
	}
	gl.DeleteTextures(1, &paltex)
	gl.Disable(gl.TEXTURE_1D)
}

type TextSprite struct {
	text             string
	fnt              *Fnt
	bank, align      int32
	x, y, xscl, yscl float32
}

func NewTextSprite() *TextSprite {
	return &TextSprite{align: 1, xscl: 1, yscl: 1}
}
func (ts *TextSprite) Draw() {
	if !sys.frameSkip && ts.fnt != nil {
		ts.fnt.DrawText(ts.text, ts.x, ts.y, ts.xscl, ts.yscl, ts.bank, ts.align)
	}
}
