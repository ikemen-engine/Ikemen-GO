attribute vec2 VertCoord;
uniform vec2 TextureSize;

void main() {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	
	vec2 TexCoord = (VertCoord + 1.0) / 2.0;
	float x = 0.5 * (1.0 / TextureSize.x);
	float y = 0.5 * (1.0 / TextureSize.y);
	vec2 dg1 = vec2( x, y);
	vec2 dg2 = vec2(-x, y);
	vec2 dx = vec2(x, 0.0);
	vec2 dy = vec2(0.0, y);
	
	gl_TexCoord[0].xy = TexCoord;
	gl_TexCoord[1].xy = gl_TexCoord[0].xy - dg1;
	gl_TexCoord[1].zw = gl_TexCoord[0].xy - dy;
	gl_TexCoord[2].xy = gl_TexCoord[0].xy - dg2;
	gl_TexCoord[2].zw = gl_TexCoord[0].xy + dx;
	gl_TexCoord[3].xy = gl_TexCoord[0].xy + dg1;
	gl_TexCoord[3].zw = gl_TexCoord[0].xy + dy;
	gl_TexCoord[4].xy = gl_TexCoord[0].xy + dg2;
	gl_TexCoord[4].zw = gl_TexCoord[0].xy - dx;
}