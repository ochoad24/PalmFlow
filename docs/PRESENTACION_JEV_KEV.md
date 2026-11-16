# Jev, System One y KEV
## Decisiones probabilísticas para el transporte de crudo de palma

> Presentación breve — PalmFlow  
> Duración sugerida: **12 a 15 minutos**, incluida la demo.  
> Guía ampliada y notas: [`GUIA_PRESENTACION_JEV_KEV.md`](GUIA_PRESENTACION_JEV_KEV.md).

---

# 1. Del dato a la decisión

En una operación logística recibimos pesos, tiempos, rutas, vehículos, conductores y políticas. El reto no es solo guardar esa información, sino decidir:

- ¿La diferencia de peso es normal?
- ¿El registro es consistente?
- ¿El tiempo de tránsito es anómalo?
- ¿Hay señales que ameritan investigación?
- ¿Qué caso debe revisar primero una persona?

**Objetivo:** producir decisiones estructuradas que el software pueda usar directamente.

---

# 2. ¿Qué es Jev?

**Jev** es el modelo principal de TypeSafe y el primer modelo de la categoría **System One**.

Un modelo System One:

1. Recibe un **estado** con datos, contexto, políticas y evidencia.
2. Evalúa preguntas concretas.
3. Devuelve respuestas tipadas, probabilidades y, según el tipo, confianza.

En vez de solicitar:

> “Analiza este viaje y escribe un reporte.”

se pregunta:

> “¿La diferencia de peso es sospechosa bajo esta política?”

```json
{
  "type": "noul",
  "noul": 0.91
}
```

El resultado es una probabilidad utilizable por el sistema, no una explicación narrativa.

---

# 3. Jev frente a ChatGPT

| Aspecto | Jev / System One | LLM conversacional |
|---|---|---|
| Objetivo | Decidir dentro de un espacio controlado | Generar, explicar o transformar contenido |
| Salida | Tipada y restringida | Texto libre o JSON solicitado |
| Ejemplos | Sí/no, categoría, puntuación | Reportes, resúmenes, código, conversación |
| Probabilidades | Parte central de la respuesta | No siempre disponibles o calibradas |
| Integración | Directa en flujos de software | Requiere validar la generación |
| Explicación | No es su objetivo principal | Sí |

## Son complementarios

- Jev/KEV clasifica y prioriza.
- Las reglas calculan límites exactos.
- Una persona revisa casos sensibles.
- Un LLM puede redactar el informe final.

---

# 4. ¿Es un “IF con inteligencia artificial”?

Es una analogía útil, con una diferencia importante.

## `if` tradicional

```go
if lossPercent > 0.5 {
    alert()
}
```

- Exacto y reproducible.
- Depende de una condición explícita.
- No interpreta ambigüedad.

## “IF semántico y probabilístico”

```text
¿La diferencia es sospechosa considerando
pesos, tara, tiempo y política?

Probabilidad de sí: 0.91
```

- Interpreta varias señales.
- Expresa incertidumbre.
- Puede equivocarse.
- Necesita umbrales y evaluación.

> Jev no reemplaza las reglas tradicionales; las complementa cuando la decisión requiere interpretar contexto.

---

# 5. Tipos de decisión y aplicaciones

## Primitivas utilizadas

- **`noul`:** probabilidad de sí/no.
- **`choice`:** selección entre categorías definidas.
- **`score`:** puntuación dentro de una escala ordenada.

## Aplicaciones reales

- Enrutar solicitudes de servicio al cliente.
- Priorizar alertas de fraude.
- Clasificar reclamaciones o incidentes.
- Evaluar consistencia documental.
- Detectar riesgos contractuales.
- Moderar contenido y estimar severidad.
- Priorizar excepciones logísticas.

La salida del modelo es una **señal de decisión**, no una verdad comprobada.

---

# 6. Caso de uso: PalmFlow

La demo analiza **1,000 viajes ficticios y reproducibles** de crudo de palma.

## Rutas

- El Triunfo → Zolic.
- Santa Rosa → Zolic.
- Repsa I → Zolic.
- Repsa II → Zolic.

## Se busca detectar

- Merma superior a la política.
- Ganancia de peso inesperada.
- Inconsistencia de báscula.
- Cambio anómalo de tara.
- Demora de tránsito.
- Posibles señales de manipulación.
- Impacto económico en USD.

Los pesos operativos se presentan en toneladas métricas.

---

# 7. Arquitectura y papel de KEV

```text
React + Nginx :3080
        │ REST + SSE
        ▼
     API Go :8080 ─────────► KEV local :8009
        │                    POST /v1/systemone
        ▼
    PostgreSQL
```

## Responsabilidades

**Go**

- Calcula pesos, porcentajes, tiempos e impacto económico.
- Aplica reglas exactas.

**KEV**

- Evalúa anomalía, calidad, riesgo y prioridad.
- Devuelve probabilidades y categorías.

**React**

- Muestra avance, hallazgos, costos y latencia.

KEV es una familia abierta de modelos “Jev-like” compatible con System One. **No es Jev**: permite ejecutar localmente un contrato similar.

---

# 8. El corazón: `state` + `questions`

System One no usa principalmente un prompt conversacional. La solicitud contiene:

```json
{
  "model": "kev-latest",
  "state": {
    "domain": "Palm crude oil bulk transport...",
    "policy": {
      "maximum_expected_loss_percent": 0.5,
      "maximum_expected_transit_overrun_percent": 35,
      "maximum_expected_weight_gain_percent": 0.2,
      "rule": "Calculations are supplied as facts. Do not invent missing facts."
    },
    "trip": {
      "origin": "El Triunfo",
      "destination": "Zolic",
      "net_departure_kg": 28000,
      "net_arrival_kg": 27680,
      "loss_percent": 1.1429,
      "transit_overrun_percent": 24.62
    }
  },
  "questions": {}
}
```

Los cálculos se envían ya resueltos para evitar delegar aritmética exacta al modelo.

---

# 9. Las cinco preguntas enviadas a KEV

```json
{
  "weight_anomaly": {
    "type": "noul",
    "instructions": "Is the measured net weight difference outside the supplied policy or physically suspicious?"
  },
  "data_quality": {
    "type": "choice",
    "instructions": "Classify the record's measurement consistency.",
    "criteria": {
      "consistent": "Arithmetic and measurements are plausible.",
      "weight_gain": "Arrival net weight is suspiciously above departure.",
      "tare_mismatch": "Tare change could materially distort the comparison.",
      "invalid_measurement": "Gross, tare, net, dates, or durations are internally invalid."
    }
  },
  "transit_anomaly": {
    "type": "noul",
    "instructions": "Is travel time materially abnormal under the supplied policy?"
  },
  "fraud_risk": {
    "type": "score",
    "instructions": "Rate evidence of manipulation, diversion, collusion, or weighbridge irregularity. High scores require evidence in this record.",
    "criteria": [
      "No material evidence",
      "Weak signal; monitor",
      "Strong signal; human investigation",
      "Critical evidence; block and investigate"
    ]
  },
  "review_priority": {
    "type": "score",
    "instructions": "Rate operational review priority considering weight, time, data quality, and financial exposure.",
    "criteria": [
      "Routine",
      "Review when convenient",
      "Review today",
      "Immediate escalation"
    ]
  }
}
```

**Buenas prácticas:** preguntas atómicas, categorías claras, escalas ordenadas, política explícita y ninguna invención de datos.

---

# 10. Configuración y decisión híbrida

## Política inicial de la demo

| Parámetro | Valor |
|---|---:|
| Tolerancia de merma | 0.5% |
| Demora máxima | 35% |
| Ganancia máxima de peso | 0.2% |
| Precio | USD 1,050/t |

## Ejecución de KEV

```bash
export KEV_API_KEY=local-demo-key
export KEV_PREFIX_CACHE=8
export KEV_DATE_FACTS=1

python -m kev.serve \
  --run "jaredpalmer/kev-0.8b@v1.0" \
  --host 0.0.0.0 \
  --port 8009
```

## Riesgo combinado

```text
30% anomalía de peso KEV
+ 20% anomalía de tránsito KEV
+ 25% riesgo de fraude KEV
+ 25% reglas determinísticas
```

Errores exactos, como una aritmética de báscula inválida, no dependen de la decisión del modelo.

---

# 11. Demostración en vivo

Abrir **<http://localhost:3080>**.

1. Confirmar que PostgreSQL y KEV están disponibles.
2. Pulsar **Cargar demo (1,000)**.
3. Mostrar pesos de origen y destino.
4. Abrir **Configuración** y explicar los umbrales.
5. Pulsar **Analizar pendientes**.
6. Observar el avance en tiempo real.
7. Filtrar por `critical` o `high`.
8. Abrir un viaje y mostrar:
   - Merma y reglas activadas.
   - Probabilidades KEV.
   - Riesgo y prioridad.
   - Impacto económico.
   - Tokens y latencia.

**Frase durante la demo:** “Go calcula los hechos; KEV interpreta el patrón; el sistema combina ambos resultados”.

---

# 12. Cierre, costos y límites

## Comparación ilustrativa

- **KEV local:** USD 0 de API incremental; sí existen costos de hardware, energía y operación.
- **Jev 1.14:** USD 0.042 por millón de tokens de entrada; salida gratuita según la tarifa consultada.
- **GPT-6.1 Sol:** escenario de USD 2/M tokens de entrada y USD 10/M de salida.

La latencia de KEV es real. Las referencias de Jev y GPT son simuladas; la demo no llama a sus APIs.

## Lo que no debe afirmarse

- Una probabilidad alta no prueba fraude.
- KEV-0.8B no tiene la misma precisión que Jev.
- Los umbrales de demo no son políticas de producción.
- No debe automatizarse sin datos reales etiquetados y evaluación.

## Mensaje final

> Un LLM tradicional está diseñado principalmente para generar.  
> Jev y KEV están diseñados para decidir dentro de un espacio controlado.

PalmFlow combina:

- Reglas para lo que debe ser exacto.
- KEV para lo que requiere juicio.
- Revisión humana para lo que implica riesgo.

---

# Fuentes

- System One: <https://docs.typesafe.ai/concepts/system-one>
- Modelos Jev: <https://docs.typesafe.ai/models>
- KEV: <https://github.com/jaredpalmer/kev>
- Implementación: `backend/analysis.go`, `backend/store.go` y `scripts/start-kev.sh`.
