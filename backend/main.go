package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type App struct {
	store   *Store
	kev     *KevClient
	events  *EventHub
	runMu   sync.Mutex
	running bool
}

func main() {
	ctx := context.Background()
	databaseURL := env("DATABASE_URL", "postgres://palm:palm_demo_local@localhost:54329/palm_demo?sslmode=disable")
	var store *Store
	var err error
	for i := 0; i < 30; i++ {
		store, err = NewStore(ctx, databaseURL)
		if err == nil {
			break
		}
		log.Printf("esperando base de datos: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	app := &App{store: store, events: NewEventHub(), kev: &KevClient{
		BaseURL: env("KEV_URL", "http://127.0.0.1:8009"), APIKey: env("KEV_API_KEY", "local-demo-key"),
		HTTP: &http.Client{Timeout: 2 * time.Minute},
	}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.health)
	mux.HandleFunc("GET /api/config", app.getConfig)
	mux.HandleFunc("PUT /api/config", app.putConfig)
	mux.HandleFunc("POST /api/demo/reset", app.resetDemo)
	mux.HandleFunc("GET /api/demo.csv", app.downloadDemo)
	mux.HandleFunc("POST /api/import", app.importCSV)
	mux.HandleFunc("GET /api/shipments", app.shipments)
	mux.HandleFunc("GET /api/dashboard", app.dashboard)
	mux.HandleFunc("POST /api/analysis/run", app.runAnalysis)
	mux.Handle("GET /api/events", app.events)
	handler := withJSON(withCORS(mux))
	addr := ":" + env("PORT", "8080")
	log.Printf("API escuchando en %s; KEV=%s", addr, app.kev.BaseURL)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	kevInfo, err := a.kev.Health(ctx)
	status := "ok"
	kevStatus := "online"
	var kevError string
	if err != nil {
		status = "degraded"
		kevStatus = "offline"
		kevError = err.Error()
	}
	writeJSON(w, 200, map[string]any{"status": status, "database": "online", "kev": kevStatus, "kevError": kevError, "kevInfo": kevInfo,
		"pricing": map[string]any{"jevInputUsdPerMillion": jevInputUSDPerMillion, "gptModel": "gpt-6.1-sol",
			"gptInputUsdPerMillion": gptInputUSDPerMillion, "gptOutputUsdPerMillion": gptOutputUSDPerMillion,
			"asOf": "2026-11-16"}})
}

func (a *App) getConfig(w http.ResponseWriter, r *http.Request) {
	c, e := a.store.Config(r.Context())
	respond(w, c, e)
}
func (a *App) putConfig(w http.ResponseWriter, r *http.Request) {
	var c Config
	if e := json.NewDecoder(r.Body).Decode(&c); e != nil {
		writeError(w, 400, e)
		return
	}
	if e := a.store.SaveConfig(r.Context(), c); e != nil {
		writeError(w, 400, e)
		return
	}
	writeJSON(w, 200, c)
}
func (a *App) resetDemo(w http.ResponseWriter, r *http.Request) {
	c, e := a.store.Config(r.Context())
	if e != nil {
		writeError(w, 500, e)
		return
	}
	n := 1000
	if s := r.URL.Query().Get("count"); s != "" {
		if v, e := strconv.Atoi(s); e == nil && v > 0 && v <= 5000 {
			n = v
		}
	}
	data := GenerateDemoData(n, c)
	if e = a.store.ResetShipments(r.Context(), data); e != nil {
		writeError(w, 500, e)
		return
	}
	a.events.Publish(AnalysisProgress{Status: "idle", Total: n, Percent: 0})
	writeJSON(w, 200, map[string]any{"inserted": n})
}
func (a *App) downloadDemo(w http.ResponseWriter, r *http.Request) {
	c, e := a.store.Config(r.Context())
	if e != nil {
		writeError(w, 500, e)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="transportes_palma_demo.csv"`)
	if e := DemoCSV(w, GenerateDemoData(1000, c)); e != nil {
		log.Printf("CSV: %v", e)
	}
}
func (a *App) importCSV(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseMultipartForm(32 << 20); e != nil {
		writeError(w, 400, e)
		return
	}
	f, _, e := r.FormFile("file")
	if e != nil {
		writeError(w, 400, errors.New("se requiere el campo file"))
		return
	}
	defer f.Close()
	c, e := a.store.Config(r.Context())
	if e != nil {
		writeError(w, 500, e)
		return
	}
	rows, e := ParseCSV(f, c.DefaultOilPriceUSD)
	if e != nil {
		writeError(w, 400, e)
		return
	}
	n, e := a.store.InsertShipments(r.Context(), rows)
	if e != nil {
		writeError(w, 500, e)
		return
	}
	writeJSON(w, 200, map[string]any{"received": len(rows), "inserted": n, "duplicates": len(rows) - n})
}
func (a *App) shipments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	x, e := a.store.Shipments(r.Context(), limit)
	respond(w, x, e)
}
func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	d, e := a.store.Dashboard(r.Context())
	respond(w, d, e)
}
func (a *App) runAnalysis(w http.ResponseWriter, r *http.Request) {
	a.runMu.Lock()
	if a.running {
		a.runMu.Unlock()
		writeError(w, 409, errors.New("ya existe un análisis en ejecución"))
		return
	}
	a.running = true
	a.runMu.Unlock()
	go a.analyzeAll()
	writeJSON(w, 202, map[string]any{"status": "started"})
}
func (a *App) analyzeAll() {
	defer func() { a.runMu.Lock(); a.running = false; a.runMu.Unlock() }()
	ctx := context.Background()
	rows, err := a.store.PendingShipments(ctx)
	if err != nil {
		a.events.Publish(AnalysisProgress{Status: "error", Error: err.Error()})
		return
	}
	c, err := a.store.Config(ctx)
	if err != nil {
		a.events.Publish(AnalysisProgress{Status: "error", Error: err.Error()})
		return
	}
	started := time.Now().UTC()
	total := len(rows)
	a.events.Publish(AnalysisProgress{Status: "running", Total: total, StartedAt: started.Format(time.RFC3339)})
	for i, x := range rows {
		result, e := a.kev.Analyze(ctx, x, c)
		if e != nil {
			a.events.Publish(AnalysisProgress{Status: "error", Processed: i, Total: total, TripID: x.TripID, Error: e.Error()})
			return
		}
		if e = a.store.SaveAnalysis(ctx, result); e != nil {
			a.events.Publish(AnalysisProgress{Status: "error", Processed: i, Total: total, TripID: x.TripID, Error: e.Error()})
			return
		}
		a.events.Publish(AnalysisProgress{Status: "running", Processed: i + 1, Total: total, Percent: round(float64(i+1)/float64(max(total, 1))*100, 1), TripID: x.TripID, StartedAt: started.Format(time.RFC3339)})
	}
	a.events.Publish(AnalysisProgress{Status: "completed", Processed: total, Total: total, Percent: 100, StartedAt: started.Format(time.RFC3339), FinishedAt: time.Now().UTC().Format(time.RFC3339)})
}

func respond(w http.ResponseWriter, v any, e error) {
	if e != nil {
		writeError(w, 500, e)
		return
	}
	writeJSON(w, 200, v)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, e error) {
	writeJSON(w, status, map[string]any{"error": e.Error()})
}
func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ = fmt.Sprintf
