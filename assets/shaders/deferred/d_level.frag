#version 410 core

#define MAX_SURFACES 1024
#define MAX_LIGHTS 256

layout(location = 0) out vec4 out_frag_color;
layout(location = 1) out vec4 out_normal;
layout(location = 2) out vec4 out_position;

in vec2 texcoord;
in vec3 normal;
in vec3 position;
flat in int surface_id;

// 32 bytes
struct Surface {
  int texIndex;
  int emiIndex;
  float emiStrength;
  float _; // padding
  vec4 color;
};

// 32 bytes
struct PointLight {
  vec4 position;
  vec4 color;
};

// 64 bytes
struct SpotLight {
  vec4 position;
  vec4 color;
  vec4 forward;
  vec4 params;
};

// binding 0
layout(std140) uniform SurfaceBlock {
  Surface surfaces[MAX_SURFACES];
};

// binding 1
layout(std140) uniform PointLightsBlock {
  PointLight pl[MAX_LIGHTS];
};

// binding 2
layout(std140) uniform SpotLightsBlock {
  SpotLight sl[MAX_LIGHTS];
};

uniform sampler2DArray u_diffuse;
uniform sampler2DArray u_emissive;

uniform bool u_wireframe = false;
uniform mat4 u_view;
uniform uint u_pointLightsNum = 0;
uniform uint u_spotLightsNum = 0;

// light falloff
float getLightAttenuation(float d, float r) {
  float constant = 1.0;
  float linear = 4.5 / r;
  float quadratic = 75.0 / pow(r, 2);
  return (1 / (constant + linear * d + quadratic * pow(d, 2)));
}

vec3 getPointLight(PointLight pl) {
  // get light params
  vec3 lPos = pl.position.xyz;
  float lRadius = pl.position.w;
  vec3 lColor = pl.color.rgb;
  float lIntensity = pl.color.w;

	// diffuse
  vec3 lightDirection = lPos - position.xyz;
  vec3 norm = normalize(normal);
  vec3 lightDirectionNorm = normalize(lightDirection);
  float diff = max(dot(lightDirectionNorm, norm), 0.0);

  // get light attenuation (falloff)
  float d = length(lightDirection);
  float attenuation = getLightAttenuation(d, lRadius);

  // return diffuse
  return diff * lColor * attenuation * lIntensity;
}

vec3 getSpotLights(SpotLight sl) {
  // get light params
  vec3 lPos = sl.position.xyz;
  float lRadius = sl.position.w;
  vec3 lColor = sl.color.rgb;
  float lIntensity = sl.color.w;
  vec3 lForward = sl.forward.xyz;
  float lCosInner = sl.forward.w;
  float lCosOuter = sl.params.x;

  // return diffuse
  return vec3(0.0);
}

vec4 getDiffuse() {
  Surface s = surfaces[surface_id];

  vec4 color = vec4(1.0);

  // diffuse
  if(s.texIndex > -1) {
    color = texture(u_diffuse, vec3(texcoord, float(s.texIndex)));
  }

  // emissive
  if(s.emiIndex > -1) {
    color += texture(u_emissive, vec3(texcoord, float(s.emiIndex))) * s.emiStrength;
  }

  // color
  color.rgb *= s.color.rgb;

  if(color.a < 0.1) {
    discard;
  }

  // compute point lights
  // TODO ambient lighting
  vec3 lighting = vec3(0.1);
  for(uint i = 0; i < u_pointLightsNum; i++) {
    lighting += getPointLight(pl[i]);
  }

  // compute spot lights
  for(uint i = 0; i < u_spotLightsNum; i++) {
    lighting += getSpotLights(sl[i]);
  }

  // apply lighting
  color.rgb *= lighting;
  return color;
}

void main() {
  vec4 result;

  if(!u_wireframe) {
    result = getDiffuse();
  } else {
    result = vec4(1.0);
  }

  // color
  out_frag_color = result;

  // normal in view space
  out_normal = vec4(normalize(mat3(u_view) * normal), result.a);

	// position in view space
  out_position = vec4((u_view * vec4(position, 1.0)).xyz, result.a);
}