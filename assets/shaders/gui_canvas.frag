#version 410 core

in vec2 texcoord;
in vec4 color;

out vec4 out_fragcolor;

uniform sampler2D u_texture;
uniform bool u_useTexture = false;

void main() {
  // mix color and texture
  if(u_useTexture)
    out_fragcolor = texture(u_texture, texcoord) * color;
  else
    out_fragcolor = color;
}