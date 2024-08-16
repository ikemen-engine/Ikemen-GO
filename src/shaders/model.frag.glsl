#if __VERSION__ >= 130
#define COMPAT_VARYING in
#define COMPAT_TEXTURE texture
out vec4 FragColor;
#else
#define COMPAT_VARYING varying
#define FragColor gl_FragColor
#define COMPAT_TEXTURE texture2D
#endif

uniform sampler2D tex;
uniform vec4 baseColorFactor;
uniform vec3 add, mult;
uniform float gray, hue;
uniform bool textured;
uniform bool neg;
uniform bool enableAlpha;
uniform float alphaThreshold;

COMPAT_VARYING vec2 texcoord;
COMPAT_VARYING vec4 vColor;

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
		FragColor = COMPAT_TEXTURE(tex, texcoord) * baseColorFactor;
	} else {
		FragColor = baseColorFactor;
	}
	FragColor *= vec4(pow(vColor.r, 1.0/2.2), pow(vColor.g, 1.0/2.2), pow(vColor.b, 1.0/2.2), vColor.a);
	FragColor.rgb *= vColor.a;
	if(!enableAlpha){
		if(FragColor.a < alphaThreshold){
			discard;
		}else{
			FragColor.a = 1;
		}
	}else if(FragColor.a<=0.0){
		discard;
	}
	vec3 neg_base = vec3(1.0);
	neg_base *= FragColor.a;
	if (hue != 0) {
		FragColor.rgb = hue_shift(FragColor.rgb,hue);			
	}
	if (neg) FragColor.rgb = neg_base - FragColor.rgb;
	FragColor.rgb = mix(FragColor.rgb, vec3((FragColor.r + FragColor.g + FragColor.b) / 3.0), gray) + add*FragColor.a;
	FragColor.rgb *= mult;
}