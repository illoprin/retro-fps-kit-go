#version 410 core

out float out_frag_color;
in vec2 texcoord;
  
uniform sampler2D u_raw_ssao;

void main() {
    vec2 texelSize = 1.0 / vec2(textureSize(u_raw_ssao, 0));
    float result = 0.0;
    for (int x = -2; x < 2; ++x) 
    {
        for (int y = -2; y < 2; ++y) 
        {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            result += texture(u_raw_ssao, texcoord + offset).r;
        }
    }
    out_frag_color = result / (4.0 * 4.0);
}  