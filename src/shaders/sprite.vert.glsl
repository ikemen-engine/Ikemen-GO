#version 400
precision highp float;

uniform mat4 modelview, projection;

in vec2 position;
in vec2 uv;
out vec2 texcoord;

void main(void) {
	texcoord = uv;
	gl_Position = projection * (modelview * vec4(position, 0.0, 1.0));
}
