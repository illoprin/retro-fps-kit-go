#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;

out vec4 out_frag_color;

void main() {
  // get texture
  out_frag_color = texture(u_color, texcoord);
}