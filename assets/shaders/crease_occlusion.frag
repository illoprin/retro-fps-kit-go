#version 410 core

// Самописный алгоритм затенения стыков...
// на основе нормалей соседних текселей

out float out_frag_color;
in vec2 texcoord;

uniform sampler2D u_normal; // Текстура нормалей (в world или view пространстве)
uniform sampler2D u_depth;  // Текстура глубины

// Настройки эффекта
uniform float u_radius = 2.0;       // Радиус выборки в пикселях
uniform float u_depthbias = 0.05;   // Порог разницы глубин (чтобы избежать "ореола")
uniform float u_intensity = 1.5;    // Сила затенения

void main() {
    vec3 normal = texture(u_normal, texcoord).rgb;
    float depth = texture(u_depth, texcoord).r;
    
    // Если в этой точке нет геометрии (например, фон)
    if (length(normal) < 0.1) {
        out_frag_color = 1.0; 
        return;
    }

    float ao = 0.0;
    vec2 texelSize = 1.0 / textureSize(u_normal, 0);

    // Выборка соседей (крест или 3x3)
    vec2 offsets[4] = vec2[](
        vec2(u_radius, 0.0), vec2(-u_radius, 0.0),
        vec2(0.0, u_radius), vec2(0.0, -u_radius)
    );

    for(int i = 0; i < 4; i++) {
        vec2 offsetCoords = texcoord + offsets[i] * texelSize;
        vec3 neighborNormal = texture(u_normal, offsetCoords).rgb;
        float neighborDepth = texture(u_depth, offsetCoords).r;

        // 1. Проверка Dot Product
        float dotP = dot(normal, neighborNormal);
        
        // 2. Проверка глубины (чтобы не затенять края на фоне неба/других стен)
        float depthDiff = abs(depth - neighborDepth);

        if (dotP < 1.0 && dotP > 0.0 && depthDiff < u_depthbias) {
            // Чем меньше dotP, тем сильнее изгиб и тем гуще тень
            ao += (1.0 - dotP);
        }
    }

    // Инвертируем и усиливаем контраст
    float shadow = clamp(1.0 - (ao * u_intensity), 0.0, 1.0);
    
    // Вывод (для отладки — просто маска затенения)
    out_frag_color = shadow;
}