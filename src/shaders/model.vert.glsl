uniform mat4 modelview, projection;
uniform sampler2D jointMatrices;
uniform int numJoints;
uniform vec4 morphTargetWeight[2];
uniform int positionTargetCount;
uniform int uvTargetCount;
attribute vec3 position;
attribute vec2 uv;
attribute vec4 vertColor;
attribute vec4 joints_0;
attribute vec4 joints_1;
attribute vec4 weights_0;
attribute vec4 weights_1;
//Unfortunately the current OpenGL/shader version does not support attribute array
//attribute vec4 morphTargets[8]
attribute vec4 morphTargets_0;
attribute vec4 morphTargets_1;
attribute vec4 morphTargets_2;
attribute vec4 morphTargets_3;
attribute vec4 morphTargets_4;
attribute vec4 morphTargets_5;
attribute vec4 morphTargets_6;
attribute vec4 morphTargets_7;
varying vec2 texcoord;
varying vec4 vColor;


mat4 getMatrixFromTexture(float index){
	mat4 mat;
	mat[0] = texture2D(jointMatrices,vec2(0.5/3.0,(index+0.5)/numJoints));
	mat[1] = texture2D(jointMatrices,vec2(1.5/3.0,(index+0.5)/numJoints));
	mat[2] = texture2D(jointMatrices,vec2(2.5/3.0,(index+0.5)/numJoints));
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
	if(morphTargetWeight[0][0] != 0){
		int idx = 0;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[0][0] * morphTargets_0;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[0][0] * vec2(morphTargets_0);
		}else{
			vColor += morphTargetWeight[0][0] * morphTargets_0;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[0][1] * morphTargets_1;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[0][1] * vec2(morphTargets_1);
		}else{
			vColor += morphTargetWeight[0][1] * morphTargets_1;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[0][2] * morphTargets_2;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[0][2] * vec2(morphTargets_2);
		}else{
			vColor += morphTargetWeight[0][2] * morphTargets_2;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[0][3] * morphTargets_3;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[0][3] * vec2(morphTargets_3);
		}else{
			vColor += morphTargetWeight[0][3] * morphTargets_3;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[1][0] * morphTargets_4;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[1][0] * vec2(morphTargets_4);
		}else{
			vColor += morphTargetWeight[1][0] * morphTargets_4;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[1][1] * morphTargets_5;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[1][1] * vec2(morphTargets_5);
		}else{
			vColor += morphTargetWeight[1][1] * morphTargets_5;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[1][2] * morphTargets_6;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[1][2] * vec2(morphTargets_6);
		}else{
			vColor += morphTargetWeight[1][2] * morphTargets_6;
		}
		idx++;
		if(idx < positionTargetCount){
			pos += morphTargetWeight[1][3] * morphTargets_7;
		}else if(idx - positionTargetCount < uvTargetCount){
			texcoord += morphTargetWeight[1][3] * vec2(morphTargets_7);
		}else{
			vColor += morphTargetWeight[1][3] * morphTargets_7;
		}
		idx++;
	}
	if(weights_0.x+weights_0.y+weights_0.z+weights_0.w+weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0){
		mat4 tmp = getJointMatrix();
		gl_Position = projection * (modelview * tmp * pos);
	}else{
		gl_Position = projection * (modelview * pos);
	}
}