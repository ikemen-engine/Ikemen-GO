#if __VERSION__ >= 450
	// VULKAN PATH
	layout(location = 0) in vec2 VertCoord;
	layout(location = 0) out vec2 texcoord;
#else
	// OPENGL / GLES PATH
	#define COMPAT_VARYING out
	#define COMPAT_ATTRIBUTE in
	#ifdef GL_ES
		precision highp float;
	#endif

	uniform vec2 TextureSize; // Not used
	COMPAT_ATTRIBUTE vec2 VertCoord;
	COMPAT_VARYING vec2 texcoord; // TODO: Casing doesn't match Go
#endif

void main() {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	// Standard quad-to-UV mapping
	texcoord = (VertCoord + 1.0) / 2.0;
}
