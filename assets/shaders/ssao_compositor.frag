#version 410 core

out vec4 out_frag_color;
in vec2 texcoord;

uniform sampler2D u_color;
uniform sampler2D u_ssao;

void main() {
  vec4 color = texture(u_color, texcoord);
  float ambientOcclusion = texture(u_ssao, texcoord).r;
  // result color
  out_frag_color = vec4(ambientOcclusion * color.rgb, color.a);
}