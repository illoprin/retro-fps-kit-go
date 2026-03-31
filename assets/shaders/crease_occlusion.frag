#version 410 core

// Самописный алгоритм затенения стыков...
// на основе нормалей соседних текселей

#define KERNEL_SIZE_MAX 256

out float out_frag_color;
in vec2 texcoord;

uniform sampler2D u_normal; // Текстура нормалей (view space)
uniform sampler2D u_depth;  // Текстура глубины
uniform sampler2D u_position; // Текстура позиций (view space)

// Настройки эффекта
uniform vec2  u_samples[KERNEL_SIZE_MAX]; // выборка соседних пикселей нормали
uniform float u_radius = 2.0;       // Радиус выборки в пикселях
uniform float u_depthbias = 0.05;   // Порог разницы глубин (чтобы избежать "ореола")
uniform float u_intensity = 1.5;    // Сила затенения
uniform int u_kernel_size;

void main() {
  vec3 normal = texture(u_normal, texcoord).rgb;
  float depth = texture(u_depth, texcoord).r;
  vec3 position = texture(u_position, texcoord).rgb;
  
  // no geometry case
  if (length(normal) < 0.1) {
    out_frag_color = 1.0; 
    return;
  }

  float ao = 0.0;
  float totalWeight = 0.0;
  vec2 texelSize = 1.0 / textureSize(u_normal, 0);

  for(int i = 0; i < min(KERNEL_SIZE_MAX, u_kernel_size); ++i) {
    // get neighbor normal and depth
    vec2 uvWithOffset = clamp(
      texcoord + u_samples[i] * texelSize * u_radius,
      vec2(0.0),
      vec2(1.0)
    );
    vec3 neighborNormal = texture(u_normal, uvWithOffset).rgb;
    if(length(neighborNormal) < 0.001) continue;
    float neighborDepth = texture(u_depth, uvWithOffset).r;
    
    vec3 neighborPos = texture(u_position, uvWithOffset).rgb;

    // Вектор от центрального пикселя к соседу
    vec3 dirToNeighbor = normalize(neighborPos - position);

    // 1. Проверка: "смотрит" ли нормаль соседа на нас?
    // Если dot(dirToNeighbor, n) отрицательный, значит поверхность соседа 
    // наклонена в сторону центрального пикселя (впадина).
    float bend = clamp(dot(normal - neighborNormal, dirToNeighbor), 0.0, 1.0);

    // get dot with our normal
    float dotP = dot(normal, neighborNormal);

    // find crease
    float angleWeight = dotP < 0.999 ? max(0.0, 1.0 - dotP) : 0;
    
    // check depth
    // Если разница глубин слишком большая - это разные объекты, не затеняем.
    float depthDiff = abs(depth - neighborDepth);
    float depthRange = smoothstep(u_depthbias, 0.0, depthDiff);

    // 3. Затухание по расстоянию (Distance Falloff)
    // Сэмплы на краю радиуса влияют меньше.
    float distWeight = 1.0 - length(u_samples[i]);

    totalWeight += distWeight;

    ao += depthRange * angleWeight * distWeight * bend;
  }
  // average weight
  ao /= max(totalWeight, 0.001);

  // invert and contrast
  float shadow = clamp(1.0 - (ao * u_intensity), 0.0, 1.0);
  
  // out
  out_frag_color = shadow;
}