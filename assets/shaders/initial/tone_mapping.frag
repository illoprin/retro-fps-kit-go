#version 410 core

// converting to linear space

in vec2 texcoord;

uniform sampler2D u_color;
uniform float u_gamma;
uniform int u_tonemap_type = 1;

// ACES Filmic Tone Mapping Curve
vec3 ACESTonemap(vec3 x) {
  const float a = 2.51;
  const float b = 0.03;
  const float c = 2.43;
  const float d = 0.59;
  const float e = 0.14;
  vec3 result = (x * (a * x + b)) / (x * (c * x + d) + e);
  return clamp(result, 0.0, 1.0);
}

// Uncharted 2 tonemapping curve
vec3 UnchartedTonemap(vec3 x) {
  float A = 0.15; // Shoulder Strength
  float B = 0.50; // Linear Strength
  float C = 0.10; // Linear Angle
  float D = 0.20; // Toe Strength
  float E = 0.02; // Toe Numerator
  float F = 0.30; // Toe Denominator

  return ((x * (A * x + C * B) + D * E) / (x * (A * x + B) + D * F)) - E / F;
}

// Simple Reinchard Tonemap
vec3 ReinchardTonemap(vec3 hdr) {
  return hdr / (hdr + vec3(1.0));
}

out vec4 out_frag_color;

void main() {
  // get texture
  vec4 result = texture(u_color, texcoord);
  vec3 color = result.rgb;

  // apply tonemapping
  switch(u_tonemap_type) {
    case 1:
      color = ACESTonemap(color);
    case 2:
      color = UnchartedTonemap(color);
    case 3:
      color = ReinchardTonemap(color);
  }

  // gamma correction
  color = pow(color, vec3(1.0 / u_gamma));

  out_frag_color = vec4(color, result.a);
}