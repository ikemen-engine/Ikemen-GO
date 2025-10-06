#if __VERSION__ >= 130
#define COMPAT_VARYING out
#define COMPAT_ATTRIBUTE in
#define COMPAT_TEXTURE texture
#else
#extension GL_EXT_gpu_shader4 : enable
#define COMPAT_VARYING varying 
#define COMPAT_ATTRIBUTE attribute 
#define COMPAT_TEXTURE texture2D
#endif

uniform mat4 model, view, projection;
uniform mat4 normalMatrix;
uniform mat4 outlineMatrix;
uniform mat4 lightMatrices[4];
uniform sampler2D jointMatrices;
//uniform highp sampler2D morphTargetValues;
uniform sampler2D morphTargetValues;
uniform int numJoints;
uniform int numTargets;
uniform int morphTargetTextureDimension;
uniform vec4 morphTargetWeight[2];
uniform vec4 morphTargetOffset;
uniform int numVertices;
uniform float meshOutline;
uniform vec3 cameraPosition;
//gl_VertexID is not available in 1.2
COMPAT_ATTRIBUTE float vertexId;
COMPAT_ATTRIBUTE vec3 position;
COMPAT_ATTRIBUTE vec3 normalIn;
COMPAT_ATTRIBUTE vec4 tangentIn;
COMPAT_ATTRIBUTE vec2 uv;
COMPAT_ATTRIBUTE vec4 vertColor;
COMPAT_ATTRIBUTE vec4 joints_0;
COMPAT_ATTRIBUTE vec4 joints_1;
COMPAT_ATTRIBUTE vec4 weights_0;
COMPAT_ATTRIBUTE vec4 weights_1;
COMPAT_ATTRIBUTE vec4 outlineAttribute;
COMPAT_VARYING vec3 normal;
COMPAT_VARYING vec3 tangent;
COMPAT_VARYING vec3 bitangent;
COMPAT_VARYING vec2 texcoord;
COMPAT_VARYING vec4 vColor;
COMPAT_VARYING vec3 worldSpacePos;
COMPAT_VARYING vec4 lightSpacePos[4];


mat4 getMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = COMPAT_TEXTURE(jointMatrices,vec2(0.5/6.0,(index+0.5)/numJoints));
	mat[1] = COMPAT_TEXTURE(jointMatrices,vec2(1.5/6.0,(index+0.5)/numJoints));
	mat[2] = COMPAT_TEXTURE(jointMatrices,vec2(2.5/6.0,(index+0.5)/numJoints));
	mat[3] = vec4(0,0,0,1);
	return transpose(mat);
}
mat4 getNormalMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = COMPAT_TEXTURE(jointMatrices,vec2(3.5/6.0,(index+0.5)/numJoints));
	mat[1] = COMPAT_TEXTURE(jointMatrices,vec2(4.5/6.0,(index+0.5)/numJoints));
	mat[2] = COMPAT_TEXTURE(jointMatrices,vec2(5.5/6.0,(index+0.5)/numJoints));
	mat[3] = vec4(0,0,0,1);
	return transpose(mat);
}
mat4 getJointMatrix(){
	mat4 ret = mat4(0);
	ret += weights_0.x*getMatrixFromTexture(joints_0.x);
	ret += weights_0.y*getMatrixFromTexture(joints_0.y);
	ret += weights_0.z*getMatrixFromTexture(joints_0.z);
	ret += weights_0.w*getMatrixFromTexture(joints_0.w);
	ret += weights_1.x*getMatrixFromTexture(joints_1.x);
	ret += weights_1.y*getMatrixFromTexture(joints_1.y);
	ret += weights_1.z*getMatrixFromTexture(joints_1.z);
	ret += weights_1.w*getMatrixFromTexture(joints_1.w);
	if(ret == mat4(0.0)){
		return mat4(1.0);
	}
	return ret;
}
mat3 getJointNormalMatrix(){
	mat4 ret = mat4(0);
	ret += weights_0.x*getNormalMatrixFromTexture(joints_0.x);
	ret += weights_0.y*getNormalMatrixFromTexture(joints_0.y);
	ret += weights_0.z*getNormalMatrixFromTexture(joints_0.z);
	ret += weights_0.w*getNormalMatrixFromTexture(joints_0.w);
	ret += weights_1.x*getNormalMatrixFromTexture(joints_1.x);
	ret += weights_1.y*getNormalMatrixFromTexture(joints_1.y);
	ret += weights_1.z*getNormalMatrixFromTexture(joints_1.z);
	ret += weights_1.w*getNormalMatrixFromTexture(joints_1.w);
	if(ret == mat4(0.0)){
		return mat3(1.0);
	}
	return mat3(ret);
}
void main(void) {
	texcoord = uv;
	vColor = vertColor;
	vec4 pos = vec4(position, 1.0);
	normal = normalIn;
	tangent = vec3(tangentIn);
	if(morphTargetWeight[0][0] != 0){
		for(int idx = 0; idx < numTargets; ++idx)
		{
			float i = idx*numVertices+vertexId;
			vec2 xy = vec2((i+0.5)/morphTargetTextureDimension-floor(i/morphTargetTextureDimension),(floor(i/morphTargetTextureDimension)+0.5)/morphTargetTextureDimension);
			if(idx < morphTargetOffset[0]){
				pos += morphTargetWeight[idx/4][idx%4] * COMPAT_TEXTURE(morphTargetValues,xy);
			}else if(idx < morphTargetOffset[1]){
				normal += morphTargetWeight[idx/4][idx%4] * vec3(COMPAT_TEXTURE(morphTargetValues,xy));
			}else if(idx < morphTargetOffset[2]){
				tangent += morphTargetWeight[idx/4][idx%4] * vec3(COMPAT_TEXTURE(morphTargetValues,xy));
			}else if(idx < morphTargetOffset[3]){
				texcoord += morphTargetWeight[idx/4][idx%4] * vec2(COMPAT_TEXTURE(morphTargetValues,xy));
			}else{
				vColor += morphTargetWeight[idx/4][idx%4] * COMPAT_TEXTURE(morphTargetValues,xy);
			}
		}
	}
	if(weights_0.x+weights_0.y+weights_0.z+weights_0.w+weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0){
		
		mat4 jointMatrix = getJointMatrix();
		mat3 jointNormalMatrix = getJointNormalMatrix();
		normal = mat3(normalMatrix) * jointNormalMatrix * normal;
		vec4 tmp2 = model * jointMatrix * pos;
		
		if(outlineAttribute.w > 0){
			vec3 p = normalize(mat3(normalMatrix) * outlineAttribute.xyz)*outlineAttribute.w*meshOutline*length(cameraPosition-tmp2.xyz);
			tmp2.xyz += p;
		}else{
			vec3 p = normal*meshOutline*length(cameraPosition-tmp2.xyz);
			tmp2.xyz += p;
		}

		gl_Position = projection * view * tmp2;
		worldSpacePos = vec3(tmp2);
		for(int i = 0;i < 4;i++){
			lightSpacePos[i] = lightMatrices[i] * tmp2;
		}
	}else{
		normal = normalize(mat3(normalMatrix) * normal);
		if(tangent.x+tangent.y+tangent.z != 0){
			tangent = normalize(vec3(model * vec4(tangent,0)));
			bitangent = cross(normal, tangent) * tangentIn.w;
		}
		vec4 tmp2 = model * pos;

		if(outlineAttribute.w > 0){
			vec3 p = normalize(mat3(normalMatrix) * outlineAttribute.xyz)*outlineAttribute.w*meshOutline*length(cameraPosition-tmp2.xyz);
			tmp2.xyz += p;
		}else{
			vec3 p = normal*meshOutline*length(cameraPosition-tmp2.xyz);
			tmp2.xyz += p;
		}

		gl_Position = projection * view * tmp2;
		worldSpacePos = vec3(tmp2);
		for(int i = 0;i < 4;i++){
			lightSpacePos[i] = lightMatrices[i] * tmp2;
		}
	}
}