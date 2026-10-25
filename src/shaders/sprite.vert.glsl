#if __VERSION__ >= 450
	// VULKAN PATH
	layout(binding = 0) uniform UniformBufferObject {
		mat4 modelview, projection;
	};
	layout(location = 0) in vec2 position;
	layout(location = 1) in vec2 uv;
	layout(location = 0) out vec2 texcoord;
#else
	// OPENGL / GLES PATH
	#define COMPAT_VARYING out
	#define COMPAT_ATTRIBUTE in
	#define COMPAT_TEXTURE texture
	#ifdef GL_ES
		// Mandatory for GLES: High precision for vertex position math
		precision highp float;
		precision highp int;
	#endif

	uniform mat4 modelview, projection;
	uniform vec4 quad[4]; // x, y, u, v per vertex
	uniform bool useQuad;

	COMPAT_ATTRIBUTE vec2 position;
	COMPAT_ATTRIBUTE vec2 uv;
	COMPAT_VARYING vec2 texcoord;
#endif

void main(void) {
	vec2 pos = position;
	texcoord = uv;
	#if __VERSION__ < 450
		// Single quads come from uniforms to skip a vertex buffer upload per sprite
		if (useQuad) {
			pos = quad[gl_VertexID].xy;
			texcoord = quad[gl_VertexID].zw;
		}
	#endif
	gl_Position = projection * (modelview * vec4(pos, 0.0, 1.0));
	
	#if __VERSION__ >= 450
		// Vulkan's Y-axis is inverted compared to OpenGL
		gl_Position.y = -gl_Position.y;
	#endif
}