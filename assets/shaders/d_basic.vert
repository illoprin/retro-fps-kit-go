#version 410 core

layout(location = 0) in vec3 in_position;
layout(location = 1) in vec2 in_texcoords;
layout(location = 2) in vec3 in_normal;

uniform mat4 u_projection;
uniform mat4 u_view;
uniform mat4 u_model;

out vec2 uv;
out vec3 normal;
out vec3 position;

void main() {
	position = vec3(u_model * vec4(in_position, 1.0));
	normal = mat3(transpose(inverse(u_model))) * in_normal;
	uv = in_texcoords;
	gl_Position = u_projection * u_view * vec4(position, 1.0);
}