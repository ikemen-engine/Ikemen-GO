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
	vec2 texCoord = (VertCoord + 1.0) / 2.0;
	float x = 0.001;
	float y = 0.001;
	
	vec2 dg1 = vec2( x,y);  vec2 dg2 = vec2(-x,y);
	vec2 sd1 = dg1*0.5;     vec2 sd2 = dg2*0.5;
	vec2 ddx = vec2(x,0.0); vec2 ddy = vec2(0.0,y);
	
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	
	TexCoord[0].xy = texCoord;
	TexCoord[1].xy = texCoord - sd1;
	TexCoord[2].xy = texCoord - sd2;
	TexCoord[3].xy = texCoord + sd1;
	TexCoord[4].xy = texCoord + sd2;
	TexCoord[5].xy = texCoord - dg1;
	TexCoord[6].xy = texCoord + dg1;
	TexCoord[5].zw = texCoord - dg2;
	TexCoord[6].zw = texCoord + dg2;
	TexCoord[1].zw = texCoord - ddy;
	TexCoord[2].zw = texCoord + ddx;
	TexCoord[3].zw = texCoord + ddy;
	TexCoord[4].zw = texCoord - ddx;
}