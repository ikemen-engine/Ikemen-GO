#if __VERSION__ >= 130
#define COMPAT_VARYING out
#define COMPAT_ATTRIBUTE in
#define COMPAT_TEXTURE texture
#else
#define COMPAT_VARYING varying 
#define COMPAT_ATTRIBUTE attribute 
#define COMPAT_TEXTURE texture2D
#endif

uniform vec2 TextureSize;

COMPAT_ATTRIBUTE vec2 VertCoord;
COMPAT_VARYING vec2 texcoord;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	texcoord = (VertCoord + 1.0) / 2.0;
}
