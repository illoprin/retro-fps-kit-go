#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;

out vec4 fragColor;

void main() {
  vec4 tex = texture(u_color, texcoord);
  vec3 color = tex.rgb;
  fragColor = vec4(color, tex.a);
}