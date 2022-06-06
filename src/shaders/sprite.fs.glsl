uniform sampler2D tex;
uniform sampler2D pal;

uniform vec4 x1x2x4x3;
uniform vec3 add, mul;
uniform float alpha, gray;
uniform int mask;
uniform bool isRgba, isTrapez, neg;

varying vec2 texcoord;

void main(void) {
	vec2 uv = texcoord;
	if (isTrapez) {
		// ここから台形用のテクスチャ座標計算/ Compute texture coordinates for trapezoid from here
		float left = -mix(x1x2x4x3[2], x1x2x4x3[0], uv[1]);
		float right = mix(x1x2x4x3[3], x1x2x4x3[1], uv[1]);
		uv[0] = (left + gl_FragCoord.x) / (left + right); // ここまで / To this point
	}
	vec4 c = texture2D(tex, uv);
	vec3 neg_base = vec3(1.0);
	vec3 final_add = add;
	vec4 final_mul = vec4(mul, alpha);
	if (isRgba) {
		neg_base *= alpha;
		final_add *= c.a;
		final_mul.rgb *= alpha;
	} else {
		if (int(255.25*c.r) == mask) {
			c.a = 0.0;
		} else {
			c = texture2D(pal, vec2(c.r*0.9966, 0.5));
		}
	}
	if (neg) c.rgb = neg_base - c.rgb;
	c.rgb = mix(c.rgb, vec3((c.r + c.g + c.b) / 3.0), gray) + final_add;
	gl_FragColor = c * final_mul;
}
