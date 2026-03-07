#if __VERSION__ >= 450
	// VULKAN PATH
	#define ENABLE_SHADOW
	#define COMPAT_TEXTURE texture
	#define COMPAT_TEXTURE_CUBE texture
	#define COMPAT_TEXTURE_CUBE_LOD textureLod
	#define COMPAT_SHADOW_MAP_TEXTURE() texture(shadowCubeMap,vec4(1.0, -(xy.y*2-1),-(xy.x*2-1),index)).r
	#define COMPAT_SHADOW_CUBE_MAP_TEXTURE() texture(shadowCubeMap,vec4(xyz,index)).r
	struct Light {
		vec3 direction; float range;
		vec3 color; float intensity;
		vec3 position; float innerConeCos;
		float outerConeCos; int type;
		float shadowBias; float shadowMapFar;
	};
	layout(binding = 0) uniform EnvironmentUniform {
		layout(offset = 384) Light lights[4];
		layout(offset = 640) mat3 environmentRotation;
		layout(offset = 688) vec3 cameraPosition;
		layout(offset = 700) float environmentIntensity;
		layout(offset = 704) int mipCount;
	};
	layout(binding = 1) uniform MaterialUniform {
		mat3 texTransform,normalMapTransform,metallicRoughnessMapTransform,ambientOcclusionMapTransform,emissionMapTransform;
		vec4 baseColorFactor; vec3 emission; vec2 metallicRoughness;
		float ambientOcclusionStrength; float alphaThreshold;
		bool unlit; bool enableAlpha;
	};
	layout(binding = 2) uniform UniformBufferObject {
		layout(offset = 192) float meshOutline;
		layout(offset = 196) float gray;
		layout(offset = 200) float hue;
		layout(offset = 208) vec3 add;
		layout(offset = 224) vec3 mult;
		layout(offset = 240) int neg;
	};
	layout (constant_id = 6) const bool useTexture = false;
	layout (constant_id = 7) const bool useNormalMap = false;
	layout (constant_id = 8) const bool useMetallicRoughnessMap = false;
	layout (constant_id = 9) const bool useEmissionMap = false;
	layout (constant_id = 10) const bool useShadowMap = false;
	layout(binding = 5) uniform samplerCube lambertianEnvSampler;
	layout(binding = 6) uniform samplerCube GGXEnvSampler;
	layout(binding = 7) uniform sampler2D GGXLUT;
	layout(binding = 8) uniform sampler2D tex;
	layout(binding = 9) uniform sampler2D normalMap;
	layout(binding = 10) uniform sampler2D metallicRoughnessMap;
	layout(binding = 11) uniform sampler2D ambientOcclusionMap;
	layout(binding = 12) uniform sampler2D emissionMap;
	layout(binding = 13) uniform samplerCubeArray shadowCubeMap;

	layout(location = 0) in vec3 normal;
	layout(location = 1) in vec3 tangent;
	layout(location = 2) in vec3 bitangent;
	layout(location = 3) in vec2 texcoord;
	layout(location = 4) in vec4 vColor;
	layout(location = 5) in vec3 worldSpacePos;
	layout(location = 6) in vec4 lightSpacePos[4];
	layout(location = 0) out vec4 FragColor;
#else
	// OPENGL / GLES PATH
	#ifdef GL_ES
		#extension GL_EXT_texture_cube_map_array : enable
	#else
		#extension GL_ARB_texture_cube_map_array : enable
	#endif
	#define COMPAT_VARYING in
	#define COMPAT_TEXTURE texture
	#define COMPAT_TEXTURE_CUBE texture
	#define COMPAT_TEXTURE_CUBE_LOD textureLod
	#ifdef GL_ES
		precision highp float;
		precision highp int;
	#endif
	#ifdef ENABLE_SHADOW
		#ifdef GL_ES
		// Avoid sampler-array dynamic indexing on GLES by declaring 4 separate samplers
		uniform samplerCube shadowCubeMap0;
		uniform samplerCube shadowCubeMap1;
		uniform samplerCube shadowCubeMap2;
		uniform samplerCube shadowCubeMap3;
		#else
		uniform samplerCubeArray shadowCubeMap;
		#define COMPAT_SHADOW_MAP_TEXTURE() texture(shadowCubeMap,vec4(1.0, -(xy.y*2.0-1.0),-(xy.x*2.0-1.0),index)).r
		#define COMPAT_SHADOW_CUBE_MAP_TEXTURE() texture(shadowCubeMap,vec4(xyz,index)).r
		#endif
		const bool useShadowMap = true;
	#else
		const bool useShadowMap = false;
	#endif
	out vec4 FragColor;

	struct Light {
		vec3 direction; float range;
		vec3 color; float intensity;
		vec3 position; float innerConeCos;
		float outerConeCos; int type;
		float shadowBias; float shadowMapFar;
	};

	uniform sampler2D tex, normalMap, metallicRoughnessMap, ambientOcclusionMap, emissionMap, GGXLUT;
	uniform samplerCube lambertianEnvSampler, GGXEnvSampler;
	uniform mat3 texTransform, normalMapTransform, metallicRoughnessMapTransform, ambientOcclusionMapTransform, emissionMapTransform;
	uniform float environmentIntensity, gray, hue, alphaThreshold, meshOutline, ambientOcclusionStrength;
	uniform mat3 environmentRotation;
	uniform int mipCount, neg;
	uniform vec3 cameraPosition, add, mult, emission;
	uniform vec4 baseColorFactor;
	uniform vec2 metallicRoughness;
	uniform bool unlit, useTexture, useNormalMap, useMetallicRoughnessMap, useEmissionMap, enableAlpha;
	uniform Light lights[4];

	COMPAT_VARYING vec2 texcoord;
	COMPAT_VARYING vec4 vColor;
	COMPAT_VARYING vec3 normal, tangent, bitangent, worldSpacePos;
	COMPAT_VARYING vec4 lightSpacePos[4];
#endif


const float PI = 3.14159265358979;
const int LightType_Directional = 0;
const int LightType_Point = 1;
const int LightType_Spot = 2;

float clampedDot(vec3 x, vec3 y)
{
	return clamp(dot(x, y), 0.0, 1.0);
}

float DirectionalLightShadowCalculation(int index, vec4 lightSpacePos,float NdotL,float shadowBias)
{
	#ifdef ENABLE_SHADOW
	if(!useShadowMap){
		return 1.0;
	}
	// perform perspective divide
	vec3 projCoords = lightSpacePos.xyz / lightSpacePos.w;
	// transform to [0,1] range
	projCoords = projCoords * 0.5 + 0.5;
	// get closest depth value from light's perspective (using [0,1] range fragPosLight as coords)
	float epsilon = 1.0 / 1024.0;
	vec2 xy = vec2(clamp(projCoords.x,epsilon,1.0-epsilon),clamp(projCoords.y,epsilon,1.0-epsilon));
	float closestDepth;
	#if defined(GL_ES)
		// sample using separate samplers with dynamic branch (avoids sampler array indexing)
		vec3 coord0 = vec3(1.0, -(xy.y*2.0-1.0), -(xy.x*2.0-1.0));
		if(index == 0) closestDepth = texture(shadowCubeMap0, coord0).r;
		else if(index == 1) closestDepth = texture(shadowCubeMap1, coord0).r;
		else if(index == 2) closestDepth = texture(shadowCubeMap2, coord0).r;
		else closestDepth = texture(shadowCubeMap3, coord0).r;
	#else
		closestDepth = COMPAT_SHADOW_MAP_TEXTURE();
	#endif
	// get depth of current fragment from light's perspective
	float currentDepth = projCoords.z;
	// check whether current frag pos is in shadow
	float bias = shadowBias*tan(acos(NdotL));
	float shadow = closestDepth-currentDepth > -bias  ? 1.0 : 0.0;
	#else
	float shadow = 1.0;
	#endif
	return shadow;
}

float SpotLightShadowCalculation(int index, vec3 pointToLight, vec4 lightSpacePos,float NdotL,float farPlane,float shadowBias)
{
	#ifdef ENABLE_SHADOW
	if(!useShadowMap){
		return 1.0;
	}
	float epsilon = 1.0 / 1024.0;
	vec2 xy = vec2(clamp(lightSpacePos.x,epsilon,1.0-epsilon),clamp(lightSpacePos.y,epsilon,1.0-epsilon));
	float closestDepth;
	#if defined(GL_ES)
		vec3 coord0 = vec3(1.0, -(xy.y*2.0-1.0), -(xy.x*2.0-1.0));
		if(index == 0) closestDepth = texture(shadowCubeMap0, coord0).r;
		else if(index == 1) closestDepth = texture(shadowCubeMap1, coord0).r;
		else if(index == 2) closestDepth = texture(shadowCubeMap2, coord0).r;
		else closestDepth = texture(shadowCubeMap3, coord0).r;
	#else
		closestDepth = COMPAT_SHADOW_MAP_TEXTURE();
	#endif
	// it is currently in linear range between [0,1]. Re-transform back to original value
	closestDepth *= farPlane;
	// get depth of current fragment from light's perspective
	float currentDepth = length(pointToLight);
	float bias = shadowBias*tan(acos(NdotL));
	float shadow = currentDepth-closestDepth < bias  ? 1.0 : 0.0;
	#else
	float shadow = 1.0;
	#endif
	return shadow;
}

float PointLightShadowCalculation(int index, vec3 pointToLight,float NdotL,float farPlane,float shadowBias)
{
	#ifdef ENABLE_SHADOW
	if(!useShadowMap){
		return 1.0;
	}
	vec3 xyz = -pointToLight;
	float closestDepth;
	#if defined(GL_ES)
		if(index == 0) closestDepth = texture(shadowCubeMap0, xyz).r;
		else if(index == 1) closestDepth = texture(shadowCubeMap1, xyz).r;
		else if(index == 2) closestDepth = texture(shadowCubeMap2, xyz).r;
		else closestDepth = texture(shadowCubeMap3, xyz).r;
	#else
		closestDepth = COMPAT_SHADOW_CUBE_MAP_TEXTURE();
	#endif
	// it is currently in linear range between [0,1]. Re-transform back to original value
	closestDepth *= farPlane;
	// now get current linear depth as the length between the fragment and light position
	float currentDepth = length(pointToLight);

	float bias = shadowBias*tan(acos(NdotL));

	float shadow = currentDepth-closestDepth < bias  ? 1.0 : 0.0;
	
	#else
	float shadow = 1.0;
	#endif
	return shadow;
}

vec3 getNormal()
{
	vec2 uv_dx = dFdx(texcoord);
	vec2 uv_dy = dFdy(texcoord);
	if (length(uv_dx) <= 1e-2) {
	  uv_dx = vec2(1.0, 0.0);
	}

	if (length(uv_dy) <= 1e-2) {
	  uv_dy = vec2(0.0, 1.0);
	}
	vec3 t_ = (uv_dy.y * dFdx(worldSpacePos) - uv_dx.y * dFdy(worldSpacePos)) /
		(uv_dx.x * uv_dy.y - uv_dy.x * uv_dx.y);
	vec3 n, t, b, ng;
	if(normal.x+normal.y+normal.z != 0.0){
		if(tangent.x+tangent.y+tangent.z != 0.0){
			t = normalize(tangent);
			b = normalize(bitangent);
			ng = normalize(normal);
		}else{
			ng = normalize(normal);
			t = normalize(t_ - ng * dot(ng, t_));
			b = cross(ng, t);
		}
	}else{
		ng = normalize(cross(dFdx(worldSpacePos), dFdy(worldSpacePos)));
		t = normalize(t_ - ng * dot(ng, t_));
		b = cross(ng, t);
	}
	if (gl_FrontFacing == false)
	{
		t *= -1.0;
		b *= -1.0;
		ng *= -1.0;
	}
	if(useNormalMap){
		return normalize(mat3(t, b, ng) * normalize(COMPAT_TEXTURE(normalMap, vec2(normalMapTransform*vec3(texcoord,1))).xyz * 2.0 - vec3(1.0)));
	}else{
		return ng;
	}
}

// https://github.com/KhronosGroup/glTF/blob/master/extensions/2.0/Khronos/KHR_lights_punctual/README.md#range-property
float getRangeAttenuation(float range, float distance)
{
	if (range <= 0.0)
	{
		// negative range means unlimited
		return 1.0 / pow(distance, 2.0);
	}
	return max(min(1.0 - pow(distance / range, 4.0), 1.0), 0.0) / pow(distance, 2.0);
}


// https://github.com/KhronosGroup/glTF/blob/master/extensions/2.0/Khronos/KHR_lights_punctual/README.md#inner-and-outer-cone-angles
float getSpotAttenuation(vec3 pointToLight, vec3 spotDirection, float outerConeCos, float innerConeCos)
{
	float actualCos = dot(normalize(spotDirection), normalize(-pointToLight));
	if (actualCos > outerConeCos)
	{
		if (actualCos < innerConeCos)
		{
			float angularAttenuation = (actualCos - outerConeCos) / (innerConeCos - outerConeCos);
			return angularAttenuation * angularAttenuation;
		}
		return 1.0;
	}
	return 0.0;
}
vec3 getLighIntensity(Light light, vec3 pointToLight)
{
	float rangeAttenuation = 1.0;
	float spotAttenuation = 1.0;

	if (light.type != LightType_Directional)
	{
		rangeAttenuation = getRangeAttenuation(light.range, length(pointToLight));
	}
	if (light.type == LightType_Spot)
	{
		spotAttenuation = getSpotAttenuation(pointToLight, light.direction, light.outerConeCos, light.innerConeCos);
	}

	return rangeAttenuation * spotAttenuation * light.intensity * light.color;
}
vec3 F_Schlick(vec3 f0, vec3 f90, float VdotH)
{
	return f0 + (f90 - f0) * pow(clamp(1.0 - VdotH, 0.0, 1.0), 5.0);
}
// Smith Joint GGX
// Note: Vis = G / (4 * NdotL * NdotV)
// see Eric Heitz. 2014. Understanding the Masking-Shadowing Function in Microfacet-Based BRDFs. Journal of Computer Graphics Techniques, 3
// see Real-Time Rendering. Page 331 to 336.
// see https://google.github.io/filament/Filament.md.html#materialsystem/specularbrdf/geometricshadowing(specularg)
float V_GGX(float NdotL, float NdotV, float alphaRoughness)
{
	float alphaRoughnessSq = alphaRoughness * alphaRoughness;

	float GGXV = NdotL * sqrt(NdotV * NdotV * (1.0 - alphaRoughnessSq) + alphaRoughnessSq);
	float GGXL = NdotV * sqrt(NdotL * NdotL * (1.0 - alphaRoughnessSq) + alphaRoughnessSq);

	float GGX = GGXV + GGXL;
	if (GGX > 0.0)
	{
		return 0.5 / GGX;
	}
	return 0.0;
}

// The following equation(s) model the distribution of microfacet normals across the area being drawn (aka D())
// Implementation from "Average Irregularity Representation of a Roughened Surface for Ray Reflection" by T. S. Trowbridge, and K. P. Reitz
// Follows the distribution function recommended in the SIGGRAPH 2013 course notes from EPIC Games [1], Equation 3.
float D_GGX(float NdotH, float alphaRoughness)
{
	float alphaRoughnessSq = alphaRoughness * alphaRoughness;
	float f = (NdotH * NdotH) * (alphaRoughnessSq - 1.0) + 1.0;
	return alphaRoughnessSq / (PI * f * f);
}
vec3 BRDF_lambertian(vec3 f0, vec3 f90, vec3 diffuseColor, float specularWeight, float VdotH)
{
	// see https://seblagarde.wordpress.com/2012/01/08/pi-or-not-to-pi-in-game-lighting-equation/
	return (1.0 - specularWeight * F_Schlick(f0, f90, VdotH)) * (diffuseColor / PI);
}
vec3 BRDF_specularGGX(vec3 f0, vec3 f90, float alphaRoughness, float specularWeight, float VdotH, float NdotL, float NdotV, float NdotH)
{
	vec3 F = F_Schlick(f0, f90, VdotH);
	float Vis = V_GGX(NdotL, NdotV, alphaRoughness);
	float D = D_GGX(NdotH, alphaRoughness);

	return specularWeight * F * Vis * D;
}
vec3 getDiffuseLight(vec3 n)
{
	return COMPAT_TEXTURE_CUBE(lambertianEnvSampler, environmentRotation * n).rgb * environmentIntensity;
}
vec4 getSpecularSample(vec3 reflection, float lod)
{
	return COMPAT_TEXTURE_CUBE_LOD(GGXEnvSampler, environmentRotation * reflection, lod) * environmentIntensity;
}

// --- Forward declarations
vec3 getLighIntensity(Light light, vec3 pointToLight);
vec3 BRDF_lambertian(vec3 f0, vec3 f90, vec3 diffuseColor, float specularWeight, float VdotH);
vec3 BRDF_specularGGX(vec3 f0, vec3 f90, float alphaRoughness, float specularWeight, float VdotH, float NdotL, float NdotV, float NdotH);
// --- end forward declarations

vec3 getIBLGGXFresnel(vec3 n, vec3 v, float roughness, vec3 F0, float specularWeight)
{
	// see https://bruop.github.io/ibl/#single_scattering_results at Single Scattering Results
	// Roughness dependent fresnel, from Fdez-Aguera
	float NdotV = clampedDot(n, v);
	vec2 brdfSamplePoint = clamp(vec2(NdotV, roughness), vec2(0.0, 0.0), vec2(1.0, 1.0));
	vec2 f_ab = COMPAT_TEXTURE(GGXLUT, brdfSamplePoint).rg;
	vec3 Fr = max(vec3(1.0 - roughness), F0) - F0;
	vec3 k_S = F0 + Fr * pow(1.0 - NdotV, 5.0);
	vec3 FssEss = specularWeight * (k_S * f_ab.x + f_ab.y);

	// Multiple scattering, from Fdez-Aguera
	float Ems = (1.0 - (f_ab.x + f_ab.y));
	vec3 F_avg = specularWeight * (F0 + (1.0 - F0) / 21.0);
	vec3 FmsEms = Ems * FssEss * F_avg / (1.0 - F_avg * Ems);

	return FssEss + FmsEms;
}
vec3 getIBLRadianceGGX(vec3 n, vec3 v, float roughness)
{
	float NdotV = clampedDot(n, v);
	float lod = roughness * float(mipCount - 1);
	vec3 reflection = normalize(reflect(-v, n));
	vec4 specularSample = getSpecularSample(reflection, lod);

	vec3 specularLight = specularSample.rgb;

	return specularLight;
}
vec3 ibl(vec3 n,vec3 v,float metallic,float roughness,vec3 albedo){
	vec3 f_diffuse = getDiffuseLight(n) * albedo.rgb;
	vec3 f_specular_metal = getIBLRadianceGGX(n, v, roughness);
	vec3 f_specular_dielectric = f_specular_metal;
	vec3 f_metal_fresnel_ibl = getIBLGGXFresnel(n, v, roughness, albedo.rgb, 1.0);
	vec3 f_metal_brdf_ibl = f_metal_fresnel_ibl * f_specular_metal;
	vec3 f_dielectric_fresnel_ibl = getIBLGGXFresnel(n, v, roughness, vec3(0.04), 1.0);
	vec3 f_dielectric_brdf_ibl = f_diffuse*(1.0-f_dielectric_fresnel_ibl) + f_specular_dielectric * f_dielectric_fresnel_ibl;

	vec3 color = f_dielectric_brdf_ibl*(1.0-metallic)+f_metal_brdf_ibl*metallic;
	return color;
}
vec3 pbr(vec3 worldSpacePos,vec3 v,vec3 n,vec3 albedo,float metallic,float roughness,float ao){
	vec3 f0 = vec3(0.04)+(albedo-vec3(0.04))*metallic;
	vec3 f90 = vec3(1.0);
	float ior = 1.5;
	float specularWeight = 1.0;
	vec3 f_specular = vec3(0.0);
	vec3 f_diffuse = vec3(0.0);
	vec3 c_diff = albedo*(1.0-metallic);

	for(int i = 0; i < 4; ++i) 
	{
		if(lights[i].color.r+lights[i].color.g+lights[i].color.b > 0.0){
			vec3 pointToLight = vec3(0.0);
			if(lights[i].type == LightType_Directional){
				pointToLight = -lights[i].direction;
			}else{
				pointToLight = lights[i].position - worldSpacePos;
			}
			vec3 l = normalize(pointToLight);
			vec3 h = normalize(l + v);
			float NdotL = clampedDot(n, l);
			float NdotV = clampedDot(n, v);
			float NdotH = clampedDot(n, h);
			float VdotH = clampedDot(v, h);
			if (NdotL > 0.0 || NdotV > 0.0){
				vec3 intensity = getLighIntensity(lights[i], pointToLight);
				vec3 l_diffuse = vec3(0.0);
				vec3 l_specular = vec3(0.0);
				l_diffuse += intensity * NdotL *  BRDF_lambertian(f0, f90, c_diff, specularWeight, VdotH);
				l_specular += intensity * NdotL * BRDF_specularGGX(f0, f90, roughness*roughness, specularWeight, VdotH, NdotL, NdotV, NdotH);
				float shadow = 1.0;
				if(lights[i].type == LightType_Point){
					shadow = PointLightShadowCalculation(i,pointToLight,NdotL,lights[i].shadowMapFar,lights[i].shadowBias);
				}else if(lights[i].type == LightType_Directional){
					shadow = DirectionalLightShadowCalculation(i,lightSpacePos[i],NdotL,lights[i].shadowBias);
				}else{
					shadow = SpotLightShadowCalculation(i,pointToLight,lightSpacePos[i],NdotL,lights[i].shadowMapFar,lights[i].shadowBias);
				}
				f_diffuse += l_diffuse * shadow;
				f_specular += l_specular * shadow;
			}
		}
	}   
	vec3 f_ibl = vec3(0.0);
	if(environmentIntensity > 0.0){
		f_ibl = ibl(n,v,metallic,roughness,albedo);
	}
	//vec3 color = clamp(f_diffuse+f_specular+ao*f_ibl,0,1);
	vec3 color = f_diffuse+f_specular+ao*f_ibl;
	
	//color = color / (color + vec3(1.0));
	//color = pow(color, vec3(1.0/2.2));
	return color;
}

// Hue shift using unrolled matrix for GLES stability
vec3 hue_shift(vec3 color, float dhue) {
	float s = sin(dhue);
	float c = cos(dhue);
	vec3 row1 = vec3(0.167444, 0.329213, -0.496657);
	vec3 row2 = vec3(-0.327948, 0.035669, 0.292279);
	vec3 row3 = vec3(1.250268, -1.047561, -0.202707);
	return (color * c) + (color * s) * vec3(dot(row1, color), dot(row2, color), dot(row3, color)) + dot(vec3(0.299, 0.587, 0.114), color) * (1.0 - c);
}

void main(void) {
	FragColor = vec4(1.0);
	if(useTexture){
		FragColor = COMPAT_TEXTURE(tex, vec2(texTransform*vec3(texcoord,1)));
		// convert texture to linear space
		FragColor.rgb = pow(FragColor.rgb,vec3(2.2));
	}
	FragColor *= baseColorFactor;
	FragColor *= vColor;
	if(meshOutline > 0.0) {
		FragColor.rgb = vec3(0.0,0.0,0.0);
	}else if(!unlit){
		vec3 normalF = normal;
		if(useNormalMap){
			normalF = getNormal();
		}
		vec2 metallicRoughnessF = metallicRoughness;
		if(useMetallicRoughnessMap){
			metallicRoughnessF = COMPAT_TEXTURE(metallicRoughnessMap, vec2(metallicRoughnessMapTransform*vec3(texcoord,1.0))).bg;
		}
		float ambientOcclusion = 1.0;
		if(ambientOcclusionStrength > 0.0){
			ambientOcclusion = 1.0+ambientOcclusionStrength*(COMPAT_TEXTURE(ambientOcclusionMap, vec2(ambientOcclusionMapTransform*vec3(texcoord,1.0))).r-1.0);
		}
		// PBR returns color in linear space
		FragColor.rgb = pbr(worldSpacePos,normalize(cameraPosition - worldSpacePos),normalize(normalF),FragColor.rgb,metallicRoughnessF[0],metallicRoughnessF[1],ambientOcclusion);
		
		// Emission is also added in linear space
		if(useEmissionMap){
			FragColor.rgb += emission * pow(COMPAT_TEXTURE(emissionMap, vec2(emissionMapTransform*vec3(texcoord,1.0))).rgb,vec3(2.2));
		}else{
			FragColor.rgb += emission;
		}
	}
	
	// PBR output (FragColor.rgb) is now LINEAR, pre-multiplied by vColor.a here.
	FragColor.rgb *= vColor.a;

	if(!enableAlpha){
		if(FragColor.a < alphaThreshold){
			discard;
		}else{
			FragColor.a = 1.0;
		}
	}else if(FragColor.a<=0.0){
		discard;
	}

	// Final Gamma Correction (Required for Vulkan path rendering to an sRGB target)
	// This happens *after* all linear operations and re-premultiplication.
	FragColor.rgb = pow(FragColor.rgb, vec3(1.0/2.2));

	vec3 c_linear = FragColor.rgb;
	float alpha = FragColor.a;

	// Un-premultiply to get True Linear Color
	if (alpha > 0.0) {
		c_linear /= alpha;
		c_linear = clamp(c_linear, 0.0, 1.0);
	}
	
	// Convert ADD uniform from sRGB space (assuming user input) to Linear
	// We assume 'add' is defined in sRGB space.
	// vec3 linear_add = sign(add) * pow(abs(add), vec3(2.2));

	// Apply PalFX (All math in True Linear space)
	if (hue != 0.0) {
		c_linear = hue_shift(c_linear, hue);           
	}
	
	// INVERSION FIX: Correctly applied on linear, un-premultiplied color
	if (neg != 0) {
		c_linear = vec3(1.0) - c_linear; 
	}
	
	// Grayscale / Add / Mult
	c_linear = mix(c_linear, vec3((c_linear.r + c_linear.g + c_linear.b) / 3.0), gray) + add;
	c_linear *= mult;
	c_linear = clamp(c_linear, 0.0, 1.0);

	// Re-premultiply alpha
	FragColor.rgb = c_linear * alpha;
}