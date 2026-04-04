#version 410 core

// levels and multiply single color texture

out vec4 out_frag_color;
in vec2 texcoord;

uniform sampler2D u_color; // RGBA
uniform sampler2D u_overlay; // R8
uniform float u_whitepoint = 1.0;
uniform float u_blackpoint = 0.0;

void main() {
  float overlay = texture(u_overlay, texcoord).r;

  // levels (use max to avoid zero division)
  float range = max(u_whitepoint - u_blackpoint, 0.0001);
  float overlayFinal = clamp((overlay - u_blackpoint) / range, 0.0, 1.0);

  // result color
  vec4 color = texture(u_color, texcoord);
  out_frag_color = vec4(overlayFinal * color.rgb, color.a);
}