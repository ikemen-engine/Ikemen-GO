uniform sampler2D tex;
uniform vec4 baseColorFactor;
uniform vec3 add, mult;
uniform float gray, hue;
uniform bool textured;
uniform bool neg;
uniform bool enableAlpha;
uniform float alphaThreshold;

varying vec2 texcoord;
varying vec4 vColor;

vec3 hue_shift(vec3 color, float dhue) {
	float s = sin(dhue);
	float c = cos(dhue);
	return (color * c) + (color * s) * mat3(
		vec3(0.167444, 0.329213, -0.496657),
		vec3(-0.327948, 0.035669, 0.292279),
		vec3(1.250268, -1.047561, -0.202707)
	) + dot(vec3(0.299, 0.587, 0.114), color) * (1.0 - c);
}
void main(void) {
	if(textured){
		gl_FragColor = texture2D(tex, texcoord) * baseColorFactor;
	} else {
		gl_FragColor = baseColorFactor;
	}
	gl_FragColor *= vec4(pow(vColor.r, 1.0/2.2), pow(vColor.g, 1.0/2.2), pow(vColor.b, 1.0/2.2), vColor.a);
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
	if (hue != 0) {
		gl_FragColor.rgb = hue_shift(gl_FragColor.rgb,hue);			
	}
	if (neg) gl_FragColor.rgb = neg_base - gl_FragColor.rgb;
	gl_FragColor.rgb = mix(gl_FragColor.rgb, vec3((gl_FragColor.r + gl_FragColor.g + gl_FragColor.b) / 3.0), gray) + add*gl_FragColor.a;
	gl_FragColor.rgb *= mult;
}