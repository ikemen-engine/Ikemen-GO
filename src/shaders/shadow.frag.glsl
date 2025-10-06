#if __VERSION__ >= 130
#define COMPAT_VARYING in
#define COMPAT_TEXTURE texture
#else
#define COMPAT_VARYING varying
#define COMPAT_TEXTURE texture2D
#endif

uniform sampler2D tex;
uniform mat3 texTransform;
uniform bool enableAlpha;
uniform bool useTexture;
uniform float alphaThreshold;
uniform vec4 baseColorFactor;
uniform vec3 lightPos[4];
uniform int lightType[4];
uniform int lightIndex;
uniform float farPlane[4];
COMPAT_VARYING vec4 FragPos;
COMPAT_VARYING float vColorAlpha;
COMPAT_VARYING vec2 texcoord0;

const int LightType_None = 0;
const int LightType_Directional = 1;
const int LightType_Point = 2;
const int LightType_Spot = 3;
void main()
{
    vec4 color = baseColorFactor;
    if(useTexture){
        color = color * COMPAT_TEXTURE(tex, vec2(texTransform*vec3(texcoord0,1)));
    }
    color.a *= vColorAlpha;
    if((enableAlpha && color.a <= 0) || (color.a < alphaThreshold)){
        discard;
    }
    if(lightType[lightIndex] != LightType_Directional){
        float lightDistance = length(FragPos.xyz - lightPos[lightIndex]);
    
        lightDistance = lightDistance / farPlane[lightIndex];
        
        gl_FragDepth = lightDistance;
    }else{
        gl_FragDepth = gl_FragCoord.z/gl_FragCoord.w;
    }
}