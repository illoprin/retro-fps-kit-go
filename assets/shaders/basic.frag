#version 410 core

in vec2 uv;
in vec3 normal;
in vec3 position;


uniform sampler2D u_texture;
uniform bool u_useTexture;
uniform vec3 u_lightPos;
uniform vec3 u_lightColor;
uniform vec3 u_color;
uniform float u_gamma = 2.2;
uniform float u_exposure = 1.05;

out vec4 out_fragColor;

// ACES Filmic Tone Mapping Curve by Krzysztof Narkowicz
vec3 aces(vec3 x) {
    const float a = 2.51;
    const float b = 0.03;
    const float c = 2.43;
    const float d = 0.59;
    const float e = 0.14;
		vec3 result = (x * (a * x + b)) / (x * (c * x + d) + e); 
    return clamp(result, 0.0, 1.0);
}

void main() {
	// ambient
	float ambientStrength = 0.2;
	vec3 ambient = ambientStrength * u_lightColor;
	
	// diffuse
	vec3 norm = normalize(normal);
	vec3 lightDir = normalize(u_lightPos - position);
	float diff = max(dot(norm, lightDir), 0.0);
	vec3 diffuse = diff * u_lightColor;

	vec4 result = vec4((ambient + diffuse) * u_color, 1.0);

	// apply texture if needed
	if (u_useTexture) {
		vec4 texColor = texture(u_texture, uv);
		result *= texColor;
	}

	// apply exposure
	result.rgb *= u_exposure;

	// apply aces
	result.rgb = aces(result.rgb);

	// gamma correction
	result.rgb = pow(result.rgb, vec3(1.0 / u_gamma));
	
	out_fragColor = result;
}