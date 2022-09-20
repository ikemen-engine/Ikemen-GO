#version 400
precision highp float;

uniform vec2 TextureSize;

in vec2 VertCoord;
out vec2 texcoord;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	texcoord = (VertCoord + 1.0) / 2.0;
}
