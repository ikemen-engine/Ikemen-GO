#if __VERSION__ >= 450
	// VULKAN PATH
	#define COMPAT_TEXTURE texture
	#define COMPAT_FRAGCOLOR FragColor
	layout(location = 0) out vec4 FragColor;
	layout(location = 0) in vec2 fragTexCoord;

	layout(push_constant, std430) uniform u {
		vec4 textColor;
		vec4 palAddGray;   // xyz add, w gray
		vec4 palMulHue;    // xyz mul, w hue
		vec4 palNegPad;    // x neg (0/1), y scroll
	};
	layout(binding = 0) uniform sampler2D tex;
#else
	// OPENGL / GLES PATH
	#ifdef GL_ES
		precision highp float;
		precision highp int;
	#endif

	#define COMPAT_VARYING in
	#define COMPAT_TEXTURE texture
	#define COMPAT_FRAGCOLOR FragColor
	out vec4 FragColor;

	// These must be STANDALONE for RegisterUniforms to work in GLES
	uniform vec4 textColor;
	uniform vec3 palAdd;
	uniform vec3 palMul;
	uniform float palGray;
	uniform float palHue;
	uniform float palScroll;
	uniform int palNeg;
	uniform sampler2D tex;
	COMPAT_VARYING vec2 fragTexCoord;
#endif

// Rotates hue around the gray axis with Rodrigues' rotation formula, then restores luma so brightness stays about the same
vec3 hue_shift(vec3 color, float dhue) {
	float angle = dhue * 6.28318531;
	float cs = cos(angle);
	vec3 k = vec3(0.57735); // Unit gray axis
	vec3 c = color * cs + cross(k, color) * sin(angle) + k * dot(k, color) * (1.0 - cs);
	// Rotating around gray keeps only the RGB average, so add back the lost luma
	return clamp(c + dot(color - c, vec3(0.299, 0.587, 0.114)), 0.0, 1.0);
}


void main()
{
	vec4 texColor = COMPAT_TEXTURE(tex, fragTexCoord);
	vec4 sampled = vec4(1.0, 1.0, 1.0, texColor.r);
	vec4 c = min(textColor, vec4(1.0, 1.0, 1.0, 1.0)) * sampled;

	vec3 addV;
	vec3 mulV;
	float grayV;
	float hueV;
	float scrollV;
	float negV;
	#if __VERSION__ >= 450
		addV = palAddGray.xyz;
		grayV = palAddGray.w;
		mulV = palMulHue.xyz;
		hueV = palMulHue.w;
		scrollV = palNegPad.y;
		negV = palNegPad.x;
	#else
		addV = palAdd;
		mulV = palMul;
		grayV = palGray;
		hueV = palHue;
		scrollV = palScroll;
		negV = float(palNeg);
	#endif

	// Apply PalFX
	// Hue
	if (hueV != 0.0) {
		c.rgb = hue_shift(c.rgb, hueV);
	}
	// Invertall
	if (negV > 0.5) {
		c.rgb = c.aaa - c.rgb;
	}
	// Color
	c.rgb = mix(vec3((c.r + c.g + c.b) / 3.0), c.rgb, 1.0 - grayV);
	// Scroll: shifts brightness and wraps from bright back to dark
	// Text color isn't multiplied by the glyph's alpha, so no alpha correction here
	if (scrollV != 0.0) {
		float l = (max(max(c.r, c.g), c.b) + min(min(c.r, c.g), c.b)) * 0.5;
		float n = fract(l + scrollV);
		// Keep saturation by scaling the color difference to the room left at the new lightness
		c.rgb = n + (c.rgb - l) * min(n, 1.0 - n) / max(min(l, 1.0 - l), 1e-5);
	}
	// Add
	c.rgb += addV * c.a;
	// Mul
	c.rgb *= mulV;

	COMPAT_FRAGCOLOR = c;
}