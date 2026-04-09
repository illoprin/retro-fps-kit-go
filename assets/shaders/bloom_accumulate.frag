#version 410 core

in vec2 texcoord;
out vec4 out_fragcolor;

uniform sampler2D u_add;
uniform float u_weight;

void main() {
  vec3 color = texture(u_add, texcoord).rgb * u_weight;
  out_fragcolor = vec4(color, 1.0);
}