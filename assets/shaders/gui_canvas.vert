#version 410 core

layout(location = 0) in vec2 in_position;
layout(location = 1) in vec2 in_texcoord;
layout(location = 2) in vec4 in_color;

uniform mat4 u_projection;

out vec2 texcoord;
out vec4 color;

void main() {
  // project into clip space
  gl_Position = u_projection * vec4(in_position, 0.0, 1.0);
  texcoord = in_texcoord;
  color = in_color;
}