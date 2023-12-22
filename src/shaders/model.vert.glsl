uniform mat4 modelview, projection;
uniform sampler2D jointMatrices;
uniform int numJoints;
attribute vec3 position;
attribute vec2 uv;
attribute vec4 vertColor;
attribute vec4 joints_0;
attribute vec4 joints_1;
attribute vec4 weights_0;
attribute vec4 weights_1;
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
	if(weights_0.x+weights_0.y+weights_0.z+weights_0.w+weights_1.x+weights_1.y+weights_1.z+weights_1.w > 0){
		mat4 tmp = getJointMatrix();
		gl_Position = projection * (modelview * tmp * vec4(position, 1.0));
	}else{
		gl_Position = projection * (modelview * vec4(position, 1.0));
	}
}