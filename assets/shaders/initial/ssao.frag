#version 410 core

out float out_frag_color;

in vec2 texcoord;

uniform sampler2D u_normal; // view space
uniform sampler2D u_position; // view space
uniform sampler2D u_noise;

uniform vec2 u_proj_info; // cam_projection[0][0]; cam_projection[1][1]
uniform vec3 u_samples[64];
uniform int u_kernel_size;
uniform vec2 u_noise_scale;

uniform float u_radius;
uniform float u_bias;

void main() {

  // Читаем нормаль (уже в view space)
  vec3 normal = texture(u_normal, texcoord).xyz;

  if (length(normal) < 0.1) {out_frag_color = 1.0; return; };

  // Читаем позицию (можно заменить на восстановление ReconstructPos)
  vec3 position = texture(u_position, texcoord).rgb;

  // Случайный вектор из noise текстуры (тайлинг по экрану)
  vec3 randomVec = normalize(texture(u_noise, texcoord * u_noise_scale).xyz * 2.0 - 1.0);

  // TBN матрица для ориентации ядра вдоль нормали (Gram-Schmidt)
  vec3 tangent   = normalize(randomVec - normal * dot(randomVec, normal));
  vec3 bitangent = cross(normal, tangent);
  mat3 TBN = mat3(tangent, bitangent, normal);

  float occlusion = 0.0;

  for (int i = 0; i < u_kernel_size; ++i) {
    // Трансформируем семпл в view space
    vec3 samplePos = position + TBN * u_samples[i] * u_radius;

    // 1. Проецируем сэмпл вручную (View Space -> Clip Space)
    // В View Space камера смотрит в -Z, поэтому делим на -samplePos.z
    float invZ = 1.0 / -samplePos.z;
    vec2 ndc;
    ndc.x = (samplePos.x * u_proj_info.x) * invZ;
    ndc.y = (samplePos.y * u_proj_info.y) * invZ;
    // 2. Переходим из NDC [-1, 1] в UV [0, 1]
    vec2 offsetUV = ndc * 0.5 + 0.5;

    if(offsetUV.x < 0.0 || offsetUV.y < 0.0 || offsetUV.x > 1.0 || offsetUV.y > 1.0) continue;

    // Глубина сцены в точке семпла
    vec3 sampleScenePos = texture(u_position, offsetUV.xy).rgb;

    // Range check: исключаем семплы вне радиуса
    float fragSampleDist = distance(position, sampleScenePos);
    float rangeCheck = smoothstep(u_radius, u_radius * 0.5, fragSampleDist);

    // Семпл загораживает фрагмент, если он глубже (с учётом bias)
    float occlusionValue = (sampleScenePos.z >= samplePos.z + u_bias ? 1.0 : 0.0);
    occlusion += occlusionValue * rangeCheck;
  }

  occlusion = 1.0 - (occlusion / float(u_kernel_size));
  out_frag_color = occlusion;
}