#version 420 core

#define MAX_SURFACES 1024

in vec2 texcoord;
in vec3 normal;
flat in int surface_id;

// 32 bytes
struct Surface {
  int texIndex;
  int emiIndex;
  float emiStrength;
  vec4 color;
};

layout(std140, binding = 0) uniform SurfaceBlock {
  Surface surfaces[MAX_SURFACES];
};

uniform sampler2DArray u_diffuse;
uniform sampler2DArray u_emissive;

uniform bool u_wireframe = false;

out vec4 out_frag_color;

void main() {
  vec4 result;
  Surface s = surfaces[surface_id];
  if(!u_wireframe) {
    // diffuse
    if(s.texIndex > -1)
      result = texture(u_diffuse, vec3(texcoord, float(s.texIndex)));
    // emissive
    result += texture(u_emissive, vec3(texcoord, float(s.emiIndex))) * s.emiStrength;
    // color
    result.rgb *= s.color.rgb;
    // apply lights ...
  } else {
    result = vec4(1.0);
  }

  out_frag_color = result;
}