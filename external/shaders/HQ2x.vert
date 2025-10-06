#if __VERSION__ >= 130
#define COMPAT_VARYING out
#define COMPAT_ATTRIBUTE in
#define COMPAT_TEXTURE texture
out vec4 TexCoord[7];
#else
#define COMPAT_VARYING varying
#define COMPAT_ATTRIBUTE attribute
#define COMPAT_TEXTURE texture2D
#define TexCoord gl_TexCoord
#endif

COMPAT_ATTRIBUTE vec2 VertCoord;
uniform vec2 TextureSize;

void main() {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	
	vec2 texCoord = (VertCoord + 1.0) / 2.0;
	float x = 0.5 * (1.0 / TextureSize.x);
	float y = 0.5 * (1.0 / TextureSize.y);
	vec2 dg1 = vec2( x, y);
	vec2 dg2 = vec2(-x, y);
	vec2 dx = vec2(x, 0.0);
	vec2 dy = vec2(0.0, y);
	
	TexCoord[0].xy = texCoord;
	TexCoord[1].xy = texCoord - dg1;
	TexCoord[1].zw = texCoord - dy;
	TexCoord[2].xy = texCoord - dg2;
	TexCoord[2].zw = texCoord + dx;
	TexCoord[3].xy = texCoord + dg1;
	TexCoord[3].zw = texCoord + dy;
	TexCoord[4].xy = texCoord + dg2;
	TexCoord[4].zw = texCoord - dx;
}