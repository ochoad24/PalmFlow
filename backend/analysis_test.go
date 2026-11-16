package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDeterministicFlags(t *testing.T) {
	c := Config{LossTolerancePct: 0.5, MaxTransitOverPct: 35, MaxWeightGainPct: 0.2}
	x := Shipment{
		GrossOutKg: 35000, TareOutKg: 10000, NetOutKg: 25000,
		GrossInKg: 34000, TareInKg: 11000, NetInKg: 23000,
		ExpectedHours: 4, ActualHours: 7, DistanceKm: 180,
	}
	flags := deterministicFlags(x, c, 8, 75, 1000)
	got := strings.Join(flags, ",")
	for _, expected := range []string{"loss_above_tolerance", "transit_delay", "tare_shift_over_500kg"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in %q", expected, got)
		}
	}
}

func TestSeverityForCriticalLoss(t *testing.T) {
	c := Config{LossTolerancePct: 0.5}
	if got := severityFor(.2, nil, 2.1, c); got != "critical" {
		t.Fatalf("expected critical, got %s", got)
	}
}

func TestGenerateDemoDataIsDeterministic(t *testing.T) {
	c := Config{DefaultOilPriceUSD: 1.05}
	a, b := GenerateDemoData(5, c), GenerateDemoData(5, c)
	if a[0].NetOutKg != b[0].NetOutKg || a[4].ActualHours != b[4].ActualHours {
		t.Fatal("demo data should be reproducible")
	}
	wantOrigins := []string{"El Triunfo", "Santa Rosa", "Repsa I", "Repsa II"}
	for i, want := range wantOrigins {
		if a[i].Origin != want || a[i].Destination != "Zolic" {
			t.Fatalf("unexpected route %d: %s -> %s", i, a[i].Origin, a[i].Destination)
		}
	}
	fullDemo := GenerateDemoData(1000, c)
	if fullDemo[len(fullDemo)-1].ArrivalAt.After(time.Now().UTC()) {
		t.Fatalf("demo data must not include future arrivals: %s", fullDemo[len(fullDemo)-1].ArrivalAt)
	}
}

func TestDemoCSVUsesTonnesAndCanBeParsed(t *testing.T) {
	c := Config{DefaultOilPriceUSD: 1.05}
	var csvData strings.Builder
	if err := DemoCSV(&csvData, GenerateDemoData(2, c)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(csvData.String(), "gross_out_t") || !strings.Contains(csvData.String(), "oil_price_usd_per_t") {
		t.Fatalf("expected tonne columns in CSV header: %s", csvData.String())
	}
	rows, err := ParseCSV(strings.NewReader(csvData.String()), c.DefaultOilPriceUSD)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Destination != "Zolic" || rows[0].NetOutKg < 24000 {
		t.Fatalf("unexpected parsed rows: %+v", rows)
	}
}

func TestKevAnalysisAndCostProjection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header %q", got)
		}
		var request kevRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if len(request.Questions) != 5 {
			t.Fatalf("expected 5 questions, got %d", len(request.Questions))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "model":"kev-latest",
		  "answers":{
		    "weight_anomaly":{"type":"noul","noul":0.9},
		    "transit_anomaly":{"type":"noul","noul":0.1},
		    "data_quality":{"type":"choice","choice":"consistent","confidence":0.8,"probabilities":{"consistent":0.9}},
		    "fraud_risk":{"type":"score","score":1.5,"confidence":0.6,"probabilities":{"0":0.1,"1":0.4,"2":0.4,"3":0.1}},
		    "review_priority":{"type":"score","score":2.1,"confidence":0.7,"probabilities":{"0":0.0,"1":0.2,"2":0.5,"3":0.3}}
		  },
		  "usage":{"input_tokens":1000,"output_tokens":50},
		  "latency_ms":25
		}`))
	}))
	defer server.Close()

	client := KevClient{BaseURL: server.URL, APIKey: "test-key", HTTP: server.Client()}
	now := time.Now().UTC()
	shipment := Shipment{
		ID: 7, TripID: "T-7", Vehicle: "V-1", Driver: "D", Origin: "A", Destination: "B",
		DepartureAt: now, ArrivalAt: now.Add(4 * time.Hour),
		GrossOutKg: 35000, TareOutKg: 10000, NetOutKg: 25000,
		GrossInKg: 34800, TareInKg: 10000, NetInKg: 24800,
		DistanceKm: 150, ExpectedHours: 4, ActualHours: 4, OilPriceUSDPerKg: 1,
	}
	config := Config{LossTolerancePct: .5, MaxTransitOverPct: 35, MaxWeightGainPct: .2, GPTOutputTokens: 200, JevReferenceMS: 100, GPTReferenceMS: 8000}
	result, err := client.Analyze(context.Background(), shipment, config)
	if err != nil {
		t.Fatal(err)
	}
	if result.InputTokens != 1000 || result.KevLatencyMS != 25 {
		t.Fatalf("unexpected usage result: %+v", result)
	}
	if result.JevCostUSD <= 0 || result.GPTCostUSD <= result.JevCostUSD {
		t.Fatalf("unexpected projected costs: jev=%f gpt=%f", result.JevCostUSD, result.GPTCostUSD)
	}
}
