uniform sampler2D tex;
uniform sampler2D pal;

uniform vec4 x1x2x4x3;
uniform vec4 tint;
uniform vec3 add, mult;
uniform float alpha, gray;
uniform int mask;
uniform bool isFlat, isRgba, isTrapez, neg;

varying vec2 texcoord;

void main(void) {
	if (isFlat) {
		gl_FragColor = tint;

	} else {
		vec2 uv = texcoord;
		if (isTrapez) {
			// Compute left/right trapezoid bounds at height uv.y
			vec2 bounds = mix(x1x2x4x3.zw, x1x2x4x3.xy, uv.y);
			// Correct uv.x from the fragment position on that segment
			uv.x = (gl_FragCoord.x - bounds[0]) / (bounds[1] - bounds[0]);
		}

		vec4 c = texture2D(tex, uv);
		vec3 neg_base = vec3(1.0);
		vec3 final_add = add;
		vec4 final_mul = vec4(mult, alpha);
		if (isRgba) {
			// RGBA sprites use premultiplied alpha for transparency
			neg_base *= c.a;
			final_add *= c.a;
			final_mul.rgb *= alpha;
		} else {
			// Colormap sprites use the old “buggy” Mugen way
			if (int(255.25*c.r) == mask) {
				final_mul = vec4(0.0);
			} else {
				c = texture2D(pal, vec2(c.r*0.9966, 0.5));
			}
		}

		if (neg) c.rgb = neg_base - c.rgb;
		c.rgb = mix(c.rgb, vec3((c.r + c.g + c.b) / 3.0), gray) + final_add;
		c *= final_mul;

		// Add a final tint (used for shadows); make sure the result has premultiplied alpha
		c.rgb = mix(c.rgb, tint.rgb * c.a, tint.a);

		gl_FragColor = c;
	}
}
