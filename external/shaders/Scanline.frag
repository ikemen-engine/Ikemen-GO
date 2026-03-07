#if __VERSION__ >= 450
#define COMPAT_TEXTURE texture
layout(push_constant, std430) uniform u {
	uniform vec2 TextureSize;
};
layout(binding = 0) uniform sampler2D Texture;
layout(location = 0) in vec3 vTexCoord;
layout(location = 0) out vec4 FragColor;
#else
#if __VERSION__ >= 130
#define COMPAT_VARYING in
#define COMPAT_TEXTURE texture
out vec4 FragColor;
in vec3 vTexCoord;
#else
#define COMPAT_TEXTURE texture2D
#define vTexCoord gl_TexCoord[0]
#define FragColor gl_FragColor
#endif
uniform vec2 TextureSize;
uniform sampler2D Texture;
#endif


void main(void) {
    vec4 rgb = COMPAT_TEXTURE(Texture, vTexCoord.xy);
    vec4 intens;
    if (fract(gl_FragCoord.y * (0.5 * 4.0 / 3.0)) > 0.5)
        intens = vec4(0);
    else
        intens = smoothstep(0.2, 0.8, rgb) + normalize(vec4(rgb.xyz, 1.0));

    float level = (4.0 - vTexCoord.z) * 0.19;
    FragColor = intens * (0.5 - level) + rgb * 1.1;
}