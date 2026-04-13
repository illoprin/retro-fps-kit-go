#version 410 core

// #extension GL_ARB_shading_language_include : enable

float luminance(vec3 color) {
  return dot(color, vec3(0.2126, 0.7152, 0.0722));
}

#define KERNEL_SIZE 16

uniform sampler2D u_color;
uniform vec2 u_samples[KERNEL_SIZE];
uniform float u_radius = 20.0;

out float out_luminance;

void main() {
  vec2 texelSize = vec2(1.0) / textureSize(u_color, 0);
  vec2 texcoord = vec2(0.5);

  // calculate luminance for samples near center
  float avg = 0.0;
  for(int i = 0; i < KERNEL_SIZE; ++i) {
    vec2 offset = texcoord + u_samples[i] * u_radius * texelSize;
    vec3 color = texture(u_color, offset).rgb;
    avg += luminance(color);
  }

  out_luminance = avg / (KERNEL_SIZE + 0.0);
}