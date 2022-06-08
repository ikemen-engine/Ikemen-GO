uniform sampler2D Texture;

varying vec2 texcoord;

void main(void) {
	gl_FragColor = texture2D(Texture, texcoord);
}
