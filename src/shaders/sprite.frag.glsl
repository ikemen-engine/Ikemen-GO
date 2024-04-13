uniform sampler2D tex;
uniform sampler2D pal;

uniform vec4 x1x2x4x3;
uniform vec4 tint;
uniform vec3 add, mult;
uniform float alpha, gray, hue;
uniform int mask;
uniform bool isFlat, isRgba, isTrapez, neg;

varying vec2 texcoord;

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
			if (mask == -1) {
				c.a = 1.0;
			}
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
			if (mask == -1) {
				c.a = 1.0;
			}
		}
		if (hue != 0) {
			c.rgb = hue_shift(c.rgb,hue);			
		}
		if (neg) c.rgb = neg_base - c.rgb;
		c.rgb = mix(c.rgb, vec3((c.r + c.g + c.b) / 3.0), gray) + final_add;
		c *= final_mul;

		// Add a final tint (used for shadows); make sure the result has premultiplied alpha
		c.rgb = mix(c.rgb, tint.rgb * c.a, tint.a);

		gl_FragColor = c;
	}
}
