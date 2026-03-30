#version 410 core

in vec2 texcoord;

out vec4 out_frag_color;

uniform sampler2D u_color;
uniform sampler2D u_occlusion;

void main() {
  float occlusion = texture(u_occlusion, texcoord).r;
  vec4 color = texture(u_color, texcoord);
  color.rgb *= occlusion;

  out_frag_color = color;
}