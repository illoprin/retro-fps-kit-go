#version 410 core

// Самописный алгоритм затенения стыков...
// на основе нормалей соседних текселей

#define KERNEL_SIZE_MAX 256

out float out_frag_color;
in vec2 texcoord;

uniform sampler2D u_normal; // Текстура нормалей (view space)
uniform sampler2D u_depth;  // Текстура глубины
uniform sampler2D u_noise; // шум для поворота сэмплов

// Настройки эффекта
uniform vec2  u_samples[KERNEL_SIZE_MAX]; // выборка соседних пикселей нормали
uniform vec2 u_noise_scale;
uniform float u_radius = 2.0;       // Радиус выборки в пикселях
uniform float u_depthbias = 0.05;   // Порог разницы глубин (чтобы избежать "ореола")
uniform float u_intensity = 1.5;    // Сила затенения
uniform int u_kernel_size;
uniform mat4 u_invprojection;

// ReconstructPosition transform uv and depth...
// to coords in view space
vec3 ReconstructPosition(vec2 uv, float depth) {
  // clip space
  vec4 ndc = vec4(uv * 2.0 - 1.0, depth * 2.0 - 1.0, 1.0);
  // view space
  vec4 viewPos = u_invprojection * ndc;
  return viewPos.xyz / viewPos.w;
}

void main() {
  vec3 normal = texture(u_normal, texcoord).rgb;
  float depth = texture(u_depth, texcoord).r;
  vec3 position = ReconstructPosition(texcoord, depth);
  vec3 noiseRot = texture(u_noise, texcoord * u_noise_scale).rgb * 2.0 - 1.0;

  if (length(normal) < 0.1) {out_frag_color = 1.0; return; };
  
  // no geometry case
  if (length(normal) < 0.1) {
    out_frag_color = 1.0; 
    return;
  }

  // total occlusion value
  float ao = 0.0;
  float totalWeight = 0.0;
  vec2 texelSize = 1.0 / textureSize(u_normal, 0);

  for(int i = 0; i < min(KERNEL_SIZE_MAX, u_kernel_size); ++i) {

    // Получаем позицию соседа с учётом шума поворотов
    // проще отразить, чем делать поворот :)
    vec2 sampleOffset = reflect(u_samples[i], noiseRot.xy);
    vec2 uvWithOffset = texcoord + sampleOffset * texelSize * u_radius;
    uvWithOffset = clamp(uvWithOffset, 0.0, 1.0);

    // get neighbor depth
    float nDepth = texture(u_depth, uvWithOffset).r;
    // Если разница глубин слишком большая - это разные объекты, не затеняем.
    float depthDiff = abs(depth - nDepth);

    // Ранний вылет: если разница глубин слишком велика, сэмпл бесполезен
    if (depthDiff > u_depthbias) continue;

    // get neighbor normal
    vec3 nNormal = normalize(texture(u_normal, uvWithOffset).rgb);
    if(length(nNormal) < 0.001) continue;
    
    // Реконструируем позицию соседа из UV и глубины
    vec3 nPosition = ReconstructPosition(uvWithOffset, nDepth);

    // Вектор от центрального пикселя к соседу
    vec3 dirToNeighbor = normalize(nPosition - position);

    // 1. Проверка: "смотрит" ли нормаль соседа на нас?
    // Если dot(dirToNeighbor, n) отрицательный, значит поверхность соседа 
    // наклонена в сторону центрального пикселя (впадина).
    float bend = clamp(dot(normal - nNormal, dirToNeighbor), 0.0, 1.0);

    // get dot with our normal
    float dotP = dot(normal, nNormal);

    // find crease
    float angleWeight = dotP < 0.999 ? max(0.0, 1.0 - dotP) : 0;
    
    // check depth
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