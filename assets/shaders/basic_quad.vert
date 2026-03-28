#version 410 core

layout(location = 0) in vec2 in_position;

out vec2 texcoord;

void main() {
  // get texcoord
  texcoord = clamp(in_position + vec2(1, 1), vec2(0, 0), vec2(1, 1));

  // out position
  gl_Position = vec4(in_position, 0, 1);
}