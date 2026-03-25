#version 410 core

in vec2 uv;
in vec3 normal;
in vec3 position;

uniform sampler2D u_texture;
uniform bool u_useTexture;
uniform vec3 u_lightPos;
uniform vec3 u_lightColor;
uniform vec3 u_color;

out vec4 out_fragColor;

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
	
	out_fragColor = result;
}