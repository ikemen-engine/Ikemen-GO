uniform mat4 modelview, projection;
attribute vec3 position;
attribute vec2 uv;
attribute vec4 vertColor;
varying vec2 texcoord;
varying vec4 vColor;

void main(void) {
		texcoord = uv;
		vColor = vertColor;
		gl_Position = projection * (modelview * vec4(position, 1.0));
}