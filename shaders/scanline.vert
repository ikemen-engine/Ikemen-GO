uniform vec2 TextureSize;
attribute vec2 VertCoord;

void main(void) {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	gl_TexCoord[0].xy = (VertCoord + 1.0) / 2.0;
}