// “Flat” fragment shader, for shadows and plain, untextured quads
#version 400
precision highp float;

uniform sampler2D tex;
uniform vec3 color;
uniform float alpha;
uniform bool isShadow;

in vec2 texcoord;
out vec4 FragColor;

void main() {
	vec4 p = vec4(color, alpha);
	if (isShadow)
		p *= texture(tex, texcoord).a;
	FragColor = p;
}
