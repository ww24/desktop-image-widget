#version 410 core

in vec2 TexCoord;

out vec4 color;

uniform sampler2D texture0;

void main()
{
    color = texture(texture0, TexCoord);

    // gamma correction
    color = pow(color, vec4(0.4545));
}
