/*
 * Scanline.vert
 * Inputs: VertCoord attribute for quad vertices and optional TexCoord.z for line phase; TextureSize uniform scales UVs.
 * Coordinate spaces: Converts VertCoord from NDC to UV and passes through user-supplied scanline index.
 * Performance: Minimal arithmetic and no texture access; vertex cost is trivial.
 */
#if __VERSION__ >= 130
#define COMPAT_VARYING out
#define COMPAT_ATTRIBUTE in
in vec3 TexCoord;
out vec3 vTexCoord;
#else
#define COMPAT_VARYING varying 
#define COMPAT_ATTRIBUTE attribute 
#endif

uniform vec2 TextureSize;
COMPAT_ATTRIBUTE vec2 VertCoord;

void main(void) {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
#if __VERSION__ >= 130
    vTexCoord = vec3((VertCoord + 1.0) / 2.0, TexCoord.z);
#else
    gl_TexCoord[0].xy = (VertCoord + 1.0) / 2.0;
#endif
}
