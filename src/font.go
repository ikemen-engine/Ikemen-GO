package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

const MaxFontBatchSize = 250

type FontRenderer interface {
	Init(renderer interface{})
	LoadFont(file string, scale int32, windowWidth int, windowHeight int) (interface{}, error)
	//LoadTrueTypeFont(program uint32, r io.Reader, scale int32, low, high rune, dir Direction) (Font, error)
	//newProgram(GLSLVersion uint, vertexShaderSource, fragmentShaderSource string) (uint32, error)
}

type Font interface {
	SetColor(red float32, green float32, blue float32, alpha float32)
	UpdateResolution(windowWidth int, windowHeight int)
	Printf(x, y float32, scale float32, spacingXAdd float32, align int32, blend bool, window [4]int32, fs string, argv ...interface{}) error
	Width(scale float32, spacingXAdd float32, fs string, argv ...interface{}) float32
}

type character struct {
	textureID uint32 // ID handle of the glyph texture
	uv        [4]float32
	width     int //glyph width
	height    int //glyph height
	advance   int //glyph advance
	bearingH  int //glyph bearing horizontal
	bearingV  int //glyph bearing vertical
}

type color struct {
	r float32
	g float32
	b float32
	a float32
}

// Direction represents the direction in which strings should be rendered.
type Direction uint8

// Known directions.
const (
	LeftToRight Direction = iota // E.g.: Latin
	RightToLeft                  // E.g.: Arabic
	TopToBottom                  // E.g.: Chinese
)

// FntCharImage stores sprite and position
type FntCharImage struct {
	ofs, w uint16
	img    []Sprite
}

// TtfFont implements TTF font rendering on supported platforms
type TtfFont interface {
	SetColor(red float32, green float32, blue float32, alpha float32)
	Width(scale float32, spacingXAdd float32, fs string, argv ...interface{}) float32
	Printf(x, y float32, scale float32, spacingXAdd float32, align int32, blend bool, window [4]int32, fs string, argv ...interface{}) error
	UpdateResolution(windowWidth int, windowHeight int)
}

// Fnt is a interface for basic font information
type Fnt struct {
	images      map[int32]map[rune]*FntCharImage
	palettes    [][256]uint32
	coldepth    []byte
	ver, ver2   uint16
	Type        string
	BankType    string
	Size        [2]uint16
	Spacing     [2]int32
	colors      int32
	offset      [2]int32
	ttf         TtfFont
	paltex      Texture
	lastPalBank int32
	lastPalBase *uint32
	paltexCache map[*uint32]Texture
}

func newFnt() *Fnt {
	return &Fnt{
		images:      make(map[int32]map[rune]*FntCharImage),
		BankType:    "palette",
		lastPalBank: -1,
		paltexCache: make(map[*uint32]Texture),
	}
}

func loadFnt(filename string, height int32) (*Fnt, error) {
	if HasExtension(filename, ".fnt") {
		return loadFntV1(filename)
	}

	return loadFntV2(filename, height)
}

func loadFntV1(filename string) (*Fnt, error) {
	f := newFnt()
	f.images[0] = make(map[rune]*FntCharImage)

	fp, err := os.Open(filename)

	if err != nil {
		return nil, Error("File not found")
	}

	defer func() { chk(fp.Close()) }()

	// Read header
	buf := make([]byte, 12)
	n, err := fp.Read(buf)

	// Error reading file
	if err != nil {
		return nil, err
	}

	// Error is not a valid fnt file
	if string(buf[:n]) != "ElecbyteFnt\x00" {
		return nil, Error("Unrecognized FNT file: " + string(buf[:n]))
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

	var pcxDataOffset, pcxDataLength, txtDataOffset, txtDataLength uint32
	if err := read(&pcxDataOffset); err != nil {
		return nil, err
	}

	if err := read(&pcxDataLength); err != nil {
		return nil, err
	}

	if err := read(&txtDataOffset); err != nil {
		return nil, err
	}

	if err := read(&txtDataLength); err != nil {
		return nil, err
	}

	spr := newSprite()
	if err := spr.readPcxHeader(fp, int64(pcxDataOffset)); err != nil {
		return nil, err
	}

	fp.Seek(int64(pcxDataOffset)+128, 0)
	px := make([]byte, pcxDataLength-128-768)
	if err := read(px); err != nil {
		return nil, err
	}

	spr.Pal = make([]uint32, 256)
	var rgb [3]byte
	for i := range spr.Pal {
		if err := read(rgb[:]); err != nil {
			return nil, err
		}
		var alpha byte = 255
		if i == 0 {
			alpha = 0
		}
		spr.Pal[i] = uint32(alpha)<<24 | uint32(rgb[2])<<16 | uint32(rgb[1])<<8 | uint32(rgb[0])
	}

	px = spr.RlePcxDecode(px)
	fp.Seek(int64(txtDataOffset), 0)
	buf = make([]byte, txtDataLength)
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
				re := regexp.MustCompile(`(\S+)(?:\s+(\S+)(?:\s+(\S+))?)?`)
				ofs := uint16(0)
				w := int32(0)
				for ; i < len(lines); i++ {
					if len(lines[i]) > 0 && lines[i][0] == '[' {
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
						f.images[0][c] = fci
						if len(cap[3]) > 0 {
							w = Atoi(cap[3])
							if w < 0 {
								ofs += I32ToU16(int32(ofs) - w)
								w = 0 - w
							}
							fci.w = I32ToU16(w)
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
				loadDefInfo(f, filename, is, 0)
			}
		}
	}
	c := Min(255, int32(math.Ceil(float64(f.colors)/16))*16)
	f.palettes = make([][256]uint32, 255/c)
	for i := int32(0); int(i) < len(f.palettes); i++ {
		copy(f.palettes[i][:256-c], spr.Pal[:256-c])
		copy(f.palettes[i][256-c:], spr.Pal[256-c*(i+1):256-c*i])
	}
	copyCharRect := func(dst []byte, dw int, src []byte, x, w, h int) {
		dw2 := dw
		if x+dw > w {
			dw2 = w - x
		}
		if dw2 > 0 {
			for i := 0; i < h; i++ {
				copy(dst[dw*i:dw*i+dw2], src[w*i+x:w*i+x+dw2])
			}
		}
	}
	for _, fci := range f.images[0] {
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
				i, fci := i, fci
				sys.mainThreadTask <- func() {
					fci.img[i].shareCopy(&fci.img[0])
					fci.img[i].Size[0] = fci.w
				}
			}
			fci.img[i].Offset[0], fci.img[i].Offset[1], fci.img[i].Pal = 0, 0, p[:]
		}
	}
	return f, nil
}

func loadFntV2(filename string, height int32) (*Fnt, error) {
	f := newFnt()

	content, err := LoadText(filename)

	if err != nil {
		return nil, Error("File not found")
	}

	lines := SplitAndTrim(string(content), "\n")
	i := 0
	var name string

	for ; i < len(lines); i++ {
		name, _ = SectionName(lines[i])
		if len(name) > 0 {
			is := NewIniSection()
			i++
			is.Parse(lines, &i)
			i--
			switch name {
			case "def":
				loadDefInfo(f, filename, is, height)
			}
		}
	}
	return f, nil
}

func loadDefInfo(f *Fnt, filename string, is IniSection, height int32) {
	f.Type = strings.ToLower(is["type"])
	if _, ok := is["banktype"]; ok {
		f.BankType = strings.ToLower(is["banktype"])
	}
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
	f.colors = Clamp(Atoi(is["colors"]), 1, 255)
	ary = SplitAndTrim(is["offset"], ",")
	if len(ary[0]) > 0 {
		f.offset[0] = Atoi(ary[0])
	}
	if len(ary) > 1 && len(ary[1]) > 0 {
		f.offset[1] = Atoi(ary[1])
	}

	if len(is["file"]) > 0 {
		if f.Type == "truetype" {
			LoadFntTtf(f, filename, is["file"], height)
		} else {
			LoadFntSff(f, filename, is["file"])
		}
	}
}

func LoadFntSff(f *Fnt, fontfile string, filename string) {
	fileDir := SearchFile(filename, []string{fontfile, "font/", sys.motif.Def, "", "data/"})
	sff, err := loadSff(fileDir, false, false, false)

	if err != nil {
		panic(err)
	}

	// Load sprites
	var pal_default []uint32
	for k, sprite := range sff.sprites {
		s := sff.getOwnPalSprite(sprite.Group, sprite.Number, &sff.palList)
		if sprite.Group == 0 || f.BankType == "sprite" {
			if f.images[int32(sprite.Group)] == nil {
				f.images[int32(sprite.Group)] = make(map[rune]*FntCharImage)
			}
			if pal_default == nil && sff.header.Version[0] == 1 {
				pal_default = s.Pal
			}
			offsetX := uint16(s.Offset[0])
			sizeX := uint16(s.Size[0])

			fci := &FntCharImage{
				ofs: offsetX,
				w:   sizeX,
			}
			fci.img = make([]Sprite, 1)
			fci.img[0] = *s
			f.images[int32(sprite.Group)][rune(k[1])] = fci
		}
	}

	// Load palettes
	f.palettes = make([][256]uint32, sff.header.NumberOfPalettes)
	f.coldepth = make([]byte, sff.header.NumberOfPalettes)
	var idef int
	for i := 0; i < int(sff.header.NumberOfPalettes); i++ {
		var pal []uint32
		si, ok := sff.palList.PalTable[[...]uint16{0, uint16(i)}]
		if ok && si >= 0 {
			pal = sff.palList.Get(si)
			if i == 0 {
				idef = si
			}
			switch sff.palList.numcols[[...]uint16{0, uint16(i)}] {
			case 256:
				f.coldepth[i] = 8
			case 32:
				f.coldepth[i] = 5
			}
		} else {
			pal = sff.palList.Get(idef)
		}
		copy(f.palettes[i][:], pal)
	}
	if len(f.palettes) == 0 && pal_default != nil {
		f.palettes = make([][256]uint32, 1)
		copy(f.palettes[0][:], pal_default)
	}
}

// CharWidth returns the width that has a specified character
func (f *Fnt) CharWidth(c rune, bt int32) int32 {
	if c == ' ' {
		return int32(f.Size[0])
	}
	fci := f.images[bt][c]
	if fci == nil {
		return 0
	}
	return int32(fci.w)
}

// TextWidth returns the width that has a specified text.
// This depends on each char's width and font spacing
func (f *Fnt) TextWidth(txt string, bank int32, spacingXAdd int32) (w int32) {
	if len(txt) == 0 {
		return 0
	}
	// TTF: ask renderer for width in one pass (and include spacingXAdd).
	if f.Type == "truetype" {
		if f.ttf == nil {
			return 0
		}
		return int32(math.Round(float64(f.ttf.Width(1, float32(spacingXAdd), "%s", txt))))
	}
	if f.BankType != "sprite" {
		bank = 0
	}
	// Sprite fonts: avoid byte-index "last rune" bug by tracking rune index.
	runeCount := utf8.RuneCountInString(txt)
	ri := 0
	for _, c := range txt {
		cw := f.CharWidth(c, bank)
		// in mugen negative spacing matching char width seems to skip calc,
		// even for 1 symbol string (which normally shouldn't use spacing)
		if cw+f.Spacing[0]+spacingXAdd > 0 {
			w += cw
			if ri < runeCount-1 {
				w += f.Spacing[0] + spacingXAdd
			}
		}
		ri++
	}
	return
}

func (f *Fnt) getCharSpr(c rune, bank, bt int32) *Sprite {
	fci := f.images[bt][c]
	if fci == nil {
		return nil
	}

	if bank < int32(len(fci.img)) {
		return &fci.img[bank]
	}

	return &fci.img[0]
}

func (f *Fnt) drawChar(
	x, y,
	xscl, yscl float32,
	bank, bt int32,
	c rune, pal []uint32,
	rp RenderParams,
) float32 {
	if c == ' ' {
		return float32(f.Size[0]) * xscl
	}

	spr := f.getCharSpr(c, bank, bt)
	if spr == nil || spr.Tex == nil {
		return 0
	}

	// Only paletted sprites (<=8bpp) use palette mapping.
	if spr.coldepth <= 8 {
		// In case of mismatched color depth between bank palette and the sprite's own palette,
		// Mugen 1.1 uses the latter, ignoring the bank
		if len(f.palettes) != 0 && len(f.coldepth) > int(bank) &&
			f.images[bt][c].img[0].coldepth != 32 &&
			f.coldepth[bank] != f.images[bt][c].img[0].coldepth {
			pal = f.images[bt][c].img[0].Pal[:] //palfx.getFxPal(f.images[bt][c].img[0].Pal[:], false)
		}
	}

	x -= xscl * float32(spr.Offset[0])
	y -= yscl * float32(spr.Offset[1])
	if spr.coldepth <= 8 {
		// Cache palette textures per palette, not per sprite
		var base *uint32
		if len(pal) > 0 {
			base = &pal[0]
		}
		if base == nil {
			f.paltex = nil
			f.lastPalBase = nil
		} else if f.lastPalBase == base && f.paltex != nil {
			// fast path
		} else if tx, ok := f.paltexCache[base]; ok && tx != nil {
			f.paltex = tx
			f.lastPalBase = base
		} else {
			// first time seeing this palette: upload once and reuse
			f.paltex = NewTextureFromPalette(pal)
			f.paltexCache[base] = f.paltex
			f.lastPalBase = base
		}
	}

	// Update only the render parameters that change between each character
	rp.tex = spr.Tex
	if spr.coldepth <= 8 {
		rp.paltex = f.paltex
	} else {
		// RGBA / truecolor glyphs must NOT use paltex.
		// Leaving paltex non-nil here makes the renderer think the glyph is indexed.
		rp.paltex = nil
	}
	rp.size = spr.Size
	rp.x = -x * sys.widthScale
	rp.y = -y * sys.heightScale

	RenderSprite(rp)
	return float32(spr.Size[0]) * xscl
}

func (f *Fnt) Print(txt string, x, y, xscl, yscl, rxadd float32, rot Rotation, projectionMode int32, fLength float32, bank, align int32,
	window *[4]int32, palfx *PalFX, frgba [4]float32) {
	if !sys.frameSkip {
		if f.Type == "truetype" {
			f.DrawTtf(txt, x, y, xscl, yscl, align, true, window, frgba, 0)
		} else {
			f.DrawText(txt, x, y, xscl, yscl, rxadd, rot, projectionMode, fLength, bank, align, window, palfx, frgba[3], 0)
		}
	}
}

// DrawText prints on screen a specified text with the current font sprites
func (f *Fnt) DrawText(txt string, x, y, xscl, yscl, rxadd float32,
	rot Rotation, projectionMode int32, fLength float32, bank, align int32, window *[4]int32, palfx *PalFX, alpha float32, spacingXAdd int32) {

	if len(txt) == 0 || xscl == 0 || yscl == 0 {
		return
	}

	var bt int32
	if f.BankType == "sprite" {
		bt = bank
		bank = 0
	} else if bank < 0 || len(f.palettes) <= int(bank) {
		bank = 0
	}

	// not existing characters treated as space
	for i, c := range txt {
		if c != ' ' && f.images[bt][c] == nil {
			//txt = strings.Replace(txt, string(c), " ", -1)
			txt = txt[:i] + string(' ') + txt[i+1:]
		}
	}

	x += float32(f.offset[0])*xscl + float32(sys.gameWidth-320)/2
	y += float32(f.offset[1]-int32(f.Size[1])+1)*yscl + float32(sys.gameHeight-240)

	var rcx, rcy float32

	if rot.IsZero() {
		if xscl < 0 {
			x *= -1
		}
		if yscl < 0 {
			y *= -1
		}
		rcx, rcy = rcx*sys.widthScale, 0
	} else {
		rcx, rcy = (x+rcx)*sys.widthScale, y*sys.heightScale
		x, y = AbsF(xscl)*float32(f.offset[0]), AbsF(yscl)*float32(f.offset[1])
	}

	if align == 0 {
		x -= float32(f.TextWidth(txt, bank, spacingXAdd)) * xscl * 0.5
	} else if align < 0 {
		x -= float32(f.TextWidth(txt, bank, spacingXAdd)) * xscl
	}

	var pal []uint32
	if len(f.palettes) != 0 {
		pal = f.palettes[bank][:] //palfx.getFxPal(f.palettes[bank][:], false)
	}

	// Only force a new paltex on bank change; otherwise reuse the previous one
	if f.lastPalBank != bank {
		f.paltex = nil
		f.lastPalBank = bank
		f.lastPalBase = nil
	}

	// Set the trans type
	tt := TT_none
	if alpha < 1.0 {
		tt = TT_add
	}

	alphaVal := int32(255 * sys.brightness * alpha)

	// Initialize common render parameters
	rp := RenderParams{
		tex:            nil,
		paltex:         nil,
		size:           [2]uint16{0, 0},
		x:              0,
		y:              0,
		tile:           notiling,
		xts:            xscl * sys.widthScale,
		xbs:            xscl * sys.widthScale,
		ys:             yscl * sys.heightScale,
		vs:             1,
		rxadd:          rxadd,
		xas:            1,
		yas:            1,
		rot:            rot,
		tint:           0,
		blendMode:      tt,
		blendAlpha:     [2]int32{alphaVal, 255 - alphaVal},
		mask:           0,
		pfx:            palfx,
		window:         window,
		rcx:            rcx,
		rcy:            rcy,
		projectionMode: projectionMode,
		fLength:        fLength,
		xOffset:        0,
		yOffset:        0,
	}

	for _, c := range txt {
		x += f.drawChar(x, y, xscl, yscl, bank, bt, c, pal, rp) + xscl*float32(f.Spacing[0]+spacingXAdd)
	}
}

func (f *Fnt) DrawTtf(txt string, x, y, xscl, yscl float32, align int32,
	blend bool, window *[4]int32, frgba [4]float32, spacingXAdd float32) {

	if len(txt) == 0 {
		return
	}

	if f.ttf != nil {
		f.ttf.UpdateResolution(int(sys.gameWidth), int(sys.gameHeight))
	}

	x += float32(f.offset[0])*xscl + float32(sys.gameWidth-320)/2
	//y += float32(f.offset[1]-int32(f.Size[1])+1)*yscl + float32(sys.gameHeight-240)

	win := [4]int32{(*window)[0], sys.scrrect[3] - ((*window)[1] + (*window)[3]),
		(*window)[2], (*window)[3]}

	f.ttf.SetColor(frgba[0], frgba[1], frgba[2], frgba[3])
	f.ttf.Printf(x, y, (xscl+yscl)/2, spacingXAdd, align, blend, win, "%s", txt) //x, y, scale, spacingXAdd, align, blend, window, string, printf args
}

type TextSprite struct {
	ownerid          int32
	id               int32
	text, textInit   string
	template         string
	params           []interface{}
	fnt              *Fnt
	bank, align      int32
	x, y, xscl, yscl float32
	window           [4]int32
	xshear           float32
	rot              Rotation
	projection       int32
	fLength          float32
	xvel, yvel       float32
	localScale       float32
	offsetX          int32
	layerno          int16
	palfx            *PalFX
	frgba            [4]float32 // ttf fonts
	forcecolor       bool
	removetime       int32 // text sctrl
	elapsedTicks     float32
	textSpacing      [2]float32
	textDelay        float32
	textWrap         bool
	friction         [2]float32
	accel            [2]float32
	vel              [2]float32
	maxDist          [2]float32
	// initial, unscaled values
	offsetInit   [2]float32
	scaleInit    [2]float32
	windowInit   [4]float32
	velocityInit [2]float32
}

func NewTextSprite() *TextSprite {
	ts := &TextSprite{}
	ts.loadDefaults()
	return ts
}

// More like a reset but that name is taken
func (ts *TextSprite) Clear() {
	ts.loadDefaults()
}

func (ts *TextSprite) loadDefaults() {
	pfx := ts.palfx
	prm := ts.params[:0]

	*ts = TextSprite{
		id:         -1,
		align:      1,
		xscl:       1,
		yscl:       1,
		window:     sys.scrrect,
		frgba:      [...]float32{1.0, 1.0, 1.0, 1.0},
		removetime: 1,
		localScale: 1,
		friction:   [2]float32{1.0, 1.0},
		scaleInit:  [2]float32{1.0, 1.0},
	}

	ts.params = prm

	if pfx == nil {
		ts.palfx = newPalFX()
	} else {
		ts.palfx = pfx
		ts.palfx.clear()
	}
	ts.palfx.setColor(255, 255, 255)
}

// Creates a shallow copy with independent palette mapping
func (ts *TextSprite) Copy() *TextSprite {
	if ts == nil {
		return nil
	}
	// Shallow copy all value fields and pointers first.
	nt := &TextSprite{}
	*nt = *ts
	// Deep-copy PalFX so instances do not share effect state.
	if ts.palfx != nil {
		pf := newPalFX()
		*pf = *ts.palfx
		nt.palfx = pf
	}
	// sharing ts.fnt is intentional/harmless.
	return nt
}

func (ts *TextSprite) SetLocalcoord(lx, ly float32) {
	if lx <= 0 || ly <= 0 {
		return
	}
	v := lx
	if lx*3 > ly*4 {
		v = ly * 4 / 3
	}
	ts.localScale = float32(v / 320)
	ts.offsetX = -int32(math.Floor(float64(lx)/(float64(v)/320)-320) / 2)
}

func (ts *TextSprite) SetPos(x, y float32) {
	ts.offsetInit[0] = x
	ts.offsetInit[1] = y
	ts.x = x/ts.localScale + float32(ts.offsetX)
	ts.y = y / ts.localScale
}

func (ts *TextSprite) AddPos(x, y float32) {
	ts.x += x / ts.localScale
	ts.y += y / ts.localScale
}

func (ts *TextSprite) SetScale(xscl, yscl float32) {
	ts.scaleInit[0] = xscl
	ts.scaleInit[1] = yscl
	ts.xscl = xscl / ts.localScale
	ts.yscl = yscl / ts.localScale
}

func (ts *TextSprite) SetWindow(window [4]float32) {
	if window == [4]float32{0, 0, 0, 0} {
		return
	}
	ts.windowInit = window
	x := window[0]/ts.localScale + float32(ts.offsetX)
	y := window[1] / ts.localScale
	w := (window[2] - window[0]) / ts.localScale
	h := (window[3] - window[1]) / ts.localScale
	ts.window[0] = int32((x + float32(sys.gameWidth-320)/2) * sys.widthScale)
	// TODO: test if this truetype adjustment is needed
	//ts.window[1] = int32((y + float32(sys.gameHeight-240)) * sys.heightScale)
	// Keep scissor Y consistent with the respective draw paths:
	//  - Sprite fonts (DrawText) add +(sys.gameHeight-240) to Y
	//  - TTF fonts (DrawTtf) do NOT add that offset
	if ts.fnt != nil && ts.fnt.Type == "truetype" {
		ts.window[1] = int32(y * sys.heightScale)
	} else {
		ts.window[1] = int32((y + float32(sys.gameHeight-240)) * sys.heightScale)
	}
	ts.window[2] = int32(w*sys.widthScale + 0.5)
	ts.window[3] = int32(h*sys.heightScale + 0.5)
}

func (ts *TextSprite) SetColor(r, g, b, a int32) {
	ts.forcecolor = true
	ts.palfx.setColor(r, g, b)
	ts.frgba = [...]float32{float32(r) / 255, float32(g) / 255,
		float32(b) / 255, float32(a) / 255}
}

func (ts *TextSprite) SetTextSpacing(xs, ys float32) {
	ts.textSpacing = [2]float32{xs, ys} // TODO: / ts.localScale?
}

func (ts *TextSprite) SetVelocity(xvel, yvel float32) {
	ts.velocityInit[0] = xvel
	ts.velocityInit[1] = yvel
	ts.xvel = xvel / ts.localScale
	ts.yvel = yvel / ts.localScale
	ts.vel = [2]float32{}
}

func (ts *TextSprite) SetMaxDist(x, y float32) {
	ts.maxDist[0] = x / ts.localScale
	ts.maxDist[1] = y / ts.localScale
}

func (ts *TextSprite) SetAccel(xacc, yacc float32) {
	ts.accel[0] = xacc / ts.localScale
	ts.accel[1] = yacc / ts.localScale
}

func (ts *TextSprite) IsFullyTyped() bool {
	if ts.textDelay <= 0 {
		return true // There's no partial logic
	}
	if int32(len(ts.text)) <= int32(ts.elapsedTicks/ts.textDelay) {
		return true
	}
	return false
}

func (ts *TextSprite) TextWidth(txt string) int32 {
	if ts.fnt == nil {
		return 0
	}
	spacingXAdd := int32(math.Round(float64(ts.textSpacing[0])))
	return ts.fnt.TextWidth(txt, ts.bank, spacingXAdd)
}

func (ts *TextSprite) getLineLength(windowWrap bool) int32 {
	// Compute the available line length in the same coordinate space that TextWidth() use ("local 320" space).
	if ts.fnt == nil {
		return 0
	}

	// Base "full screen" width in text space.
	lcWidth := float32(sys.motif.Info.Localcoord[0])
	if ts.localScale > 0 {
		lcWidth = lcWidth / ts.localScale
	}
	// Window boundaries expressed in text space.
	var left, right float32
	if windowWrap && ts.windowInit != [4]float32{0, 0, 0, 0} && ts.localScale > 0 {
		// windowInit is in motif localcoords; convert to text space.
		left = ts.windowInit[0]/ts.localScale + float32(ts.offsetX)
		right = ts.windowInit[2]/ts.localScale + float32(ts.offsetX)
	} else {
		// No explicit window: use the whole localcoord width.
		left = float32(ts.offsetX)
		right = left + lcWidth
	}

	switch ts.align {
	case 1: // left
		return int32(math.Round(float64(right - ts.x)))
	case 0: // center
		leftCap := float64(ts.x - left)
		rightCap := float64(right - ts.x)
		return int32(math.Round(math.Min(leftCap, rightCap) * 2))
	default: // right
		return int32(math.Round(float64(ts.x - left)))
	}
}

func (ts *TextSprite) decodeEscapes(text string) string {
	return strings.ReplaceAll(text, "\\n", "\n")
}

// splitNewlines breaks lines into words and inserts explicit "\n" tokens.
func (ts *TextSprite) splitNewlines(text string) []string {
	text = ts.decodeEscapes(text)
	parts := strings.Split(text, "\n")
	var result []string
	for i, p := range parts {
		tokens := strings.Fields(p)
		if len(tokens) > 0 {
			result = append(result, tokens...)
		}
		if i < len(parts)-1 {
			// Insert a newline token to indicate forced line break
			result = append(result, "\n")
		}
	}
	return result
}

// wrapText wraps the fullLine text based on the typedLen and other parameters
func (ts *TextSprite) wrapText(fullLine string, typedLen int) {
	// If typedLen <= 0, nothing to display
	if ts.fnt == nil || ts.window[2] <= 0 {
		return
	}

	// If typing hasn't started yet, make sure nothing is shown.
	if typedLen <= 0 {
		ts.text = ""
		return
	}

	// Return byte index of the first 'n' runes in s.
	toByteIndex := func(s string, n int) int {
		if n <= 0 {
			return 0
		}
		for idx := range s {
			if n == 0 {
				return idx
			}
			n--
		}
		// If we consumed exactly the number of runes or ran out, return len(s)
		if n == 0 {
			return len(s)
		}
		return len(s)
	}

	// No wrapping: cut to typedLen, omit a trailing '\' before a pending "\n"
	if !ts.textWrap {
		// 'typedLen' is in RUNES; convert to byte index safely.
		endByte := toByteIndex(fullLine, typedLen)
		if endByte > len(fullLine) {
			endByte = len(fullLine)
		}
		// Avoid showing a dangling '\' if the next byte is 'n'
		if endByte > 0 && endByte < len(fullLine) &&
			fullLine[endByte-1] == '\\' && fullLine[endByte] == 'n' {
			endByte--
		}
		ts.text = fullLine[:endByte]
		return
	}

	// Split the text into words, preserving newlines
	words := ts.splitNewlines(fullLine)

	var result strings.Builder
	var currentLine strings.Builder
	var currentLineWidth int32
	if false {
		fmt.Println(currentLineWidth) // dummy due to compiler error
	}

	remainingChars := typedLen

	// Determine the maximum line length based on alignment and window
	lineLen := ts.getLineLength(true)

	spacingXAdd := int32(math.Round(float64(ts.textSpacing[0])))

	for _, w := range words {
		if w == "\n" {
			// Forced newline: commit currentLine to result and start a new line
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLineWidth = 0
			continue
		}

		// Determine the prefix (space) if not the first word in the line
		spacePrefix := ""
		if currentLine.Len() > 0 {
			spacePrefix = " "
		}

		// Candidate word with prefix
		candidate := spacePrefix + w

		// Measure the width if we add this word to the current line
		tentativeLine := currentLine.String() + candidate
		tentativeWidth := int32(math.Round(float64(ts.fnt.TextWidth(tentativeLine, ts.bank, spacingXAdd)) * float64(ts.xscl)))

		if tentativeWidth > lineLen && currentLine.Len() > 0 {
			// Current line is full; commit it and start a new line with the current word
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentLineWidth = 0
			// Recalculate candidate without space prefix since it's a new line
			candidate = w
			tentativeLine = candidate
			tentativeWidth = int32(math.Round(float64(ts.fnt.TextWidth(tentativeLine, ts.bank, spacingXAdd)) * float64(ts.xscl)))
			if tentativeWidth > lineLen {
				// Optionally handle long words by splitting them
				// For simplicity, we'll commit the long word to a new line as is
				currentLine.WriteString(candidate)
				currentLineWidth = tentativeWidth
				continue
			}
		}

		// Now, add the candidate to the current line
		currentLine.WriteString(candidate)
		currentLineWidth = tentativeWidth

		// Count runes in this word (spaces ignored); final rune limit is enforced later.
		charsToAdd := utf8.RuneCountInString(w)
		if remainingChars < charsToAdd {
			charsToAdd = remainingChars
		}

		// Update remainingChars
		remainingChars -= charsToAdd

		// If no more characters to add, break
		if remainingChars <= 0 {
			break
		}
	}

	// After processing all words, add any remaining text in currentLine to result
	if currentLine.Len() > 0 {
		result.WriteString(currentLine.String())
	}

	// Final trimming: ensure we keep ONLY 'typedLen' RUNES across lines.
	{
		finalResult := ""
		runesAdded := 0
		lines := strings.Split(result.String(), "\n")
		for i, line := range lines {
			if runesAdded >= typedLen {
				break
			}
			lineRunes := 0
			for range line {
				lineRunes++
			}
			if runesAdded+lineRunes <= typedLen {
				finalResult += line
				runesAdded += lineRunes
				if i < len(lines)-1 {
					finalResult += "\n"
				}
			} else {
				remaining := typedLen - runesAdded
				endByte := toByteIndex(line, remaining)
				if endByte > len(line) {
					endByte = len(line)
				}
				finalResult += line[:endByte]
				runesAdded = typedLen
				break
			}
		}

		// Remove the trailing newline if present
		finalResult = strings.TrimRight(finalResult, "\n")
		ts.text = finalResult
		return
	}
}

// StepTypewriter advances a "typewriter" cursor (typedLen / charDelayCounter / lineFullyRendered)
// for a given line of text. It does not touch ts.text directly; callers are expected to feed the
// resulting typedLen into wrapText() or their own rendering logic.
func StepTypewriter(fullLine string, typedLen *int, charDelayCounter *int32,
	lineFullyRendered *bool, delay float32) {

	// Empty line: trivially "done".
	if fullLine == "" {
		*typedLen = 0
		*lineFullyRendered = true
		*charDelayCounter = 0
		return
	}

	totalRunes := utf8.RuneCountInString(fullLine)

	// textdelay <= 0 means "show everything immediately"
	if delay <= 0 {
		*typedLen = totalRunes
		*charDelayCounter = 0
		*lineFullyRendered = true
		return
	}

	if *typedLen > totalRunes {
		*typedLen = totalRunes
	}
	if *lineFullyRendered {
		return
	}
	if *charDelayCounter <= 0 {
		*typedLen++
		d := int32(delay)
		if d <= 0 {
			d = 1
		}
		*charDelayCounter = d - 1
	} else {
		*charDelayCounter--
	}
	if *typedLen >= totalRunes {
		*typedLen = totalRunes
		*lineFullyRendered = true
	}
}

func (ts *TextSprite) updateVel() {
	// candidate new displacement
	nx := ts.vel[0] + ts.xvel
	ny := ts.vel[1] + ts.yvel

	// clamp to maxDist per axis, if set (non-zero)
	if ts.maxDist[0] != 0 {
		lim := ts.maxDist[0]
		if (lim > 0 && nx >= lim) || (lim < 0 && nx <= lim) {
			nx = lim
			ts.xvel = 0
		}
	}
	if ts.maxDist[1] != 0 {
		lim := ts.maxDist[1]
		if (lim > 0 && ny >= lim) || (lim < 0 && ny <= lim) {
			ny = lim
			ts.yvel = 0
		}
	}

	ts.vel[0] = nx
	ts.vel[1] = ny

	// apply friction/accel only while we're within the maxDist on that axis
	if ts.maxDist[0] == 0 ||
		math.Abs(float64(ts.vel[0])) < math.Abs(float64(ts.maxDist[0])) {
		ts.xvel *= ts.friction[0]
		ts.xvel += ts.accel[0]
		if math.Abs(float64(ts.xvel)) < 0.1 && math.Abs(float64(ts.friction[0])) < 1 {
			ts.xvel = 0
		}
	} else {
		ts.xvel = 0
	}

	if ts.maxDist[1] == 0 ||
		math.Abs(float64(ts.vel[1])) < math.Abs(float64(ts.maxDist[1])) {
		ts.yvel *= ts.friction[1]
		ts.yvel += ts.accel[1]
		if math.Abs(float64(ts.yvel)) < 0.1 && math.Abs(float64(ts.friction[1])) < 1 {
			ts.yvel = 0
		}
	} else {
		ts.yvel = 0
	}
}

func (ts *TextSprite) Update() {
	ts.elapsedTicks++
	ts.updateVel()
	if ts.palfx != nil && !ts.forcecolor {
		ts.palfx.step()
	}
}

func (ts *TextSprite) Draw(ln int16) {
	if sys.frameSkip || ts.layerno != ln || ts.fnt == nil || len(ts.text) == 0 {
		return
	}

	// Replace each tab with 4 spaces
	// We do this first so that length checks are accurate
	text := strings.ReplaceAll(ts.text, "\t", "    ")
	text = ts.decodeEscapes(text)

	maxChars := int32(len(text))

	// If textDelay is greater than 0, it controls the maximum number of characters
	if ts.textDelay > 0 {
		// Offset the delay so that we show the first character immediately
		elapsed := ts.elapsedTicks + ts.textDelay
		maxChars = int32(elapsed / ts.textDelay)
	}

	maxChars = Clamp(maxChars, 0, int32(len(text)))
	// Control of total displayed characters
	totalCharsShown := 0

	// "phantom pixel" adjustment to match mugen flipping behavior (extra pixel)
	var phantomX float32
	if ts.align == -1 {
		phantomX = 1 / ts.localScale
	}

	lines := strings.Split(text, "\n")

	spacingXAdd := int32(math.Round(float64(ts.textSpacing[0])))

	for i, line := range lines {
		lineLength := len(line)

		// Shows the characters progressively
		charsToShow := int(Min(int32(lineLength), maxChars-int32(totalCharsShown)))
		if charsToShow <= 0 {
			continue
		}

		newY := ts.y + float32(i)*((float32(ts.fnt.Size[1])+float32(ts.fnt.Spacing[1])+ts.textSpacing[1])*ts.yscl)

		// Xshear offset correction
		xshear := -ts.xshear
		xsoffset := xshear * (float32(ts.fnt.offset[1]) * ts.yscl)

		// Draw the visible line
		if ts.fnt.Type == "truetype" {
			ts.fnt.DrawTtf(line[:charsToShow], ts.x+ts.vel[0]+phantomX, newY+ts.vel[1], ts.xscl, ts.yscl, ts.align, true, &ts.window, ts.frgba, float32(spacingXAdd))
		} else {
			ts.fnt.DrawText(line[:charsToShow], ts.x+ts.vel[0]-xsoffset+phantomX, newY+ts.vel[1], ts.xscl, ts.yscl,
				xshear, ts.rot, ts.projection, ts.fLength, ts.bank, ts.align, &ts.window, ts.palfx, ts.frgba[3], spacingXAdd)
		}

		totalCharsShown += charsToShow
		if totalCharsShown >= int(maxChars) {
			break
		}
	}
}

func (ts *TextSprite) Reset() {
	ts.SetPos(ts.offsetInit[0], ts.offsetInit[1])
	ts.SetScale(ts.scaleInit[0], ts.scaleInit[1])
	ts.SetWindow(ts.windowInit)
	ts.SetVelocity(ts.velocityInit[0], ts.velocityInit[1])
	ts.text = ts.textInit
	if ts.palfx != nil {
		ts.palfx.clear()
	}
	ts.elapsedTicks = 0
}
