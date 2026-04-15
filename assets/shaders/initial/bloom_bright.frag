#version 410 core

in vec2 texcoord;
out vec4 out_fragcolor;

float luminance(vec3 color) {
  return dot(color, vec3(0.2126, 0.7152, 0.0722));
}

uniform sampler2D u_color;
uniform float u_threshold;
uniform float u_knee = 0.5;

void main() {
  vec3 color = texture(u_color, texcoord).rgb;

  float luma = luminance(color);

  float knee = u_threshold * u_knee; // например 0.5

  float soft = luma - u_threshold + knee;
  soft = clamp(soft, 0.0, 2.0 * knee);
  soft = soft * soft / (4.0 * knee + 0.00001);

  float contribution = max(luma - u_threshold, soft);

  vec3 bright = color * contribution / max(luma, 0.00001);
  out_fragcolor = vec4(bright, 1.0);
}