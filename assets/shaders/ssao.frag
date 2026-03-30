#version 410 core

out float out_frag_color;

in vec2 texcoord;

uniform sampler2D u_normal;
uniform sampler2D u_depth;
uniform sampler2D u_noise;

uniform mat4 u_projection;
uniform mat4 u_invprojection;
uniform vec3 u_samples[64];
uniform int u_kernel_size;
uniform vec2 u_noise_scale;

uniform float u_radius;
uniform float u_bias;

// Восстанавливает позицию в view space из depth и UV
vec3 ReconstructPosition(vec2 uv, float depth) {
    // NDC координаты: x, y в [-1,1], z в [-1,1] (OpenGL)
    vec4 ndc = vec4(uv * 2.0 - 1.0, depth * 2.0 - 1.0, 1.0);
    vec4 viewPos = u_invprojection * ndc;
    return viewPos.xyz / viewPos.w;
}

void main() {
    vec2 uv = texcoord;

    // Читаем нормаль (уже в view space, предполагаем [0,1] → [-1,1])
    vec3 normal = normalize(texture(u_normal, uv).xyz * vec3(2.0) - vec3(1.0));
    // vec3 normal = normalize(texture(u_normal, uv).xyz);

    // Читаем depth и восстанавливаем позицию
    float depth = texture(u_depth, uv).r;
    vec3 fragPos = ReconstructPosition(uv, depth);

    // Случайный вектор из noise текстуры (тайлинг по экрану)
    vec3 randomVec = normalize(texture(u_noise, uv * u_noise_scale).xyz * 2.0 - 1.0);

    // TBN матрица для ориентации ядра вдоль нормали (Gram-Schmidt)
    vec3 tangent   = normalize(randomVec - normal * dot(randomVec, normal));
    vec3 bitangent = cross(normal, tangent);
    mat3 TBN = mat3(tangent, bitangent, normal);

    float occlusion = 0.0;

    for (int i = 0; i < u_kernel_size; ++i) {
        // Трансформируем семпл в view space
        vec3 samplePos = TBN * u_samples[i];
        samplePos = fragPos + samplePos * u_radius;

        // Проецируем семпл в NDC → UV
        vec4 offset = vec4(samplePos, 1.0);
        offset = u_projection * offset; // WARN (оптимизация) матричная операция (можно просто спроецировать xy)
        offset.xyz /= offset.w;
        offset.xyz = (offset.xyz + 1) / 2; // [−1,1] → [0,1]

        // Глубина сцены в точке семпла
        float sampleDepth = texture(u_depth, offset.xy).r;
        // WARN оптимизация
        vec3 sampleScenePos = ReconstructPosition(offset.xy, sampleDepth);

        // Range check: исключаем семплы вне радиуса
        float fragSampleDist = distance(fragPos, sampleScenePos);
        float rangeCheck = smoothstep(u_radius, u_radius * 0.5, fragSampleDist);

        // Семпл загораживает фрагмент, если он глубже (с учётом bias)
        float occlusionValue = (sampleScenePos.z >= samplePos.z + u_bias ? 1.0 : 0.0);
        occlusion += occlusionValue * rangeCheck;
    }

    occlusion = 1.0 - (occlusion / float(u_kernel_size));
    out_frag_color = occlusion;
}