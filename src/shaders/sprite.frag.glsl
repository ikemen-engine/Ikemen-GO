// Engine sprite shader
// The sprite drawing code lives in sprite_lib.frag.glsl so custom shaders can include it too
#include "sprite_lib.frag.glsl"

void main(void) {
	vec2 uv = ikemenSpriteUV();
	FragColor = ikemenSpriteColor(uv);
}
