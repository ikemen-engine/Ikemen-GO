#if __VERSION__ >= 450
#define COMPAT_TEXTURE texture
layout(binding = 0) uniform UniformBufferObject0 {
	mat4 view, projection;
	mat4 lightMatrices[4];
	layout(offset = 688) vec3 cameraPosition;
};

layout(binding = 2) uniform UniformBufferObject2 {
	mat4 model,normalMatrix;
	int numJoints,numTargets,morphTargetTextureDimension,numVertices;
	vec4 morphTargetWeight[2];
	vec4 morphTargetOffset;
	float meshOutline;
};

layout(binding = 3) uniform sampler2D jointMatrices;
layout(binding = 4) uniform sampler2D morphTargetValues;

layout (constant_id = 0) const bool useJoint0 = false;
layout (constant_id = 1) const bool useJoint1 = false;
layout (constant_id = 2) const bool useNormal = false;
layout (constant_id = 3) const bool useTangent = false;
layout (constant_id = 4) const bool useVertColor = false;
layout (constant_id = 5) const bool useOutlineAttribute = false;

layout(location = 0) in int inVertexId;
layout(location = 1) in vec3 position;
layout(location = 2) in vec2 uv;
layout(location = 3) in vec3 normalIn;
layout(location = 4) in vec4 tangentIn;
layout(location = 5) in vec4 vertColor;
layout(location = 6) in vec4 joints_0;
layout(location = 7) in vec4 weights_0;
layout(location = 8) in vec4 joints_1;
layout(location = 9) in vec4 weights_1;
layout(location = 10) in vec4 outlineAttributeIn;

layout(location = 0) out vec3 normal;
layout(location = 1) out vec3 tangent;
layout(location = 2) out vec3 bitangent;
layout(location = 3) out vec2 texcoord;
layout(location = 4) out vec4 vColor;
layout(location = 5) out vec3 worldSpacePos;
layout(location = 6) out vec4 lightSpacePos[4];
#else
	// GLES 3.2 / ANDROID PATH - Standard Uniforms
	#if __VERSION__ >= 130 || defined(GL_ES)
		#define COMPAT_VARYING out
		#define COMPAT_ATTRIBUTE in
		#define COMPAT_TEXTURE texture
		#ifdef GL_ES
			precision highp float;
			precision highp int;
			precision highp sampler2D;
		#endif
	#else
		#define COMPAT_VARYING varying 
		#define COMPAT_ATTRIBUTE attribute 
		#define COMPAT_TEXTURE texture2D
	#endif

	uniform mat4 model, view, projection, normalMatrix;
	uniform mat4 lightMatrices[4];
	uniform sampler2D jointMatrices, morphTargetValues;
	uniform int numJoints, numTargets, morphTargetTextureDimension, numVertices;
	uniform vec4 morphTargetWeight[2], morphTargetOffset;
	uniform float meshOutline;
	uniform vec3 cameraPosition;

	// Use float to match standard gl.VertexAttribPointer from Go
	COMPAT_ATTRIBUTE float inVertexId; 
	COMPAT_ATTRIBUTE vec3 position, normalIn;
	COMPAT_ATTRIBUTE vec4 tangentIn, vertColor, joints_0, joints_1, weights_0, weights_1, outlineAttributeIn;
	COMPAT_ATTRIBUTE vec2 uv;

	COMPAT_VARYING vec3 normal, tangent, bitangent, worldSpacePos;
	COMPAT_VARYING vec2 texcoord;
	COMPAT_VARYING vec4 vColor, lightSpacePos[4];

	#define useJoint0 (weights_0.x+weights_0.y+weights_0.z+weights_0.w+weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0.001)
	#define useJoint1 (weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0.001)
	#define useNormal true
	#define useTangent true
	#define useVertColor true
	#define useOutlineAttribute true
#endif

mat4 getMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = COMPAT_TEXTURE(jointMatrices,vec2(0.5/6.0,(index+0.5)/float(numJoints)));
	mat[1] = COMPAT_TEXTURE(jointMatrices,vec2(1.5/6.0,(index+0.5)/float(numJoints)));
	mat[2] = COMPAT_TEXTURE(jointMatrices,vec2(2.5/6.0,(index+0.5)/float(numJoints)));
	mat[3] = vec4(0.0,0.0,0.0,1.0);
	return transpose(mat);
}
mat4 getNormalMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = COMPAT_TEXTURE(jointMatrices,vec2(3.5/6.0,(index+0.5)/float(numJoints)));
	mat[1] = COMPAT_TEXTURE(jointMatrices,vec2(4.5/6.0,(index+0.5)/float(numJoints)));
	mat[2] = COMPAT_TEXTURE(jointMatrices,vec2(5.5/6.0,(index+0.5)/float(numJoints)));
	mat[3] = vec4(0.0,0.0,0.0,1.0);
	return transpose(mat);
}
mat4 getJointMatrix(){
	mat4 ret = mat4(0);
	ret += weights_0.x*getMatrixFromTexture(joints_0.x);
	ret += weights_0.y*getMatrixFromTexture(joints_0.y);
	ret += weights_0.z*getMatrixFromTexture(joints_0.z);
	ret += weights_0.w*getMatrixFromTexture(joints_0.w);
	if(useJoint1){
		ret += weights_1.x*getMatrixFromTexture(joints_1.x);
		ret += weights_1.y*getMatrixFromTexture(joints_1.y);
		ret += weights_1.z*getMatrixFromTexture(joints_1.z);
		ret += weights_1.w*getMatrixFromTexture(joints_1.w);
	}
	if(ret == mat4(0.0)){
		return mat4(1.0);
	}
	return ret;
}
mat3 getJointNormalMatrix(){
	mat4 ret = mat4(0);
	vec4 w1 = useJoint1?weights_1:vec4(0);
	ret += weights_0.x*getNormalMatrixFromTexture(joints_0.x);
	ret += weights_0.y*getNormalMatrixFromTexture(joints_0.y);
	ret += weights_0.z*getNormalMatrixFromTexture(joints_0.z);
	ret += weights_0.w*getNormalMatrixFromTexture(joints_0.w);
	ret += w1.x*getNormalMatrixFromTexture(joints_1.x);
	ret += w1.y*getNormalMatrixFromTexture(joints_1.y);
	ret += w1.z*getNormalMatrixFromTexture(joints_1.z);
	ret += w1.w*getNormalMatrixFromTexture(joints_1.w);
	if(ret == mat4(0.0)){
		return mat3(1.0);
	}
	return mat3(ret);
}
void main(void) {
	// Initialize the crap so the ES driver STFU's
	normal = vec3(0.0);
	tangent = vec3(0.0);
	bitangent = vec3(0.0);
	worldSpacePos = vec3(0.0);
	texcoord = vec2(0.0);
	vColor = vec4(1.0);
	for(int i=0; i<4; i++) lightSpacePos[i] = vec4(0.0);

	texcoord = uv;
	vColor = useVertColor?vertColor:vec4(1.0,1.0,1.0,1.0);
	vec4 pos = vec4(position, 1.0);
	normal = useNormal?normalIn:vec3(0.0,0.0,0.0);
	tangent = useTangent?vec3(tangentIn):vec3(0.0,0.0,0.0);
	vec4 outlineAttribute = useOutlineAttribute?outlineAttributeIn:vec4(0);
	if(morphTargetWeight[0][0] != 0.0){
		for(int idx = 0; idx < numTargets; ++idx)
		{
			float fIdx = float(idx);
			float i = fIdx * float(numVertices) + inVertexId;
			vec2 xy = vec2((i+0.5)/float(morphTargetTextureDimension)-floor(i/float(morphTargetTextureDimension)),(floor(i/float(morphTargetTextureDimension))+0.5)/float(morphTargetTextureDimension));

			// Mali-safe weight selection
			vec4 w = (idx < 4) ? morphTargetWeight[0] : morphTargetWeight[1];

			// Need to do this for OpenGL 2.1
			int m = idx - (idx / 4) * 4;
			float weight = (m == 0) ? w.x : (m == 1) ? w.y : (m == 2) ? w.z : w.w;

			vec4 mSample = COMPAT_TEXTURE(morphTargetValues, xy);

			if(fIdx < morphTargetOffset[0]) pos += weight * mSample;
			else if(fIdx < morphTargetOffset[1]) normal += weight * mSample.xyz;
			else if(fIdx < morphTargetOffset[2]) tangent += weight * mSample.xyz;
			else if(fIdx < morphTargetOffset[3]) texcoord += weight * mSample.xy;
			else vColor += weight * mSample;
		}
	}
	if(useJoint0){
		
		mat4 jointMatrix = getJointMatrix();
		mat3 jointNormalMatrix = getJointNormalMatrix();
		normal = mat3(normalMatrix) * jointNormalMatrix * normal;
		vec4 tmp2 = model * jointMatrix * pos;
		
		if(outlineAttribute.w > 0.0){
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
		if(normal.x+normal.y+normal.z != 0.0){
			normal = normalize(mat3(normalMatrix) * normal);
		}
		if(tangent.x+tangent.y+tangent.z != 0.0){
			tangent = normalize(vec3(model * vec4(tangent,0.0)));
			bitangent = cross(normal, tangent) * (useTangent?tangentIn.w:0.0);
		}
		vec4 tmp2 = model * pos;
		if(outlineAttribute.w > 0.0){
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
	#if __VERSION__ >= 450
	gl_Position.y = -gl_Position.y;
	#endif
}