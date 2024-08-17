#if __VERSION__ >= 130
#define COMPAT_VARYING in
#define COMPAT_TEXTURE texture
out vec4 FragColor;
#else
#define COMPAT_VARYING varying
#define FragColor gl_FragColor
#define COMPAT_TEXTURE texture2D
#endif

uniform sampler2D Texture;

COMPAT_VARYING vec2 texcoord;

void main(void) {
    FragColor = COMPAT_TEXTURE(Texture, texcoord);
}