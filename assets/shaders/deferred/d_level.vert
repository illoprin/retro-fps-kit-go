#version 410 core

layout(location = 0) in vec3 in_position;
layout(location = 1) in vec2 in_texcoord;
layout(location = 2) in vec3 in_normal;
layout(location = 3) in int in_surface_id;

uniform mat4 u_projection;
uniform mat4 u_view;

out vec2 texcoord;
out vec3 normal;
out vec3 position;
flat out int surface_id;

void main() {
  surface_id = in_surface_id;
  normal = in_normal;
  texcoord = in_texcoord;
  position = in_position;
  gl_Position = u_projection * u_view * vec4(in_position, 1.0);
}