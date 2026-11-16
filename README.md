# PalmFlow — demo local de System One con KEV

Demo técnica y funcional para analizar transportes de crudo de palma entre dos básculas. Combina:

- **Reglas determinísticas** para mermas, consistencia aritmética, tiempos e impacto económico.
- **KEV 0.8B local** para decisiones tipadas y probabilísticas compatibles con `POST /v1/systemone`.
- **Go + PostgreSQL** para carga CSV, persistencia, análisis y eventos en tiempo real.
- **React** para visualizar operaciones, anomalías, riesgo, costos y latencia.
- Comparación **simulada** de costo y tiempo frente a **Jev 1.14** y **GPT-6.1 Sol**.

## Requisitos verificados para este equipo

- Linux x86_64
- NVIDIA RTX 5050, 8 GB VRAM
- Docker y Docker Compose
- Aproximadamente 28 GB libres antes de descargar modelos e imágenes

KEV se ejecuta directamente en Linux con CUDA. La aplicación y PostgreSQL se ejecutan en Docker.

## 1. Instalar KEV

```bash
make install-kev
```

El script fija el código de KEV al tag `kev-1.0` e instala Python 3.13 mediante `uv`. No usa `pip install kev`, porque ese nombre pertenece a otro paquete.

## 2. Iniciar KEV

```bash
make start-kev
```

La primera ejecución descarga `jaredpalmer/kev-0.8b@v1.0` y su base Qwen. El endpoint queda en:

```text
http://localhost:8009/v1/systemone
```

Prueba:

```bash
curl -s http://localhost:8009/v1/models \
  -H "Authorization: Bearer local-demo-key"
```

También puedes ejecutar una decisión completa de prueba:

```bash
make test-kev
```

## 3. Iniciar la aplicación

En otra terminal:

```bash
make up
```

Abre [http://localhost:3080](http://localhost:3080).

1. Pulsa **Cargar demo (1,000)**.
2. Pulsa **Analizar pendientes**.
3. Observa el avance, resultados, riesgos y comparación de costos en tiempo real.

El guion para presentar la demo, la arquitectura y las fórmulas están en
[`docs/DEMO_GUIDE.md`](docs/DEMO_GUIDE.md).

## CSV

La interfaz permite descargar una plantilla. Columnas requeridas:

```text
trip_id, vehicle, driver, origin, destination,
departure_at, arrival_at,
gross_out_t, tare_out_t, net_out_t,
gross_in_t, tare_in_t, net_in_t,
distance_km, expected_hours, actual_hours
```

`departure_at` y `arrival_at` usan RFC 3339. `oil_price_usd_per_t` es opcional.
La importación también acepta las columnas históricas terminadas en `_kg` y
`oil_price_usd_per_kg` para mantener compatibilidad.

La demo genera 1,000 viajes ficticios y reproducibles desde **El Triunfo**,
**Santa Rosa**, **Repsa I** y **Repsa II**, siempre con destino **Zolic**.

## Reglas y decisiones

Las reglas exactas detectan:

- Merma superior al umbral configurable, inicialmente **0.5%**.
- Ganancia de peso inesperada.
- Inconsistencia `gross - tare ≠ net`.
- Cambio de tara mayor a 0.5 t.
- Demora superior al porcentaje configurado.
- Fechas, pesos o métricas de ruta inválidos.

KEV recibe los valores calculados y responde cinco decisiones tipadas:

- Anomalía de peso (`noul`)
- Calidad del registro (`choice`)
- Anomalía de tiempo (`noul`)
- Riesgo de fraude (`score`)
- Prioridad de revisión (`score`)

Los cálculos financieros nunca se delegan al modelo.

## Comparación de costos

Tarifas consultadas el **16 de noviembre de 2026**:

- **Jev 1.14:** USD 0.042 por millón de tokens de entrada; salida gratuita.
- **GPT-6.1 Sol:** USD 2 por millón de tokens de entrada y USD 10 por millón de tokens de salida.
- **KEV local:** USD 0 de API incremental. No incluye amortización, electricidad ni mantenimiento.

La UI usa los tokens reales reportados por KEV como aproximación de entrada equivalente. Para GPT, los tokens de salida son un escenario configurable (220 por registro inicialmente). Las latencias de Jev/GPT son referencias simuladas y se identifican como tales.

## Limitaciones importantes

- KEV-0.8B declara **inglés** como idioma; el backend construye preguntas en inglés y traduce la presentación al español.
- Es una demo de apoyo a revisión humana, no una prueba de fraude ni un sistema listo para decisiones automáticas.
- Las probabilidades deben calibrarse con datos reales etiquetados antes de producción.
- KEV-0.8B se eligió por el hardware disponible. KEV-4B ofrece mejor precisión, pero requiere más memoria.
- El valor económico depende del precio por tonelada configurado y no incluye seguros, impuestos ni costos indirectos.

## Apagar

```bash
make down
```

KEV se detiene con `Ctrl+C` en su terminal.

## Fuentes

- KEV: <https://github.com/jaredpalmer/kev>
- Modelo KEV-0.8B: <https://huggingface.co/jaredpalmer/kev-0.8b>
- System One: <https://docs.typesafe.ai/concepts/system-one>
- Precios Jev: <https://docs.typesafe.ai/models>
- Precios OpenAI: <https://developers.openai.com/api/docs/pricing>
# PalmFlow
