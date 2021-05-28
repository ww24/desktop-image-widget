#version 410 core

in vec2 TexCoord;

out vec4 color;

uniform sampler2D texture0;

void main()
{
    color = texture(texture0, TexCoord);
    
    // discard alpha
    if (color.a < 1.0) discard;

    // gamma correction
    color = pow(color, vec4(0.4545));
}
