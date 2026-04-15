#version 410 core

// blur single color texture

out float out_frag_color;
in vec2 texcoord;

uniform sampler2D u_overlay; // R8
uniform int u_blur_size = 2;

void main() {
    // pixel size in uv
    vec2 texelSize = 1.0 / vec2(textureSize(u_overlay, 0));
    
    // perform blur
    float sum = 0.0;
    int count = 0;
    for (int x = -u_blur_size; x < u_blur_size; ++x) {
        for (int y = -u_blur_size; y < u_blur_size; ++y) {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            sum += texture(u_overlay, texcoord + offset).r;
            count++;
        }
    }
    // get average (blurred)
    float blurred = sum / float(count);

    out_frag_color = blurred;
}