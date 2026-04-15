#version 410 core

layout(location = 0) out vec4 out_fragcolor;

uniform float u_bigcell = 0.1;
uniform float u_smallcell = 0.01;
uniform float u_linewidth = 0.01;

float gridLine(float coord, float size, float width) {
  float d = mod(coord, size);
  return step(d, width) + step(size - d, width);
}

in vec3 position;

const float fogEnd = 30.0; // far clip for grid fade
const float fogStart = 20.0;

void main() {

  float big = gridLine(position.x, u_bigcell, u_linewidth * 2.0) +
    gridLine(position.z, u_bigcell, u_linewidth * 2.0);

  float small = gridLine(position.x, u_smallcell, u_linewidth) +
    gridLine(position.z, u_smallcell, u_linewidth);

  float grid = small * 0.3 + big * 1.0;

  if(grid < 0.1)
    discard;

  // calc fog
  float distance = gl_FragCoord.z / gl_FragCoord.w;
  float fog = (distance - fogEnd) / (fogEnd - fogStart);
  float d = clamp(1.0 - fog, 0.0, 1.0);

  grid *= d;

  // output
  out_fragcolor = vec4(vec3(grid), 1.0);
}