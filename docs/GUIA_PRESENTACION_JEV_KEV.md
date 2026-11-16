# Jev, System One y KEV — guía ampliada
## Decisiones probabilísticas aplicadas al transporte de crudo de palma

> Presentación ejecutiva y técnica — PalmFlow  
> Duración sugerida: **15 a 20 minutos**, incluyendo la demostración.

---

# 1. El problema: convertir información en decisiones

En una operación logística recibimos datos de:

- Básculas.
- Vehículos y conductores.
- Horarios de salida y llegada.
- Distancias y rutas.
- Políticas operativas.
- Precios e impacto económico.

El reto no es solamente almacenar estos datos. El reto es responder, de forma rápida y consistente:

1. ¿La diferencia de peso es normal o sospechosa?
2. ¿El registro es confiable?
3. ¿El tiempo de tránsito es anómalo?
4. ¿Existe suficiente evidencia para investigar?
5. ¿Qué caso debe revisar primero una persona?

**Mensaje principal:** necesitamos decisiones que el software pueda consumir, no únicamente texto generado.

### Nota para el presentador

Abrir con el problema de negocio. No iniciar hablando de modelos o infraestructura. Jev tiene sentido cuando primero se entiende la decisión que debe integrarse al proceso.

---

# 2. ¿Qué es Jev?

**Jev** es el modelo principal de TypeSafe y el primer modelo de la categoría denominada **System One**.

Un modelo System One:

- Recibe un **estado**: datos, contexto, políticas y evidencia.
- Evalúa una o varias preguntas concretas.
- Devuelve **respuestas tipadas**, probabilidades y niveles de confianza.
- Está diseñado para realizar juicios rápidos que otro sistema pueda utilizar directamente.

En lugar de pedir:

> “Analiza este viaje y escribe un reporte.”

se define:

> “Con estos datos y esta política, ¿existe una anomalía de peso?”

y se recibe una salida estructurada como:

```json
{
  "type": "noul",
  "noul": 0.91
}
```

Esto representa una probabilidad de respuesta afirmativa, no una explicación narrativa.

### ¿Por qué “System One”?

El nombre hace referencia al pensamiento rápido e intuitivo popularizado por Daniel Kahneman. El modelo está orientado a **juicios focalizados y rápidos**, no a desarrollar largas cadenas de razonamiento o redactar contenido.

---

# 3. Jev frente a un LLM conversacional

| Aspecto | Jev / System One | LLM conversacional, como ChatGPT |
|---|---|---|
| Propósito principal | Tomar decisiones estructuradas | Generar, transformar o explicar texto |
| Salida | Tipada y restringida | Texto libre o JSON solicitado por prompt |
| Ejemplos | Sí/no probabilístico, categoría, puntuación | Informe, resumen, explicación, código |
| Probabilidades | Parte central del contrato | No siempre disponibles o calibradas para la decisión |
| Integración | Directa en reglas y flujos de software | Requiere validar y controlar la respuesta generada |
| Variabilidad de salida | Espacio de respuesta predefinido | Espacio de respuesta abierto |
| Explicación narrativa | No es su objetivo | Sí |
| Uso recomendado | Clasificar, priorizar, enrutar, detectar | Conversar, redactar, resumir, razonar y crear |

## No compiten necesariamente

Pueden utilizarse juntos:

1. **Jev/KEV** clasifica y prioriza miles de registros.
2. Las reglas del sistema aplican límites exactos.
3. Una persona revisa los casos inciertos o críticos.
4. Un LLM puede redactar el informe final de los casos seleccionados.

**Mensaje principal:** usar un modelo generativo para cada decisión puede ser más abierto, costoso y difícil de controlar de lo necesario.

---

# 4. ¿Es un “IF con inteligencia artificial”?

Es una buena analogía inicial, pero requiere una precisión importante.

## Un `if` tradicional

```go
if lossPercent > 0.5 {
    alert()
}
```

- Es exacto.
- Es reproducible.
- Depende de una condición explícita.
- No interpreta contexto ambiguo.

## Una condición semántica probabilística

```text
¿La diferencia de peso es sospechosa considerando
la política, los pesajes, la tara y el patrón operativo?

Probabilidad de sí: 0.91
```

- Puede interpretar evidencia y lenguaje natural.
- Considera varias señales al mismo tiempo.
- Expresa incertidumbre.
- Puede equivocarse.
- Necesita umbrales, evaluación y supervisión.

## La definición más útil

> Jev puede verse como un **“IF semántico y probabilístico”**: el desarrollador define la pregunta y las respuestas permitidas; el modelo estima cuál corresponde y con qué probabilidad.

No sustituye los `if` tradicionales. Los complementa cuando la condición no puede describirse de forma completa con una fórmula rígida.

---

# 5. Primitivas de decisión

Jev y KEV utilizan preguntas tipadas. En esta demo usamos tres:

## `noul`: decisión binaria probabilística

Pregunta:

> ¿Existe una anomalía de peso?

Respuesta:

```json
{"noul": 0.91}
```

`0` se acerca a “no” y `1` se acerca a “sí”.

## `choice`: selección entre categorías

Pregunta:

> ¿Cómo se clasifica la calidad del registro?

Opciones:

- `consistent`
- `weight_gain`
- `tare_mismatch`
- `invalid_measurement`

La respuesta incluye la opción más probable y la distribución entre opciones.

## `score`: nivel dentro de una escala ordenada

Pregunta:

> ¿Cuál es el nivel de riesgo de fraude?

Escala:

0. Sin evidencia material.
1. Señal débil; monitorear.
2. Señal fuerte; investigar.
3. Evidencia crítica; bloquear e investigar.

**Ventaja:** el contrato de salida ya es conocido por la aplicación.

---

# 6. Aplicaciones en la vida real

## Operaciones y logística

- Priorizar viajes para revisión.
- Detectar documentos o mediciones inconsistentes.
- Clasificar incidentes y excepciones.
- Evaluar señales combinadas de desvío o manipulación.

## Servicio al cliente

- Enrutar solicitudes al departamento correcto.
- Identificar intención de cancelación o reembolso.
- Medir urgencia, frustración o riesgo de abandono.

## Finanzas y seguros

- Priorizar alertas de fraude.
- Clasificar reclamaciones.
- Evaluar consistencia documental.
- Escalar operaciones que necesitan revisión humana.

## Legal y cumplimiento

- Detectar presencia de cláusulas.
- Clasificar riesgos contractuales.
- Priorizar expedientes.
- Identificar posibles incumplimientos de una política.

## Seguridad y moderación

- Clasificar contenido.
- Estimar severidad.
- Decidir si un caso debe bloquearse, permitirse o revisarse.

> La decisión del modelo no debe confundirse con una verdad comprobada. En dominios sensibles debe formar parte de un flujo con controles, auditoría y revisión humana.

---

# 7. Nuestro caso: PalmFlow

La demo analiza **1,000 viajes ficticios y reproducibles** de crudo de palma.

## Orígenes

- El Triunfo.
- Santa Rosa.
- Repsa I.
- Repsa II.

## Destino

- Zolic.

## Información por viaje

- Vehículo y conductor.
- Fecha y hora de salida y llegada.
- Peso bruto, tara y peso neto en origen.
- Peso bruto, tara y peso neto en destino.
- Distancia.
- Tiempo esperado y tiempo real.
- Precio del aceite en `USD/t`.

## Objetivo

Detectar:

- Mermas superiores a la política.
- Ganancias de peso inesperadas.
- Inconsistencias de báscula.
- Cambios anómalos de tara.
- Tiempos de tránsito anormales.
- Señales que ameritan investigación.
- Impacto económico estimado.

---

# 8. Arquitectura de la demostración

```text
Navegador
   │
   ▼
React + Nginx :3080
   │ REST + SSE
   ▼
API Go :8080 ───────────────► KEV local :8009
   │                           POST /v1/systemone
   ▼
PostgreSQL
```

## Separación de responsabilidades

### Go: hechos y reglas exactas

- Calcula la merma.
- Calcula porcentajes.
- Verifica `bruto - tara = neto`.
- Calcula diferencias de tiempo.
- Calcula impacto económico.
- Aplica reglas determinísticas y severidad mínima.

### KEV: decisiones probabilísticas

- Anomalía de peso.
- Calidad de datos.
- Anomalía de tránsito.
- Riesgo de fraude.
- Prioridad de revisión.

### React

- Muestra avance en tiempo real.
- Permite filtrar hallazgos.
- Presenta reglas, probabilidades, costos y latencia.

**Diseño clave:** el modelo no realiza los cálculos financieros ni sustituye las validaciones matemáticas.

---

# 9. ¿Qué es KEV y por qué lo usamos?

**KEV** es una familia abierta de modelos de decisión inspirados en Jev y compatibles con el contrato de System One.

No es Jev:

- **Jev** es el modelo hospedado de TypeSafe.
- **KEV** es una implementación “Jev-like” que puede ejecutarse y ajustarse localmente.

## Configuración utilizada en la demo

| Elemento | Valor |
|---|---|
| Código KEV | Tag `kev-1.0` |
| Modelo | `jaredpalmer/kev-0.8b@v1.0` |
| Modelo base | `Qwen/Qwen3.5-0.8B-Base` |
| Dispositivo | NVIDIA CUDA |
| Precisión | BF16 |
| Endpoint | `http://localhost:8009/v1/systemone` |
| Autenticación local | Bearer token |
| Memoria observada | Aproximadamente 4.47 GB de VRAM |

Se eligió KEV-0.8B por los **8 GB de VRAM** disponibles. Su tamaño facilita la demo local, pero tiene menor capacidad que KEV-4B, KEV-9B o Jev.

## Razones para ejecutar KEV localmente

- Los datos no se envían a Jev ni a OpenAI.
- No existe costo incremental por llamada a una API externa.
- Se controla el modelo, la versión y la infraestructura.
- Se conserva el mismo estilo de integración `POST /v1/systemone`.

---

# 10. El corazón de KEV: `state` + `questions`

En System One no se trabaja principalmente con un “prompt de conversación”.

El contrato se divide en:

## 1. `state`

Contiene los hechos que el modelo debe evaluar:

- Dominio.
- Política vigente.
- Datos del viaje.
- Cálculos realizados por Go.

## 2. `questions`

Define cada decisión:

- Tipo de respuesta.
- Instrucción.
- Opciones o escala permitida.

## 3. `model`

Selecciona el modelo compatible que atenderá la solicitud.

```json
{
  "model": "kev-latest",
  "state": {},
  "questions": {}
}
```

**Principio de diseño:** enviar hechos claros, preguntas atómicas y criterios explícitos.

---

# 11. Estado enviado por PalmFlow

Ejemplo simplificado:

```json
{
  "domain": "Palm crude oil bulk transport with independent weighbridges at origin and destination.",
  "policy": {
    "maximum_expected_loss_percent": 0.5,
    "maximum_expected_transit_overrun_percent": 35,
    "maximum_expected_weight_gain_percent": 0.2,
    "rule": "Calculations are supplied as facts. Judge the operational pattern and evidence. Do not invent missing facts."
  },
  "trip": {
    "id": "TRIP-0001",
    "vehicle": "C-104",
    "driver": "Conductor 12",
    "origin": "El Triunfo",
    "destination": "Zolic",
    "net_departure_kg": 28000,
    "net_arrival_kg": 27680,
    "loss_kg": 320,
    "loss_percent": 1.1429,
    "tare_difference_kg": 30,
    "distance_km": 305,
    "expected_hours": 6.5,
    "actual_hours": 8.1,
    "transit_overrun_percent": 24.62
  }
}
```

## ¿Por qué los cálculos se envían ya resueltos?

- Evita delegar aritmética exacta al modelo.
- Reduce ambigüedad.
- Facilita auditoría.
- Permite comparar la respuesta del modelo contra reglas conocidas.
- Mantiene una única fuente de verdad para el impacto económico.

Aunque la interfaz muestra toneladas, el backend conserva kilogramos internamente para compatibilidad y precisión.

---

# 12. Preguntas exactas de la demo

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

## Reglas para diseñar buenas preguntas

1. Una pregunta debe representar una sola decisión.
2. Las categorías deben ser claras y distinguibles.
3. La escala debe estar ordenada de menor a mayor.
4. La política debe estar dentro del estado o los criterios.
5. Debe indicarse que no se inventen datos ausentes.
6. La salida debe corresponder a una acción posible del sistema.

---

# 13. Configuración: tres niveles distintos

## Nivel 1: política del negocio

Configurable desde PalmFlow:

| Parámetro | Valor inicial de demo |
|---|---:|
| Tolerancia máxima de merma | 0.5% |
| Demora máxima sobre el tiempo esperado | 35% |
| Ganancia máxima de peso tolerada | 0.2% |
| Precio de aceite | USD 1,050/t |
| Salida estimada para comparación GPT | 220 tokens |
| Latencia de referencia Jev | 100 ms |
| Latencia de referencia GPT | 8,311 ms |

Estos valores son supuestos de demostración, no recomendaciones operativas.

## Nivel 2: contrato de decisión

Se configura en el backend:

- Campos incluidos en `state`.
- Instrucciones.
- Criterios.
- Tipo de cada pregunta.
- Combinación de probabilidades y reglas.

## Nivel 3: ejecución de KEV

```bash
export KEV_API_KEY=local-demo-key
export KEV_PREFIX_CACHE=8
export KEV_DATE_FACTS=1
export HF_HUB_DISABLE_XET=1

python -m kev.serve \
  --run "jaredpalmer/kev-0.8b@v1.0" \
  --host 0.0.0.0 \
  --port 8009
```

Aquí se configura:

- Modelo y versión.
- Puerto.
- Autenticación.
- Caché.
- Ayuda para hechos relacionados con fechas.
- Mecanismo de descarga.

---

# 14. ¿Cómo se combinan reglas y probabilidades?

PalmFlow utiliza un enfoque híbrido.

## Señales determinísticas

- Peso no positivo.
- Inconsistencia entre bruto, tara y neto.
- Merma superior a la tolerancia.
- Ganancia de peso inesperada.
- Demora superior al límite.
- Línea de tiempo inválida.
- Cambio de tara mayor a 0.5 t.
- Distancia o tiempo esperado inválidos.

## Puntuación combinada

```text
riesgo =
  30% × anomalía de peso KEV
  + 20% × anomalía de tránsito KEV
  + 25% × riesgo de fraude KEV
  + 25% × riesgo determinístico
```

Después se aplican límites de severidad:

- `normal`
- `warning`
- `high`
- `critical`

Algunas inconsistencias exactas, como una línea de tiempo inválida o una aritmética de báscula incorrecta, producen severidad crítica sin depender de que el modelo las reconozca.

**Mensaje principal:** KEV aporta juicio probabilístico; Go mantiene las garantías verificables.

---

# 15. Demostración en vivo

Abrir:

```text
http://localhost:3080
```

## Paso 1 — Estado de los servicios

- Confirmar PostgreSQL disponible.
- Confirmar KEV disponible.
- Aclarar que KEV se ejecuta localmente.

## Paso 2 — Cargar los datos

- Pulsar **Cargar demo (1,000)**.
- Mostrar los cuatro orígenes y Zolic como destino.
- Mostrar pesos en toneladas métricas.
- Explicar que los casos son ficticios y reproducibles.

## Paso 3 — Mostrar la política

- Abrir **Configuración**.
- Mostrar tolerancia de merma, demora, ganancia de peso y precio.
- Cambiar un valor solamente si se desea demostrar el efecto de la política.

## Paso 4 — Ejecutar

- Pulsar **Analizar pendientes**.
- Mostrar el progreso recibido mediante SSE.
- Explicar que se envían cinco preguntas en una sola solicitud por viaje.

## Paso 5 — Investigar un caso

- Filtrar por `critical` o `high`.
- Abrir el detalle de un viaje.
- Mostrar:
  - Merma.
  - Reglas activadas.
  - Probabilidades KEV.
  - Riesgo combinado.
  - Prioridad de revisión.
  - Impacto económico.
  - Tokens y latencia.

---

# 16. Costos y latencia mostrados en la demo

## KEV local

```text
costo API incremental = USD 0
```

No significa costo total cero. Deben considerarse:

- Electricidad.
- Hardware.
- Operación.
- Monitoreo.
- Mantenimiento.
- Evaluación y ajuste.

## Jev 1.14

Tarifa utilizada por la demo, consultada el 16 de noviembre de 2026:

```text
costo = tokens de entrada / 1,000,000 × USD 0.042
```

La salida se considera gratuita de acuerdo con la tarifa publicada.

## GPT-6.1 Sol

Escenario comparativo configurado:

```text
costo =
  tokens de entrada / 1,000,000 × USD 2
  + tokens de salida / 1,000,000 × USD 10
```

## Aclaración obligatoria

- La latencia de KEV mostrada es real.
- Las latencias de Jev y GPT son referencias simuladas.
- La demo no realiza llamadas reales a Jev ni a OpenAI.
- La comparación es ilustrativa, no un benchmark controlado.

---

# 17. Qué demuestra y qué no demuestra el proyecto

## Sí demuestra

- Cómo integrar una decisión tipada dentro de software real.
- Cómo ejecutar KEV localmente.
- Cómo combinar reglas exactas con probabilidades.
- Cómo analizar registros en tiempo real.
- Cómo mantener trazabilidad de datos, decisiones y costos.
- Cómo conservar un contrato compatible con System One.

## No demuestra

- Que una probabilidad alta pruebe fraude.
- Que KEV-0.8B tenga la misma precisión que Jev.
- Que la tolerancia de 0.5% sea correcta para producción.
- Que las latencias comparativas sean un SLA.
- Que el modelo pueda operar sin evaluación humana.
- Que una demo con datos ficticios sea suficiente para automatizar decisiones reales.

> El resultado es una señal para priorizar revisión, no una acusación ni una sentencia automática.

---

# 18. Siguiente fase recomendada

1. Obtener datos reales anonimizados.
2. Definir políticas oficiales por ruta, producto, clima y vehículo.
3. Crear etiquetas de revisión humana.
4. Medir precisión, falsos positivos, falsos negativos y Brier score.
5. Evaluar calibración por cada pregunta.
6. Comparar KEV-0.8B, KEV-4B y Jev con el mismo conjunto.
7. Definir umbrales:
   - Automatizar.
   - Revisar.
   - Escalar.
8. Agregar autenticación, auditoría y exportación de expedientes.
9. Considerar ajuste fino de KEV cuando exista suficiente información etiquetada.

---

# 19. Cierre

## Idea para recordar

> Un LLM tradicional está diseñado principalmente para **generar**.  
> Jev y KEV están diseñados para **decidir dentro de un espacio controlado**.

## Propuesta de valor de PalmFlow

- Las reglas calculan lo que debe ser exacto.
- KEV interpreta lo que requiere juicio.
- El sistema combina ambas señales.
- La persona conserva el control sobre los casos sensibles.

## En una frase

> PalmFlow convierte datos de transporte en decisiones estructuradas, probabilísticas, auditables y listas para integrarse a una operación.

---

# Apéndice A — Ejemplo completo de solicitud

```json
{
  "model": "kev-latest",
  "state": {
    "domain": "Palm crude oil bulk transport with independent weighbridges at origin and destination.",
    "policy": {
      "maximum_expected_loss_percent": 0.5,
      "maximum_expected_transit_overrun_percent": 35,
      "maximum_expected_weight_gain_percent": 0.2,
      "rule": "Calculations are supplied as facts. Judge the operational pattern and evidence. Do not invent missing facts."
    },
    "trip": {
      "id": "TRIP-0001",
      "vehicle": "C-104",
      "driver": "Conductor 12",
      "origin": "El Triunfo",
      "destination": "Zolic",
      "departure_utc": "2026-11-10T08:00:00Z",
      "arrival_utc": "2026-11-10T16:06:00Z",
      "net_departure_kg": 28000,
      "net_arrival_kg": 27680,
      "gross_departure_kg": 39000,
      "tare_departure_kg": 11000,
      "gross_arrival_kg": 38710,
      "tare_arrival_kg": 11030,
      "loss_kg": 320,
      "loss_percent": 1.1429,
      "tare_difference_kg": 30,
      "distance_km": 305,
      "expected_hours": 6.5,
      "actual_hours": 8.1,
      "transit_overrun_percent": 24.62
    }
  },
  "questions": {
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
}
```

---

# Apéndice B — Fuentes

- System One: <https://docs.typesafe.ai/concepts/system-one>
- Modelos y precios de Jev: <https://docs.typesafe.ai/models>
- KEV: <https://github.com/jaredpalmer/kev>
- Modelo KEV-0.8B: <https://huggingface.co/jaredpalmer/kev-0.8b>
- Implementación de la demo:
  - `backend/analysis.go`
  - `backend/store.go`
  - `scripts/start-kev.sh`
  - `docs/DEMO_GUIDE.md`
