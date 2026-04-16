#version 410 core

#define MAX_SURFACES 1024

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
  vec4 color;
};

layout(std140) uniform SurfaceBlock {
  Surface surfaces[MAX_SURFACES];
};

uniform sampler2DArray u_diffuse;
uniform sampler2DArray u_emissive;

uniform bool u_wireframe = false;
uniform mat4 u_view;

void main() {
  vec4 result = vec4(1.0);

  Surface s = surfaces[surface_id];
  if(!u_wireframe) {
    // diffuse
    if(s.texIndex > -1)
      result = texture(u_diffuse, vec3(texcoord, float(s.texIndex)));
    // emissive
    if(s.emiIndex > -1)
      result += texture(u_emissive, vec3(texcoord, float(s.emiIndex))) * s.emiStrength;
    // color
    result.rgb *= s.color.rgb;

    if(result.a < 0.1)
      discard;

    // apply lights ...
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