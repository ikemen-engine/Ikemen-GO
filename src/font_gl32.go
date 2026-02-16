//go:build !android

package main

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"os"

	gl "github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Font_GL32 struct {
	fontChar     map[rune]*character
	ttf          *truetype.Font
	scale        int32
	windowWidth  int
	windowHeight int
	textures     []*TextureAtlas
	color        color
}

type FontRenderer_GL32 struct {
	shaderProgram *ShaderProgram_GL32
	vao           uint32
	vbo           uint32
}

func (r *FontRenderer_GL32) Init(renderer interface{}) {
	// Configure the default font vertex and fragment shaders
	r.newProgram(150, vertexFontShader, fragmentFontShader)

	// Register attributes
	r.shaderProgram.RegisterAttributes("vert", "vertTexCoord")

	// Configure VAO/VBO for texture quads
	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	// Pre-allocate for maximum batch size
	gl.BufferData(gl.ARRAY_BUFFER, MaxFontBatchSize*6*4*4, nil, gl.DYNAMIC_DRAW)

	// Configure attributes
	vLoc := uint32(r.shaderProgram.attributes["vert"])
	gl.EnableVertexAttribArray(vLoc)
	gl.VertexAttribPointer(vLoc, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	tLoc := uint32(r.shaderProgram.attributes["vertTexCoord"])
	gl.EnableVertexAttribArray(tLoc)
	gl.VertexAttribPointer(tLoc, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	// Clean up binding state, but not attribute state
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// LoadFont loads the specified font at the given scale.
func (r *FontRenderer_GL32) LoadFont(file string, scale int32, windowWidth int, windowHeight int) (interface{}, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	// Activate corresponding render state
	// gl.UseProgram(r.shaderProgram.program)

	f, err := r.LoadTrueTypeFont(fd, scale, 32, 127, LeftToRight)
	if err != nil {
		return nil, err
	}
	//set screen resolution
	f.windowWidth = windowWidth
	f.windowHeight = windowHeight
	return f, nil
}

// SetColor allows you to set the text color to be used when you draw the text
func (f *Font_GL32) SetColor(red float32, green float32, blue float32, alpha float32) {
	f.color.r = red
	f.color.g = green
	f.color.b = blue
	f.color.a = alpha
}

func (f *Font_GL32) UpdateResolution(windowWidth int, windowHeight int) {
	f.windowWidth = windowWidth
	f.windowHeight = windowHeight
}

func (r *FontRenderer_GL32) SetFontPipeline() {
	mr := gfx.(*Renderer_GL32)

	// Do nothing if we were already using the font shader
	if mr.program == r.shaderProgram.program {
		return
	}

	mr.ChangeProgram(r.shaderProgram.program)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	vLoc := uint32(r.shaderProgram.attributes["vert"])
	gl.EnableVertexAttribArray(vLoc)
	gl.VertexAttribPointer(vLoc, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	tLoc := uint32(r.shaderProgram.attributes["vertTexCoord"])
	gl.EnableVertexAttribArray(tLoc)
	gl.VertexAttribPointer(tLoc, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
}

// Printf draws a string to the screen, takes a list of arguments like printf
func (f *Font_GL32) Printf(x, y float32, scale float32, spacingXAdd float32,
	align int32, blend bool, window [4]int32, fs string, argv ...interface{}) error {

	indices := []rune(fmt.Sprintf(fs, argv...))
	r := gfx.(*Renderer_GL32)
	fr := gfxFont.(*FontRenderer_GL32)

	if len(indices) == 0 {
		return nil
	}

	// Activate corresponding render state
	fr.SetFontPipeline()
	program := fr.shaderProgram

	// Buffer to store vertex data for multiple glyphs
	batchSize := Min(MaxFontBatchSize, int32(len(indices)))
	batchVertices := make([]float32, 0, batchSize*6*4)

	//setup blending mode
	r.SetBlending(blend, BlendAdd, BlendSrcAlpha, BlendOneMinusSrcAlpha)

	//restrict drawing to a certain part of the window
	r.EnableScissor(window[0], window[1], window[2], window[3])

	// Set texture location
	r.SetUniformISub(program.uniforms["tex"], 0)

	//set text color
	r.SetUniformFSub(program.uniforms["textColor"], f.color.r, f.color.g, f.color.b, f.color.a)

	//set screen resolution
	r.SetUniformFSub(program.uniforms["resolution"], float32(f.windowWidth), float32(f.windowHeight))

	gl.ActiveTexture(gl.TEXTURE0)
	//gl.BindVertexArray(gfxFont.(*FontRenderer_GL32).vao)

	//calculate alignment position
	if align == 0 {
		x -= f.Width(scale, spacingXAdd, fs, argv...) * 0.5
	} else if align < 0 {
		x -= f.Width(scale, spacingXAdd, fs, argv...)
	}
	textureID := int32(-1)
	spacing := spacingXAdd * scale
	renderedAny := false
	// Iterate through all characters in string
	for i := range indices {
		//get rune
		runeIndex := indices[i]

		//find rune in fontChar list
		ch, ok := f.fontChar[runeIndex]

		//load missing runes in batches of 32
		if !ok {
			low := runeIndex - (runeIndex % 32)
			f.GenerateGlyphs(low, low+31)
			ch, ok = f.fontChar[runeIndex]
		}

		//skip runes that are not in font character range
		if !ok {
			//fmt.Printf("%c %d\n", runeIndex, runeIndex)
			continue
		}

		if int32(len(batchVertices)/24) >= batchSize || (textureID != -1 && textureID != int32(ch.textureID)) {
			// Render the current batch
			f.renderGlyphBatch(batchVertices, uint32(textureID))
			// Clear the batch buffers
			batchVertices = make([]float32, 0, batchSize*6*4)
		}
		textureID = int32(ch.textureID)

		if renderedAny {
			x += spacing
		}

		//calculate position and size for current rune
		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.height-ch.bearingV)*scale
		w := float32(ch.width) * scale
		h := float32(ch.height) * scale
		vertices := []float32{
			xpos + w, ypos, ch.uv[2], ch.uv[1],
			xpos, ypos, ch.uv[0], ch.uv[1],
			xpos, ypos + h, ch.uv[0], ch.uv[3],

			xpos, ypos + h, ch.uv[0], ch.uv[3],
			xpos + w, ypos + h, ch.uv[2], ch.uv[3],
			xpos + w, ypos, ch.uv[2], ch.uv[1],
		}
		// Append glyph vertices to the batch buffer
		batchVertices = append(batchVertices, vertices...)
		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		x += float32((ch.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
		renderedAny = true
	}

	// Render any remaining glyphs in the batch
	if len(batchVertices) > 0 {
		f.renderGlyphBatch(batchVertices, uint32(textureID))
	}

	// Disable scissor just in case
	r.DisableScissor()

	return nil
}

// Helper function to render a batch of glyphs
func (f *Font_GL32) renderGlyphBatch(vertices []float32, textureID uint32) {
	fr := gfxFont.(*FontRenderer_GL32)

	gl.BindBuffer(gl.ARRAY_BUFFER, fr.vbo)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices))/4)
}

func (r *FontRenderer_GL32) ReleaseFontPipeline() {
	locVert := r.shaderProgram.attributes["vert"]
	gl.DisableVertexAttribArray(uint32(locVert))
	locTex := r.shaderProgram.attributes["vertTexCoord"]
	gl.DisableVertexAttribArray(uint32(locTex))

	//gl.BindVertexArray(0)
	//gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// Width returns the width of a piece of text in pixels
func (f *Font_GL32) Width(scale float32, spacingXAdd float32, fs string, argv ...interface{}) float32 {

	var width float32

	indices := []rune(fmt.Sprintf(fs, argv...))

	if len(indices) == 0 {
		return 0
	}

	spacing := spacingXAdd * scale
	renderedAny := false

	// Iterate through all characters in string
	for i := range indices {

		//get rune
		runeIndex := indices[i]

		//find rune in fontChar list
		ch, ok := f.fontChar[runeIndex]

		//load missing runes in batches of 32
		if !ok {
			low := runeIndex & rune(32-1)
			f.GenerateGlyphs(low, low+31)
			ch, ok = f.fontChar[runeIndex]
		}

		//skip runes that are not in font character range
		if !ok {
			//fmt.Printf("%c %d\n", runeIndex, runeIndex)
			continue
		}

		if renderedAny {
			width += spacing
		}

		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		width += float32((ch.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
		renderedAny = true
	}

	return width
}

// GenerateGlyphs builds a set of textures based on a ttf files glyphs
func (f *Font_GL32) GenerateGlyphs(low, high rune) error {
	//create a freetype context for drawing
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f.ttf)
	c.SetFontSize(float64(f.scale))
	c.SetHinting(font.HintingFull)

	//create new face to measure glyph dimensions
	ttfFace := truetype.NewFace(f.ttf, &truetype.Options{
		Size:    float64(f.scale),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Add padding to prevent cropping
	// https://github.com/ikemen-engine/Ikemen-GO/issues/3085
	padding := 2

	//make each glyph
	for ch := low; ch <= high; ch++ {
		char := new(character)

		gBnd, gAdv, ok := ttfFace.GlyphBounds(ch)
		if ok != true {
			return fmt.Errorf("ttf face glyphBounds error")
		}

		gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)

		//if glyph has no dimensions set to a max value
		if gw == 0 || gh == 0 {
			gBnd = f.ttf.Bounds(fixed.Int26_6(f.scale))
			gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
			gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)

			//above can sometimes yield 0 for font smaller than 48pt, 1 is minimum
			if gw == 0 || gh == 0 {
				gw = 1
				gh = 1
			}
		}

		//The glyph's ascent and descent equal -bounds.Min.Y and +bounds.Max.Y.
		gAscent := int(-gBnd.Min.Y) >> 6
		gdescent := int(gBnd.Max.Y) >> 6

		//set w,h and adv, bearing V and bearing H in char
		//char.width = int(gw)
		//char.height = int(gh)
		char.width = int(gw) + (padding * 2)
		char.height = int(gh) + (padding * 2)
		char.advance = int(gAdv)
		char.bearingV = gdescent
		//char.bearingH = (int(gBnd.Min.X) >> 6)
		char.bearingH = (int(gBnd.Min.X) >> 6) - padding

		//create image to draw glyph
		fg, bg := image.White, image.Black
		//rect := image.Rect(0, 0, int(gw), int(gh))
		rect := image.Rect(0, 0, char.width, char.height)
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

		//set the glyph dot
		//px := 0 - (int(gBnd.Min.X) >> 6)
		//py := (gAscent)
		px := padding - (int(gBnd.Min.X) >> 6)
		py := padding + gAscent
		pt := freetype.Pt(px, py)

		// Draw the text from mask to image
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetSrc(fg)
		_, err := c.DrawString(string(ch), pt)
		if err != nil {
			return err
		}

		var uv [4]float32
		textureIndex := 0
		w, h := int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy())
		pix := rgba.Pix
		stride := int32(rgba.Stride) // This was added to unify desktop and Android

		for {
			if textureIndex >= len(f.textures) {
				f.textures = append(f.textures, CreateTextureAtlas(256, 256, 32, true))
			}

			var inserted bool
			uv, inserted = f.textures[textureIndex].AddImage(w, h, stride, pix)

			if inserted {
				break
			}

			textureIndex++
		}

		texAtlas := f.textures[textureIndex]

		// This block is no longer necessary after the padding fix and actually introduces blur
		// aw := float32(texAtlas.width)
		// ah := float32(texAtlas.height)
		// off_u := 0.5 / aw
		// off_v := 0.5 / ah
		// uv[0] += off_u
		// uv[1] += off_v
		// uv[2] -= off_u
		// uv[3] -= off_v

		char.uv = uv
		char.textureID = texAtlas.texture.(*Texture_GL32).handle

		//add char to fontChar list
		f.fontChar[ch] = char
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

// LoadTrueTypeFont builds OpenGL buffers and glyph textures based on a ttf file
func (r *FontRenderer_GL32) LoadTrueTypeFont(reader io.Reader, scale int32, low, high rune, dir Direction) (*Font_GL32, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Read the truetype font.
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	//make Font stuct type
	f := new(Font_GL32)
	f.fontChar = make(map[rune]*character)
	f.ttf = ttf
	f.scale = scale
	f.SetColor(1.0, 1.0, 1.0, 1.0) //set default white
	f.textures = append(f.textures, CreateTextureAtlas(256, 256, 32, true))

	err = f.GenerateGlyphs(low, high)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// newProgram links the frag and vertex shader programs
func (r *FontRenderer_GL32) newProgram(GLSLVersion uint, vertexShaderSource, fragmentShaderSource string) {
	shaderProgram, _ := gfx.(*Renderer_GL32).newShaderProgram(vertexShaderSource, fragmentShaderSource, "", "font shader", true)
	r.shaderProgram = shaderProgram
	r.shaderProgram.RegisterUniforms("textColor", "resolution", "tex")
}
