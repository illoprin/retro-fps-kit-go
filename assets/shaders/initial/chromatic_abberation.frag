#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;

uniform float u_strength;
uniform float u_radius;
uniform float u_power;

out vec4 u_fragcolor;

void main() {
  vec2 center = vec2(0.5);
  vec2 dir = texcoord - center;

  float dist = length(dir);

  float edgeFactor = smoothstep(u_radius, 1.0, dist);
  edgeFactor = pow(edgeFactor, u_power);

  float aberration = edgeFactor * u_strength;

  vec2 offset = normalize(dir + 1e-6) * aberration;

  vec2 uvR = texcoord + offset * .5;
  vec2 uvB = texcoord + offset * (-.5);

  float r = texture(u_color, uvR).r;
  float g = texture(u_color, texcoord).g;
  float b = texture(u_color, uvB).b;

  u_fragcolor = vec4(r, g, b, 1.0);
}