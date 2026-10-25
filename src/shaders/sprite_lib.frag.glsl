// Ikemen GO sprite shader library
// Custom sprite shaders #include "sprite_lib.frag.glsl", get the sprite coordinate with ikemenSpriteUV()
// and the normal sprite color with ikemenSpriteColor(uv)
// Shaders that use the GrabPass declare bgl_RenderedTexture themselves, since naming it enables the back buffer copy
// Vulkan build command for a custom shader:
//   glslc -fshader-stage=frag -std=450core -I <folder of this file> custom.frag -o custom.frag.spv

#if __VERSION__ >= 450
	// VULKAN PATH
	#define COMPAT_TEXTURE texture
	layout(binding = 1) uniform UniformBufferObject {
		vec4 x1x2x4x3;
		vec4 tint;
		vec3 add;
		vec3 mult;
		float alpha, gray, hue;
		int mask;
		bool isFlat, isRgba, isTrapez, neg;
		float iTime;
		vec2 iResolution;
		float aspectRatio;
		float sTime;
	};
	layout(push_constant, std430) uniform u {
		vec4 spriteUV;
		vec4 palUV;
		vec4 tex1UV;
		vec4 tex2UV;
		float param[16];
	};
	layout(binding = 2) uniform sampler2D tex;
	layout(binding = 3) uniform sampler2D pal;
	layout(binding = 5) uniform sampler2D tex1;
	layout(binding = 6) uniform sampler2D tex2;
	layout(location = 0) in vec2 texcoord;
	layout(location = 0) out vec4 FragColor;
#else
	// OPENGL / GLES PATH
	#define COMPAT_VARYING in
	#define COMPAT_TEXTURE texture
	#ifdef GL_ES
		precision highp float;
		precision highp int;
	#endif
	out vec4 FragColor;

	uniform sampler2D tex;
	uniform sampler2D pal;
	uniform vec4 x1x2x4x3;
	uniform vec4 tint;
	uniform vec3 add, mult;
	uniform float alpha, gray, hue;
	uniform int mask;
	uniform bool isFlat, isRgba, isTrapez, neg;
	uniform vec4 spriteUV;
	uniform vec4 palUV;
	uniform float iTime, sTime, aspectRatio;
	uniform vec2 iResolution;
	uniform vec4 tex1UV, tex2UV;
	uniform float param[16];
	uniform sampler2D tex1, tex2;
	COMPAT_VARYING vec2 texcoord;
#endif

vec3 ikemenRgbToHsv(vec3 c)
{
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 ikemenHsvToRgb(vec3 c)
{
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

vec3 ikemenHueShift(vec3 color, float dhue) {
	vec3 colorhsv = ikemenRgbToHsv(color);
	colorhsv.x = mod(colorhsv.x+dhue, 1.0);
	return ikemenHsvToRgb(colorhsv);
}

// Sprite-local coordinate, 0-1 across the sprite even when it's packed in an atlas
vec2 ikemenSpriteUV() {
	return (texcoord - spriteUV.xy) / (spriteUV.zw - spriteUV.xy);
}

// The engine's sprite color at a sprite-local coordinate
vec4 ikemenSpriteColor(vec2 uv) {
	vec4 c;
	vec3 neg_base = vec3(1.0);
	vec3 final_add = add;
	vec4 final_mul = vec4(mult, alpha);

	// Select flat color or textures
	if (isFlat) {
		c = tint; 

		// Treat flat colors like RGBA for math consistency
		neg_base *= c.a;
		final_add *= c.a;
		final_mul.rgb *= alpha;
	} else {
		if (isTrapez) {
			vec2 bounds = mix(x1x2x4x3.zw, x1x2x4x3.xy, clamp(uv.y, 0.0, 1.0));
			float gap = bounds[1] - bounds[0];
			#ifdef GL_ES
				if (abs(gap) < 0.0001) gap = 0.0001;
			#endif
			uv.x = clamp((gl_FragCoord.x - bounds[0]) / gap, 0.0, 1.0);
		}

		// Map into the sprite's atlas area, clamped so sampling can't reach neighboring sprites
		vec2 atlasUV = mix(spriteUV.xy, spriteUV.zw, uv);
		atlasUV = clamp(atlasUV, spriteUV.xy, spriteUV.zw - 0.0001);
		c = COMPAT_TEXTURE(tex, atlasUV);

		// Select with or without palette
		if (isRgba) {
			if (mask == -1) c.a = 1.0;
			neg_base *= c.a;
			final_add *= c.a;
			final_mul.rgb *= alpha;
		} else {
			// Palette lookup
			c = COMPAT_TEXTURE(pal, vec2(palUV[0]+palUV[2]*c.r*0.9966, palUV[1]));
			if (mask == -1) c.a = 1.0;
		}
	}

	// Apply PalFX
	// Hue
	if (hue != 0.0) {
		c.rgb = ikemenHueShift(c.rgb, hue);
	}
	// Invertall
	if (neg) {
		c.rgb = neg_base - c.rgb;
	}
	// Color
	c.rgb = mix(vec3((c.r + c.g + c.b) / 3.0), c.rgb, 1.0 - gray);
	// Add
	c.rgb += final_add;
	// Mul
	c *= final_mul;

	// Apply tint
	// Sprites only, because flat colors are already tinted
	if (!isFlat) {
		c.rgb = mix(c.rgb, tint.rgb * c.a, tint.a);
	}

	return c;
}
