// “Flat” fragment shader, for shadows and plain, untextured quads
#version 400
precision highp float;

uniform sampler2D tex;
uniform vec3 color;
uniform float alpha;

out vec4 FragColor;

void main() {
	vec4 p = vec4(color, alpha);
	FragColor = p;
}
