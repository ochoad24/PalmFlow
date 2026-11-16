import React, { useCallback, useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 4 });
const number = new Intl.NumberFormat("es-CO", { maximumFractionDigits: 2 });
const tonnes = new Intl.NumberFormat("es-CO", { minimumFractionDigits: 2, maximumFractionDigits: 3 });
const tinyMoney = (v) => v < 0.01 ? `$${Number(v || 0).toFixed(6)}` : money.format(v || 0);
const weight = (kg) => `${tonnes.format(Number(kg || 0) / 1000)} t`;

async function api(path, options) {
  const response = await fetch(path, options);
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error || `HTTP ${response.status}`);
  return body;
}

function App() {
  const [dashboard, setDashboard] = useState(null);
  const [shipments, setShipments] = useState([]);
  const [health, setHealth] = useState(null);
  const [progress, setProgress] = useState({ status: "idle", percent: 0 });
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [filter, setFilter] = useState("all");
  const [configOpen, setConfigOpen] = useState(false);
  const [selected, setSelected] = useState(null);

  const refresh = useCallback(async () => {
    const [d, s, h] = await Promise.all([
      api("/api/dashboard"), api("/api/shipments?limit=1000"), api("/api/health")
    ]);
    setDashboard(d); setShipments(s || []); setHealth(h);
  }, []);

  useEffect(() => { refresh().catch(e => setMessage(e.message)); }, [refresh]);
  useEffect(() => {
    const events = new EventSource("/api/events");
    events.onmessage = ({ data }) => {
      const next = JSON.parse(data);
      setProgress(next);
      if (next.status === "completed" || next.status === "error") refresh().catch(() => {});
      else if (next.processed && next.processed % 20 === 0) refresh().catch(() => {});
    };
    return () => events.close();
  }, [refresh]);

  const action = async (fn, success) => {
    setBusy(true); setMessage("");
    try { await fn(); setMessage(success); await refresh(); }
    catch (e) { setMessage(e.message); }
    finally { setBusy(false); }
  };

  const resetDemo = () => action(
    () => api("/api/demo/reset", { method: "POST" }),
    "Se cargaron 1,000 transportes ficticios reproducibles hacia Zolic."
  );
  const runAnalysis = () => action(
    () => api("/api/analysis/run", { method: "POST" }),
    "Análisis iniciado. El tablero se actualizará en tiempo real."
  );
  const upload = (file) => {
    if (!file) return;
    const data = new FormData(); data.append("file", file);
    action(() => api("/api/import", { method: "POST", body: data }), "Archivo cargado correctamente.");
  };

  const visible = useMemo(() => shipments.filter(s =>
    filter === "all" || (filter === "pending" ? !s.analysis : s.analysis?.severity === filter)
  ), [shipments, filter]);

  if (!dashboard) return <div className="loading"><span className="pulse" /> Preparando PalmFlow…</div>;
  const d = dashboard;

  return (
    <main>
      <header className="topbar">
        <div>
          <div className="eyebrow">SYSTEM ONE · DEMO LOCAL</div>
          <h1>Palm<span>Flow</span></h1>
          <p>Auditoría probabilística del transporte de crudo de palma</p>
        </div>
        <div className="statusRow">
          <Status label="PostgreSQL" online={health?.database === "online"} />
          <Status label="KEV 0.8B" online={health?.kev === "online"} />
          <button className="iconButton" onClick={() => setConfigOpen(true)}>⚙ Configuración</button>
        </div>
      </header>

      {health?.kev !== "online" && (
        <div className="warningBanner"><b>KEV no está activo.</b> Ejecuta <code>make start-kev</code> antes de analizar. {health?.kevError}</div>
      )}
      {message && <div className="message" onClick={() => setMessage("")}>{message}</div>}

      <section className="hero">
        <div>
          <div className="eyebrow">CONTROL OPERACIONAL</div>
          <h2>De la báscula a una decisión<br/><span>auditable en tiempo real.</span></h2>
          <p>Reglas exactas calculan mermas y consistencia. KEV clasifica riesgo, fraude, calidad y prioridad con probabilidades calibradas.</p>
        </div>
        <div className="actions">
          <button onClick={resetDemo} disabled={busy}>↻ Cargar demo (1,000)</button>
          <label className="button secondary">↑ Importar CSV<input type="file" accept=".csv" onChange={e => upload(e.target.files[0])}/></label>
          <button className="primary" onClick={runAnalysis} disabled={busy || health?.kev !== "online"}>▶ Analizar pendientes</button>
          <a className="download" href="/api/demo.csv">Descargar plantilla CSV</a>
        </div>
      </section>

      {(progress.status === "running" || progress.status === "completed" || progress.status === "error") && (
        <section className={`progressPanel ${progress.status}`}>
          <div><b>{progress.status === "running" ? "KEV analizando" : progress.status === "completed" ? "Análisis completado" : "Análisis detenido"}</b>
            <span>{progress.tripId || progress.error}</span></div>
          <div className="progressTrack"><i style={{width:`${progress.percent || 0}%`}}/></div>
          <strong>{progress.processed || 0}/{progress.total || 0}</strong>
        </section>
      )}

      <section className="kpis">
        <KPI label="Registros" value={number.format(d.totalRecords)} detail={`${d.analyzedRecords} analizados`} />
        <KPI label="Con anomalías" value={number.format(d.anomalyRecords)} detail={`${d.criticalRecords} críticos`} tone="amber" />
        <KPI label="Merma detectada" value={weight(d.totalLossKg)} detail={`Promedio ${number.format(d.avgLossPct)}%`} />
        <KPI label="Impacto económico" value={money.format(d.financialImpactUsd)} detail={`Precio configurable por tonelada`} tone="red" />
      </section>

      <section className="comparison">
        <div className="sectionTitle"><div><span>COMPARADOR DE EJECUCIÓN</span><h3>Local vs. APIs propietarias</h3></div>
          <small>Estimación sobre {d.analyzedRecords} registros · tarifas 16 nov 2026</small></div>
        <div className="modelCards">
          <ModelCard name="KEV 0.8B" type="LOCAL · REAL" cost={tinyMoney(d.kevTotalCostUsd)}
            time={`${number.format(d.avgKevLatencyMs)} ms/registro · ${number.format(d.avgKevLatencyMs*d.analyzedRecords/1000)} s total`} detail="Costo API incremental $0 · usa tu RTX 5050" accent />
          <ModelCard name="Jev 1.14" type="TYPESAFE · SIMULADO" cost={tinyMoney(d.jevTotalCostUsd)}
            time={`≈ ${number.format(d.jevTotalTimeMs/1000)} s total`} detail="$0.042/M tokens entrada · salida gratis" />
          <ModelCard name="GPT-6.1 Sol" type="OPENAI · SIMULADO" cost={tinyMoney(d.gptTotalCostUsd)}
            time={`≈ ${number.format(d.gptTotalTimeMs/1000)} s total`} detail="$2/M entrada + $10/M salida" />
        </div>
        <p className="disclaimer">La latencia de KEV es medida. Jev usa referencia ≈100 ms; GPT usa 8,311 ms de un benchmark público de 14 preguntas. Son escenarios comparativos, no garantías del proveedor.</p>
      </section>

      <section className="records">
        <div className="sectionTitle">
          <div><span>TRAZABILIDAD</span><h3>Transportes y hallazgos</h3></div>
          <div className="filters">{["all","critical","high","warning","normal","pending"].map(x =>
            <button className={filter===x?"active":""} onClick={() => setFilter(x)} key={x}>{label(x)}</button>)}</div>
        </div>
        <div className="tableWrap">
          <table>
            <thead><tr><th>Viaje</th><th>Ruta / vehículo</th><th>Salida</th><th>Entrada</th><th>Diferencia</th><th>Tiempo</th><th>Riesgo</th><th>Comparación IA</th><th>Estado</th></tr></thead>
            <tbody>{visible.map(x => <ShipmentRow key={x.id} x={x} onOpen={() => setSelected(x)}/>)}</tbody>
          </table>
          {!visible.length && <div className="empty">No hay registros en este filtro.</div>}
        </div>
      </section>

      {configOpen && <ConfigModal config={d.config} close={() => setConfigOpen(false)}
        save={(c) => action(() => api("/api/config",{method:"PUT",headers:{"Content-Type":"application/json"},body:JSON.stringify(c)}),
          "Configuración actualizada. Los próximos análisis usarán estos valores.").then(() => setConfigOpen(false))}/>}
      {selected && <DetailModal shipment={selected} close={() => setSelected(null)}/>}
    </main>
  );
}

function Status({ label, online }) { return <div className="status"><i className={online ? "on" : ""}/>{label}</div>; }
function KPI({ label, value, detail, tone="" }) { return <article className={`kpi ${tone}`}><span>{label}</span><strong>{value}</strong><small>{detail}</small></article>; }
function ModelCard({name,type,cost,time,detail,accent}) { return <article className={`modelCard ${accent?"accent":""}`}><span>{type}</span><h4>{name}</h4><b>{cost}</b><strong>{time}</strong><small>{detail}</small></article>; }
function ShipmentRow({x,onOpen}) {
  const a=x.analysis; const diff=x.netOutKg-x.netInKg;
  return <tr onClick={onOpen} className="clickable">
    <td><b>{x.tripId}</b><small>{x.driver}</small></td>
    <td>{x.origin} → {x.destination}<small>{x.vehicle} · {number.format(x.distanceKm)} km</small></td>
    <td>{weight(x.netOutKg)}<small>Tara {weight(x.tareOutKg)}</small></td>
    <td>{weight(x.netInKg)}<small>Tara {weight(x.tareInKg)}</small></td>
    <td className={diff < 0 ? "gain" : ""}><b>{weight(diff)}</b><small>{a ? `${number.format(a.lossPct)}% · ${money.format(a.financialImpactUsd)}` : "Sin analizar"}</small></td>
    <td>{number.format(x.actualHours)} h<small>Esperado {number.format(x.expectedHours)} h</small></td>
    <td>{a ? <><b>{number.format(a.riskScore*100)}%</b><small>Fraude {number.format(a.fraudProbability*100)}%</small></> : "—"}</td>
    <td>{a ? <><b>KEV {number.format(a.kevLatencyMs)} ms</b><small>Jev {tinyMoney(a.jevCostUsd)} · GPT {tinyMoney(a.gptCostUsd)}</small></> : "—"}</td>
    <td><span className={`badge ${a?.severity || "pending"}`}>{label(a?.severity || "pending")}</span>{a?.deterministicFlags?.length > 0 && <small title={a.deterministicFlags.join(", ")}>{a.deterministicFlags.length} reglas</small>}</td>
  </tr>;
}
function DetailModal({shipment:x,close}) {
  const a=x.analysis; const answers=a?.kevAnswers || {};
  const pct=(v)=>`${number.format((v||0)*100)}%`;
  return <div className="modalBackdrop"><div className="modal detailModal"><button className="close" onClick={close}>×</button>
    <span>EXPEDIENTE DEL VIAJE</span><h3>{x.tripId} · {label(a?.severity || "pending")}</h3>
    <div className="detailGrid">
      <div><small>Ruta</small><b>{x.origin} → {x.destination}</b></div>
      <div><small>Vehículo / conductor</small><b>{x.vehicle} · {x.driver}</b></div>
      <div><small>Merma calculada</small><b>{a ? `${weight(a.lossKg)} · ${number.format(a.lossPct)}%` : "Pendiente"}</b></div>
      <div><small>Impacto</small><b>{a ? money.format(a.financialImpactUsd) : "—"}</b></div>
    </div>
    {a ? <>
      <h4>Decisiones de KEV</h4>
      <div className="decisionGrid">
        <Decision label="Anomalía de peso" value={pct(answers.weight_anomaly?.noul)} />
        <Decision label="Anomalía de tiempo" value={pct(answers.transit_anomaly?.noul)} />
        <Decision label="Calidad del dato" value={answers.data_quality?.choice || a.dataQuality} sub={`confianza ${pct(answers.data_quality?.confidence)}`} />
        <Decision label="Riesgo de fraude" value={pct(a.fraudProbability)} sub={`score ${number.format(answers.fraud_risk?.score || 0)}/3`} />
        <Decision label="Prioridad" value={a.reviewPriority} sub={`score ${number.format(answers.review_priority?.score || 0)}/3`} />
        <Decision label="Riesgo combinado" value={pct(a.riskScore)} />
      </div>
      <h4>Reglas exactas activadas</h4>
      <div className="flagList">{a.deterministicFlags?.length ? a.deterministicFlags.map(f=><code key={f}>{f}</code>) : <em>Ninguna</em>}</div>
      <div className="costStrip">
        <div><small>KEV real</small><b>{number.format(a.kevLatencyMs)} ms</b></div>
        <div><small>Tokens entrada / salida</small><b>{a.inputTokens} / {a.outputTokens}</b></div>
        <div><small>Jev simulado</small><b>{tinyMoney(a.jevCostUsd)}</b></div>
        <div><small>GPT simulado</small><b>{tinyMoney(a.gptCostUsd)}</b></div>
      </div>
    </> : <div className="empty">Este registro todavía no ha sido analizado.</div>}
  </div></div>;
}
function Decision({label,value,sub}) { return <div className="decision"><small>{label}</small><b>{value}</b>{sub&&<span>{sub}</span>}</div>; }
function ConfigModal({config,close,save}) {
  const [c,setC]=useState({...config});
  const field=(key,labelText,step="0.1")=><label>{labelText}<input type="number" step={step} value={c[key]} onChange={e=>setC({...c,[key]:Number(e.target.value)})}/></label>;
  const pricePerTonne=<label>Precio del crudo (USD/t)<input type="number" step="1" value={Number(c.defaultOilPriceUsd || 0)*1000} onChange={e=>setC({...c,defaultOilPriceUsd:Number(e.target.value)/1000})}/></label>;
  return <div className="modalBackdrop"><div className="modal"><button className="close" onClick={close}>×</button><span>PARÁMETROS DE DEMO</span><h3>Política de análisis</h3>
    {field("lossTolerancePct","Merma tolerada (%)")}
    {field("maxTransitOverPct","Demora máxima (%)","1")}
    {field("maxWeightGainPct","Ganancia máxima de peso (%)")}
    {pricePerTonne}
    {field("gptOutputTokens","Salida simulada GPT (tokens)","1")}
    <p>Jev y GPT solo se simulan. KEV se ejecuta realmente en el equipo local.</p>
    <div className="modalActions"><button onClick={close}>Cancelar</button><button className="primary" onClick={()=>save(c)}>Guardar</button></div>
  </div></div>;
}
const labels={all:"Todos",critical:"Crítico",high:"Alto",warning:"Alerta",normal:"Normal",pending:"Pendiente"};
const label=(v)=>labels[v]||v;
createRoot(document.getElementById("root")).render(<App />);
