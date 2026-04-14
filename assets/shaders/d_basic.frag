#version 410 core

layout(location = 0) out vec4 out_fragcolor;
layout(location = 1) out vec3 out_normal;
layout(location = 2) out vec3 out_position;

in vec2 uv;
in vec3 normal;
in vec3 position;

// color
uniform sampler2D u_diffuse;
uniform sampler2D u_emissive;
uniform float u_emissive_strength = 1.0;
uniform bool u_useTexture;
uniform bool u_useEmissive;
uniform vec3 u_color;

// lights
uniform vec3 u_light_pos;
uniform vec3 u_light_color;
uniform float u_light_intensity = 5;
uniform float u_light_radius = 37;

// fog
uniform vec3 u_fogColor = vec3(0.17, 0.23, 0.29);

// misc
uniform mat4 u_view;
uniform float u_time = 0.0;
uniform bool u_useDithering = true;

// Light Attenuation Algo
float getLightAttenuation(float d, float r) {
	float constant = 1.0;
	float linear = 4.5 / r;
	float quadratic = 75.0 / pow(r, 2);
	return (1 / (constant + linear * d + quadratic * pow(d, 2)));
}

vec3 getLights() {
	// ambient
	float ambientStrength = 0.2;
	vec3 ambient = ambientStrength * u_light_color;

	// diffuse
	vec3 lightDirection = u_light_pos - position.xyz;
	vec3 norm = normalize(normal);
	vec3 lightDirectionNorm = normalize(lightDirection);
	float diff = max(dot(lightDirectionNorm, norm), 0.0);

	float d = length(lightDirection);
	float attenuation = getLightAttenuation(d, u_light_radius);

	vec3 diffuse = diff * u_light_color * attenuation * u_light_intensity;
	return (ambient + diffuse);
}

vec4 getColor() {
	if(u_useTexture) {
		vec4 color = texture(u_diffuse, uv);
		if(color.a < 0.1)
			discard;
		return color;
	}
	return vec4(u_color, 1.0);
}

float getFogFactor(float dist, float density) {
	float fog = exp(-density * dist);
	return clamp(fog, 0.0, 1.0);
}

vec3 getFog(vec3 src) {
	float distance = gl_FragCoord.z / gl_FragCoord.w;
	float fogFactor = getFogFactor(distance, 0.05);
	return mix(u_fogColor, src, fogFactor);
}

float getGrayScale(vec3 color) {
	return dot(color, vec3(0.299, 0.587, 0.114));
}

const int bayerMatrix[] = int[](0, 2, 3, 1);

// const int bayerMatrix[] = int[](0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5);

vec3 getDithered(vec3 color) {
	// coords
	float speed = 2.0;
	vec2 coord = gl_FragCoord.xy + vec2(u_time * speed);

	// treshold
	int matrixSize = 2;
	ivec2 matrixCoord = ivec2(coord) % matrixSize;
	int treshold = bayerMatrix[matrixCoord.x + matrixCoord.y * matrixSize];

	// dither
	float grayscale = getGrayScale(color);
	float dithered = grayscale > (float(treshold) / 16.0) ? 1.0 : 0.5;

	// out
	return vec3(dithered) * color;
}

void main() {
	vec4 result;

	// normal in view space
	out_normal = normalize(mat3(u_view) * normal);
	// position in view space
	out_position = (u_view * vec4(position, 1.0)).xyz;

	// -- texture/color
	result = getColor();
	// -- fog
	result.rgb = getFog(result.rgb);

	// -- get emissive
	vec3 emissive = vec3(0.0);
	if(u_useEmissive) {
		emissive = texture(u_emissive, uv).rgb * u_emissive_strength;
		result.rgb += emissive * u_emissive_strength;
	} else {
		// -- lights
		result.rgb *= getLights();
	}
	// -- dither
	if(u_useDithering)
		result.rgb = getDithered(result.rgb);

	// color
	out_fragcolor = result;
}