#define MATH_PI 3.1415926535897932384626433832795
#define MATH_INV_PI (1.0 / MATH_PI)

#if __VERSION__ >= 450
	// VULKAN PATH
	#extension GL_EXT_multiview : enable
	#define COMPAT_TEXTURE texture
	#define currentFace gl_ViewIndex
	layout(location = 0) in vec2 texcoord;
	layout(binding = 0) uniform sampler2D panorama;
	layout(location = 0) out vec4 FragColor;
#else
	// OPENGL / GLES PATH
	#define COMPAT_VARYING in
	#define COMPAT_TEXTURE texture
	#ifdef GL_ES
		precision highp float;
		precision highp int;
	#endif
	out vec4 FragColor;

	COMPAT_VARYING vec2 texcoord;
	uniform int currentFace;
	uniform sampler2D panorama;
#endif

vec3 uvToXYZ(int face, vec2 uv)
{
	if(face == 0)
		return vec3(     1.0,   uv.y,    -uv.x);

	else if(face == 1)
		return vec3(    -1.0,   uv.y,     uv.x);

	else if(face == 2)
		return vec3(   +uv.x,   -1.0,    +uv.y);

	else if(face == 3)
		return vec3(   +uv.x,    1.0,    -uv.y);

	else if(face == 4)
		return vec3(   +uv.x,   uv.y,      1.0);

	else //if(face == 5)
	{	return vec3(    -uv.x,  +uv.y,     -1.0);}
}

vec2 dirToUV(vec3 dir)
{
	return vec2(
		0.5f + 0.5f * atan(dir.z, dir.x) / MATH_PI,
		1.f - acos(dir.y) / MATH_PI);
}

vec3 panoramaToCubeMap(int face, vec2 texCoord)
{
	vec2 texCoordNew = texCoord*2.0-1.0;
	vec3 scan = uvToXYZ(face, texCoordNew);
	vec3 direction = normalize(scan);
	//direction.y = -direction.y;
	vec2 src = dirToUV(direction);

	return  COMPAT_TEXTURE(panorama, src).rgb;
}



void main(void)
{
	FragColor = vec4(0.0, 0.0, 0.0, 1.0);

	FragColor.rgb = panoramaToCubeMap(currentFace, texcoord);
}