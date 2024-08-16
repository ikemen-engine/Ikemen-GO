#if __VERSION__ >= 130
#define COMPAT_VARYING out
#define COMPAT_ATTRIBUTE in
#define COMPAT_TEXTURE texture
#else
#define COMPAT_VARYING varying 
#define COMPAT_ATTRIBUTE attribute 
#define COMPAT_TEXTURE texture2D
#endif

uniform mat4 modelview, projection;

COMPAT_ATTRIBUTE vec2 position;
COMPAT_ATTRIBUTE vec2 uv;
COMPAT_VARYING vec2 texcoord;

void main(void) {
	texcoord = uv;
	gl_Position = projection * (modelview * vec4(position, 0.0, 1.0));
}
