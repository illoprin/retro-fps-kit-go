#version 410 core

in vec2 texcoord;

struct vignette_t {
  bool use;
  float radius;
  float softness;
};

struct color_grading_t {
  float brightness;
  float saturation;
  float contrast;
};

uniform sampler2D u_color;
// uniform sampler2D u_position;
// uniform sampler2D u_noramal;
// uniform sampler2D u_depth;
uniform vignette_t u_vignette;
uniform color_grading_t u_color_grading;
// uniform float u_time;

vec3 applyColorGrading(vec3 color, color_grading_t cg) {
    const vec3 LumCoeff = vec3(0.2125, 0.7154, 0.0721);
    vec3 AvgLumin = vec3(0.5, 0.5, 0.5); // Pivot

    vec3 brtColor = color * cg.brightness;
    vec3 intensity = vec3(dot(brtColor, LumCoeff));
    vec3 satColor = mix(intensity, brtColor, cg.saturation);
    vec3 conColor = mix(AvgLumin, satColor, cg.contrast);
    return conColor;
}

out vec4 fragColor;

void main() {
  // get texture
  vec4 tex = texture(u_color, texcoord);
  vec3 color = tex.rgb;

  // apply color grading
  color = applyColorGrading(color, u_color_grading);

  // vignette
  float vignette_amount = smoothstep(
    u_vignette.radius,
    u_vignette.radius - u_vignette.softness,
    length(texcoord - vec2(.5))
  );
  if (u_vignette.use) {
    color.rgb *= vignette_amount;
  }

  fragColor = vec4(color, tex.a);
}