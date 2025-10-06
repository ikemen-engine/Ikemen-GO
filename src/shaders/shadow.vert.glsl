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


uniform mat4 model;
uniform sampler2D jointMatrices;
//uniform highp sampler2D morphTargetValues;
uniform sampler2D morphTargetValues;
uniform int morphTargetTextureDimension;
uniform int numJoints;
uniform int numTargets;
uniform vec4 morphTargetWeight[2];
uniform vec4 morphTargetOffset;
uniform int numVertices;

//gl_VertexID is not available in 1.2
COMPAT_ATTRIBUTE float vertexId;
COMPAT_ATTRIBUTE vec3 position;
COMPAT_ATTRIBUTE vec4 vertColor;
COMPAT_ATTRIBUTE vec2 uv;
COMPAT_ATTRIBUTE vec4 joints_0;
COMPAT_ATTRIBUTE vec4 joints_1;
COMPAT_ATTRIBUTE vec4 weights_0;
COMPAT_ATTRIBUTE vec4 weights_1;
COMPAT_VARYING vec4 vColor;
COMPAT_VARYING vec2 texcoord;


mat4 getMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = COMPAT_TEXTURE(jointMatrices,vec2(0.5/6.0,(index+0.5)/numJoints));
	mat[1] = COMPAT_TEXTURE(jointMatrices,vec2(1.5/6.0,(index+0.5)/numJoints));
	mat[2] = COMPAT_TEXTURE(jointMatrices,vec2(2.5/6.0,(index+0.5)/numJoints));
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
void main(void) {
	texcoord = uv;
	vColor = vertColor;
	vec4 pos = vec4(position, 1.0);
	if(morphTargetOffset[0] > 0){
		for(int idx = 0; idx < numTargets; ++idx)
		{
			float i = idx*numVertices+vertexId;
			vec2 xy = vec2((i+0.5)/morphTargetTextureDimension-floor(i/morphTargetTextureDimension),(floor(i/morphTargetTextureDimension)+0.5)/morphTargetTextureDimension);
			if(idx < morphTargetOffset[0]){
				pos += morphTargetWeight[idx/4][idx%4] * COMPAT_TEXTURE(morphTargetValues,xy);
			}else if(idx >= morphTargetOffset[2] && idx < morphTargetOffset[3]){
				texcoord += morphTargetWeight[idx/4][idx%4] * vec2(COMPAT_TEXTURE(morphTargetValues,xy));
			}
		}
	}
	if(weights_0.x+weights_0.y+weights_0.z+weights_0.w+weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0){
		
		mat4 jointMatrix = getJointMatrix();
		gl_Position = model * jointMatrix * pos;
	}else{
		gl_Position = model * pos;
	}
}