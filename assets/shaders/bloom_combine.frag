#version 330 core

in vec2 texcoord;
out vec4 out_fragcolor;

uniform sampler2D u_hdr;
uniform sampler2D u_bloom;
uniform sampler2D u_lens;
uniform float u_intensity;
uniform float u_lens_intensity = 1.0;
uniform vec3 u_tint;
uniform bool u_use_lens = false;

float lerp(float a, float b, float r) {
  return min(a, b) + abs(a - b) * r;
}

void main() {
  vec3 hdr = texture(u_hdr, texcoord).rgb;
  float lens = texture(u_lens, texcoord).r;
  vec3 bloom = texture(u_bloom, texcoord).rgb;

  vec3 resultBloom = bloom * u_intensity * u_tint;
  if(u_use_lens)
    resultBloom += resultBloom * lens * u_lens_intensity;

  vec3 result = hdr + resultBloom;

  out_fragcolor = vec4(result, 1.0);
}