// “Flat” fragment shader, for shadows and plain, untextured quads

uniform sampler2D tex;
uniform vec3 color;
uniform float alpha;
uniform bool isShadow;

varying vec2 texcoord;

void main(void) {
	vec4 p = vec4(color, alpha);
	if (isShadow)
		p *= texture2D(tex, texcoord).a;
	gl_FragColor = p;
}
