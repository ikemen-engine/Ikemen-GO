uniform mat4 modelview, projection;
attribute vec3 position;
attribute vec2 uv;
varying vec2 texcoord;

void main(void) {
		texcoord = uv;
		gl_Position = projection * (modelview * vec4(position, 1.0));
}