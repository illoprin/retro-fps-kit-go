#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;

uniform float u_exposure;

out vec4 out_frag_color;

void main() {
  vec4 color = texture(u_color, texcoord);
  out_frag_color = vec4(color.rgb * u_exposure, color.a);
}