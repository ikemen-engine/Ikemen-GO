attribute vec2 VertCoord;

uniform vec2 TextureSize;

varying vec2 texcoord;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	texcoord = (VertCoord + 1.0) / 2.0;
}
