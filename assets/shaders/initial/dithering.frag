#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;
uniform sampler2D u_matrix; // val / 16.0
uniform int u_matrix_size;
uniform float u_speed = 2.0;
uniform float u_time = 0.0;
uniform float u_min = 0.5;
uniform float u_max = 0.5;

float getGrayScale(vec3 color) {
  return dot(color, vec3(0.299, 0.587, 0.114));
}

out vec3 out_fragcolor;

void main() {

  // color texture
  vec3 color = texture(u_color, texcoord).rgb;

  // matrix treshold value (R8 texture)
  vec2 coord = gl_FragCoord.xy + vec2(u_time * u_speed);
  ivec2 matrixCoord = ivec2(coord) % u_matrix_size;
  float treshold = texelFetch(u_matrix, matrixCoord, 0).r;

	// dither
  float grayscale = getGrayScale(color);
  float dithered = grayscale > treshold ? u_min : u_max;

	// out
  out_fragcolor = vec3(dithered) * color;
}