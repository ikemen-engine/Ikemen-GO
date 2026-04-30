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
		vec4 palNegPad;    // x neg (0/1)
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
	uniform int palNeg;
	uniform sampler2D tex;
	COMPAT_VARYING vec2 fragTexCoord;
#endif

vec3 rgb2hsv(vec3 c)
{
	vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
	vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
	vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

	float d = q.x - min(q.w, q.y);
	float e = 1.0e-10;
	return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c)
{
	vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
	vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
	return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

vec3 hue_shift(vec3 color, float dhue) {
	vec3 colorhsv = rgb2hsv(color);
	colorhsv.x = mod(colorhsv.x+dhue, 1.0);
	return hsv2rgb(colorhsv);
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
	float negV;
	#if __VERSION__ >= 450
		addV = palAddGray.xyz;
		grayV = palAddGray.w;
		mulV = palMulHue.xyz;
		hueV = palMulHue.w;
		negV = palNegPad.x;
	#else
		addV = palAdd;
		mulV = palMul;
		grayV = palGray;
		hueV = palHue;
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
	// Add
	c.rgb += addV * c.a;
	// Mul
	c.rgb *= mulV;

	COMPAT_FRAGCOLOR = c;
}