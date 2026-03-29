#version 410 core

out float out_frag_color;
in vec2 texcoord;

uniform sampler2D u_raw_ssao;
uniform int u_blur_size = 2;
uniform float u_blackpoint = 0.0;
uniform float u_whitepoint = 1.0;

void main() {
    vec2 texelSize = 1.0 / vec2(textureSize(u_raw_ssao, 0));
    
    float ssaoSum = 0.0;
    int count = 0;

    // blur
    for (int x = -u_blur_size; x < u_blur_size; ++x) 
    {
        for (int y = -u_blur_size; y < u_blur_size; ++y) 
        {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            ssaoSum += texture(u_raw_ssao, texcoord + offset).r;
            count++;
        }
    }

    // get average (blurred)
    float blurredSSAO = ssaoSum / float(count);

    // levels (use max to avoid zero division)
    float range = max(u_whitepoint - u_blackpoint, 0.0001);
    float finalSSAO = clamp((blurredSSAO - u_blackpoint) / range, 0.0, 1.0);

    out_frag_color = finalSSAO;
}