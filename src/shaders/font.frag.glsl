#if __VERSION__ >= 450
	// VULKAN PATH
	#define COMPAT_TEXTURE texture
	#define COMPAT_FRAGCOLOR FragColor
	layout(location = 0) out vec4 FragColor;
	layout(location = 0) in vec2 fragTexCoord;

	layout(push_constant, std430) uniform u {
		vec4 textColor;
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
	uniform sampler2D tex;
	COMPAT_VARYING vec2 fragTexCoord;
#endif


void main()
{
	vec4 texColor = COMPAT_TEXTURE(tex, fragTexCoord);
	vec4 sampled = vec4(1.0, 1.0, 1.0, texColor.r);
    COMPAT_FRAGCOLOR = min(textColor, vec4(1.0, 1.0, 1.0, 1.0)) * sampled;
}