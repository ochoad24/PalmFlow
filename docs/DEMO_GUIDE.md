# Guion de demostración técnica y funcional

## Historia que cuenta la demo

Una empresa mueve crudo de palma desde El Triunfo, Santa Rosa, Repsa I o Repsa II hacia Zolic. Cada viaje tiene dos pesajes expresados en toneladas métricas:

1. Báscula de salida: peso bruto, tara y peso neto.
2. Báscula de entrada: peso bruto, tara y peso neto.

La diferencia puede ser una merma normal, un error operativo, un problema de calidad del dato o una señal que requiere investigación.

## Flujo de 8 minutos

### 1. Contexto — 1 minuto

- Abrir `http://localhost:3080`.
- Mostrar que PostgreSQL y KEV están en línea.
- Explicar que ningún dato se envía a TypeSafe ni a OpenAI.
- KEV expone localmente el contrato compatible `POST /v1/systemone`.

### 2. Datos — 1 minuto

- Pulsar **Cargar demo (1,000)**.
- Mostrar las columnas de salida y entrada.
- Descargar la plantilla CSV.
- Explicar que el conjunto de 1,000 registros ficticios es reproducible y contiene casos normales y anomalías deliberadas.

### 3. Política — 1 minuto

- Abrir **Configuración**.
- Mostrar la tolerancia inicial de merma de 0.5%.
- Mostrar demora máxima, ganancia de peso permitida y precio USD/t.
- Aclarar que son parámetros de demostración, no una política recomendada para producción.

### 4. Análisis — 2 minutos

- Pulsar **Analizar pendientes**.
- Observar el progreso por registro.
- Explicar la separación de responsabilidades:
  - Go calcula diferencias, porcentajes, consistencia e impacto económico.
  - KEV clasifica anomalía, calidad, fraude y prioridad.
  - El resultado combina reglas y probabilidades para asignar severidad.

### 5. Hallazgos — 2 minutos

- Filtrar por **Crítico**, **Alto** y **Alerta**.
- Revisar un caso de merma alta.
- Revisar un caso de entrada mayor que salida.
- Revisar un cambio de tara y una demora.
- Mostrar riesgo, probabilidad de fraude, reglas activadas y costo comparado por fila.

### 6. Costos — 1 minuto

- KEV: costo API incremental USD 0; la latencia mostrada es real.
- Jev: costo calculado con los tokens reales de entrada reportados por KEV.
- GPT-6.1 Sol: costo calculado con la misma entrada y una salida configurable.
- Recalcar que energía, hardware, operación y mantenimiento no son cero y no están incluidos.

## Arquitectura

```text
Navegador
   │
   ▼
Nginx + React :3080
   │ REST + SSE
   ▼
API Go :8080 ───────────────► KEV local :8009
   │                           POST /v1/systemone
   ▼
PostgreSQL :5432
```

KEV corre fuera de Docker para usar CUDA directamente. La API accede al host mediante `host.docker.internal`.

## Forma de cada decisión enviada a KEV

El `state` contiene:

- Política vigente de la demo.
- Identificación del viaje.
- Pesos de salida y entrada.
- Diferencia de tara.
- Merma absoluta y porcentual ya calculadas.
- Distancia y tiempos.

Las preguntas se envían juntas para que KEV lea el estado una vez:

```json
{
  "weight_anomaly": {"type": "noul"},
  "data_quality": {"type": "choice"},
  "transit_anomaly": {"type": "noul"},
  "fraud_risk": {"type": "score"},
  "review_priority": {"type": "score"}
}
```

## Fórmulas de costo

### Jev 1.14

```text
costo = tokens_entrada / 1,000,000 × USD 0.042
```

La salida es gratuita según la tarifa publicada consultada el 16 de noviembre de 2026.

### GPT-6.1 Sol

```text
costo =
  tokens_entrada / 1,000,000 × USD 2
  + tokens_salida_simulados / 1,000,000 × USD 10
```

El escenario inicial usa 220 tokens de salida por registro. Puede modificarse en la UI.

### KEV local

```text
costo API incremental = USD 0
```

Para un TCO real deben agregarse:

- Consumo eléctrico de GPU y CPU.
- Amortización del equipo.
- Operación, monitoreo y respaldo.
- Evaluación, etiquetado y eventual ajuste fino.

## Qué no debe afirmarse en la presentación

- Una probabilidad alta no prueba fraude.
- KEV-0.8B no tiene la misma precisión que Jev ni que KEV-4B/9B.
- La calibración publicada no sustituye calibración con datos propios.
- La tolerancia de 0.5% es un valor configurable de demostración.
- Las latencias simuladas de Jev y GPT no son SLA ni garantías.

## Siguiente fase recomendada

1. Conseguir registros anonimizados y etiquetas de revisión humana.
2. Definir reglas oficiales por ruta, producto, clima y tipo de vehículo.
3. Medir precisión, Brier score y calibración por pregunta.
4. Comparar KEV-0.8B con KEV-4B en hardware adecuado.
5. Fijar umbrales de automatización y revisión humana.
6. Añadir autenticación, auditoría de cambios y exportación de casos.
