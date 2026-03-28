#version 410 core

layout(location = 0) out vec4 out_fragcolor;
layout(location = 1) out vec3 out_normal;
layout(location = 2) out vec3 out_position;

in vec2 uv;
in vec3 normal;
in vec3 position;

uniform sampler2D u_texture;
uniform bool u_useTexture;
uniform vec3 u_light_pos;
uniform vec3 u_light_color;
uniform float u_light_intensity = 1.2;
uniform float u_light_radius = 50;
uniform vec3 u_color;
uniform float u_gamma = 2.2;
uniform float u_exposure = 1.05;

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

vec3 unchartedTonemap(vec3 x) {
    float A = 0.15; // Shoulder Strength
    float B = 0.50; // Linear Strength
    float C = 0.10; // Linear Angle
    float D = 0.20; // Toe Strength
    float E = 0.02; // Toe Numerator
    float F = 0.30; // Toe Denominator

    return ((x * (A * x + C * B) + D * E) / (x * (A * x + B) + D * F)) - E / F;
}

// Light Attenuation Algo
float getLightAttenuation(float d, float r) {
	float constant = 1.0;
	float linear = 4.5 / r;
	float quadratic = 75.0 / pow(r, 2);
	return (1 / (constant + linear * d + quadratic * pow(d, 2)));
}

void main() {
	// ambient
	float ambientStrength = 0.4;
	vec3 ambient = ambientStrength * u_light_color;
	
	// diffuse
	vec3 lightDirection = u_light_pos - position;
	vec3 norm = normalize(normal);
	vec3 lightDirectionNorm = normalize(lightDirection);
	float diff = max(dot(lightDirectionNorm, norm), 0.0);

	float d = length(lightDirection);
	float attenuation = getLightAttenuation(d, u_light_radius);

	vec3 diffuse = diff * u_light_color * attenuation * u_light_intensity;

	vec4 result = vec4((ambient + diffuse) * u_color, 1.0);

	result.rgb = pow(result.rgb, vec3(u_gamma));

	// apply texture if needed
	if (u_useTexture) {
		vec4 texColor = texture(u_texture, uv);
		result *= texColor;
	}

	// apply exposure
	result.rgb *= u_exposure;

	// apply tonemapping
	result.rgb = acesTonemap(result.rgb);

	// normalize white point
	float whiteScale = 1.0 / acesTonemap(vec3(11.2)).r;
  result.rgb *= whiteScale;

	// gamma correction
	result.rgb = pow(result.rgb, vec3(1.0 / u_gamma));

	// setup outs
	out_normal = normalize(normal);
	out_position = position;
	out_fragcolor = result;
}