#version 410 core

layout(location = 0) in vec2 in_position;

out vec2 texcoord;

void main() {
  texcoord = clamp(in_position + vec2(1, 1), vec2(0, 0), vec2(1, 1));

  gl_Position = vec4(in_position, 0, 1);
}


