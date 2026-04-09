#version 410 core

in vec2 texcoord;

uniform sampler2D u_color;
uniform float u_brightness;
uniform float u_saturation;
uniform float u_contrast;
uniform vec3 u_shadow_color = vec3(0.07, 0.13, 0.25);
uniform vec3 u_mid_color = vec3(0.89, 0.75, 0.38);
uniform vec3 u_highlight_color = vec3(0.9, 0.82, 0.74);
uniform float u_color_strength = 0.6;

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
  if(u_color_strength > 0.0) {
    vec3 tint;
    if(color.r < 0.333) {
      // Shadows
      float t = color.r / 0.333;
      tint = mix(u_shadow_color, u_mid_color, t);
    } else if(color.r < 0.666) {
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

  out_frag_color = vec4(color, result.a);
}