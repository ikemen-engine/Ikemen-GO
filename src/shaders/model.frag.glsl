uniform sampler2D tex;
uniform vec4 baseColorFactor;
uniform vec3 add, mult;
uniform float gray;
uniform bool textured;
uniform bool neg;
uniform bool enableAlpha;
uniform float alphaThreshold;

varying vec2 texcoord;

void main(void) {
	if(textured){
		gl_FragColor = texture2D(tex, texcoord) * baseColorFactor;
	} else {
		gl_FragColor = baseColorFactor;
	}
	
	if(!enableAlpha){
		if(gl_FragColor.a < alphaThreshold){
			discard;
		}else{
			gl_FragColor.a = 1;
		}
	}else if(gl_FragColor.a<=0.0){
		discard;
	}
	vec3 neg_base = vec3(1.0);
	neg_base *= gl_FragColor.a;
	if (neg) gl_FragColor.rgb = neg_base - gl_FragColor.rgb;
	gl_FragColor.rgb = mix(gl_FragColor.rgb, vec3((gl_FragColor.r + gl_FragColor.g + gl_FragColor.b) / 3.0), gray) + add*gl_FragColor.a;
	gl_FragColor.rgb *= mult;
}