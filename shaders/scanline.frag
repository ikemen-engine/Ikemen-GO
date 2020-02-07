uniform sampler2D Texture;
uniform vec2 TextureSize;

void main(void) {
	vec4 rgb = texture2D(Texture, gl_TexCoord[0].xy);
	vec4 intens ;
	if (fract(gl_FragCoord.y * (0.5*4.0/3.0)) > 0.5)
		intens = vec4(0);
	else
		intens = smoothstep(0.2,0.8,rgb) + normalize(vec4(rgb.xyz, 1.0));
	float level = (4.0-gl_TexCoord[0].z) * 0.19;
	gl_FragColor = intens * (0.5-level) + rgb * 1.1 ;
}