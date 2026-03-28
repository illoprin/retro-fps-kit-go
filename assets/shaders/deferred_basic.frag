#version 410 core

layout(location = 0) out vec4 out_fragcolor;
layout(location = 1) out vec4 out_normal;
layout(location = 2) out vec4 out_position;

in vec2 uv;
in vec3 normal;
in vec3 position;

uniform sampler2D u_texture;
uniform bool u_useTexture;
uniform vec3 u_light_pos;
uniform vec3 u_light_color;
uniform float u_light_intensity = 3;
uniform float u_light_radius = 37;
uniform vec3 u_color;

// Light Attenuation Algo
float getLightAttenuation(float d, float r) {
	float constant = 1.0;
	float linear = 4.5 / r;
	float quadratic = 75.0 / pow(r, 2);
	return (1 / (constant + linear * d + quadratic * pow(d, 2)));
}

void main() {
	// ambient
	float ambientStrength = 0.2;
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

	// apply texture if needed
	if (u_useTexture) {
		vec4 texColor = texture(u_texture, uv);
		result *= texColor;
	}

	// setup outs
	out_normal = vec4(normalize(normal), result.a);
	out_position = vec4(position, result.a);
	out_fragcolor = result;
}