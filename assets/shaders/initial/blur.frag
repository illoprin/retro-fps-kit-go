#version 410 core

in vec2 texcoord;
out vec4 out_fragcolor;

uniform sampler2D u_color;
uniform float u_radius;      // blur radius in texels
uniform bool u_horizontal;

const float weight[5] = float[](0.227027, 0.1945946, 0.1216216, 0.054054, 0.016216);

void main() {
  ivec2 texSize = textureSize(u_color, 0);
  vec2 texel = u_radius / vec2(texSize);  // convert radius to UV space

  vec3 result = texture(u_color, texcoord).rgb * weight[0]; // central pixel

  for(int i = 1; i < 5; i++) {
    vec2 offset = u_horizontal ? vec2(texel.x * i, 0.0) : vec2(0.0, texel.y * i);
    // clamp coordinates manually
    vec2 uv_plus = clamp(texcoord + offset, vec2(0.0), vec2(1.0));
    vec2 uv_minus = clamp(texcoord - offset, vec2(0.0), vec2(1.0));

    result += texture(u_color, uv_plus).rgb * weight[i];
    result += texture(u_color, uv_minus).rgb * weight[i];
  }

  out_fragcolor = vec4(result, 1.0);
}