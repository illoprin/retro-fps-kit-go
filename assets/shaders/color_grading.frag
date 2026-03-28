#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;
uniform float u_brightness;
uniform float u_saturation;
uniform float u_contrast;
uniform float u_gamma;
uniform float u_exposure;
uniform vec3 u_shadow_color = vec3(0.07, 0.13, 0.25);
uniform vec3 u_mid_color = vec3(0.89, 0.75, 0.38);
uniform vec3 u_highlight_color = vec3(0.9, 0.82, 0.74);
uniform float u_color_strength = 0.6;

// ACES Filmic Tone Mapping Curve
vec3 acesTonemap(vec3 x) {
  const float a = 2.51;
  const float b = 0.03;
  const float c = 2.43;
  const float d = 0.59;
  const float e = 0.14;
	vec3 result = (x * (a * x + b)) / (x * (c * x + d) + e); 
  return clamp(result, 0.0, 1.0);
}

// Uncharted 2 tonemapping curve
vec3 unchartedTonemap(vec3 x) {
  float A = 0.15; // Shoulder Strength
  float B = 0.50; // Linear Strength
  float C = 0.10; // Linear Angle
  float D = 0.20; // Toe Strength
  float E = 0.02; // Toe Numerator
  float F = 0.30; // Toe Denominator

  return ((x * (A * x + C * B) + D * E) / (x * (A * x + B) + D * F)) - E / F;
}

// (Rec. 709)
const vec3 LUMINANCE_WEIGHTS = vec3(0.2126, 0.7152, 0.0722);

vec3 applyColorGrading(vec3 color) {
  // 1. Apply brightness (multiplicatively for linear space)
  color *= u_brightness;

  // 2. Apply contrast
  // Contrast in linear space: offset from the mean (0.5)    
  color = ((color - 0.5) * u_contrast) + 0.5;

  // Safe from negative
  color = max(color, 0.0);

  // color tint

  // 3. Color tinting
  if (u_color_strength > 0.0) {
    vec3 tint;
    if (color.r < 0.333) {
      // Shadows
      float t = color.r / 0.333;
      tint = mix(u_shadow_color, u_mid_color, t);
    } else if (color.r < 0.666) {
      // Midtones
      float t = (color.r - 0.333) / 0.333;
      tint = mix(u_mid_color, u_highlight_color, t);
    } else {
      // Highlights
      tint = u_highlight_color;
    }
    color = mix(color, color * tint, u_color_strength);
  }

  // 4. Apply saturation
  // Calculate the pixel's luminance
  float luminance = dot(color, LUMINANCE_WEIGHTS);

  // Mix the original color with shades of gray
  // luminance = 0 - fully gray, 1 - fully saturated
  color = mix(vec3(luminance), color, u_saturation);

  return color;
}

out vec4 out_frag_color;

void main() {

  // get texture
  vec4 result = texture(u_color, texcoord);
  vec3 color = result.rgb;

  // apply color grading
  color = applyColorGrading(color);

  // apply tonemapping
  color *= u_exposure;
  color = acesTonemap(color);

  // gamma correction
  color = pow(color, vec3(1.0 / u_gamma));

  out_frag_color = vec4(color, result.a);
}