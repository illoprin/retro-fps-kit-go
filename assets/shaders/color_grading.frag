#version 410 core

in vec2 texcoord;

uniform float u_brightness;
uniform float u_saturation;
uniform float u_contrast;
uniform float u_gamma;
uniform float u_exposure;

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

vec3 doColorGrading(vec3 color) {
    const vec3 lum_factor = vec3(0.2125, 0.7154, 0.0721);
    vec3 avg_lumin = vec3(0.5, 0.5, 0.5); // Pivot

    vec3 brtColor = color * u_brightness;
    vec3 intensity = vec3(dot(brtColor, lum_factor));
    vec3 satColor = mix(intensity, brtColor, u_saturation);
    vec3 conColor = mix(avg_lumin, satColor, u_contrast);
    return conColor;
}

uniform sampler2D u_color;

out vec4 out_frag_color;

void main() {

  // get texture
  vec4 result = texture(u_color, texcoord);
  vec3 color = result.rgb;

  // apply gamma
  // goto nonlinear space
  color = pow(color, vec3(u_gamma));

  // apply tonemapping
  color *= u_exposure;
  color = acesTonemap(color);
  float whiteScale = 1.0 / acesTonemap(vec3(11.2)).r;
  result.rgb *= whiteScale;

  // apply color grading
  color = doColorGrading(color);

  // back to linear space
  color = pow(color, vec3(1.0 / u_gamma));

  out_frag_color = vec4(color, result.a);

}