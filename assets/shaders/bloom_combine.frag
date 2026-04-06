#version 330 core

in vec2 texcoord;
out vec4 out_fragcolor;

uniform sampler2D u_hdr;
uniform sampler2D u_bloom;
uniform float u_intensity;
uniform vec3 u_tint;

void main() {
  vec3 hdr = texture(u_hdr, texcoord).rgb;
  vec3 bloom = texture(u_bloom, texcoord).rgb;

  vec3 result = hdr + bloom * u_intensity * u_tint;

  out_fragcolor = vec4(result, 1.0);
}