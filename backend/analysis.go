package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	jevInputUSDPerMillion  = 0.042
	gptInputUSDPerMillion  = 2.00
	gptOutputUSDPerMillion = 10.00
)

type KevClient struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

type kevRequest struct {
	Model     string         `json:"model"`
	State     any            `json:"state"`
	Questions map[string]any `json:"questions"`
}

type kevResponse struct {
	Model   string                     `json:"model"`
	Answers map[string]json.RawMessage `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	LatencyMS float64 `json:"latency_ms"`
}

type noulAnswer struct {
	Noul float64 `json:"noul"`
}
type choiceAnswer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}
type scoreAnswer struct {
	Score         float64            `json:"score"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

func (k *KevClient) Health(ctx context.Context) (map[string]any, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(k.BaseURL, "/")+"/v1/models", nil)
	if k.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+k.APIKey)
	}
	resp, err := k.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("KEV respondió HTTP %d", resp.StatusCode)
	}
	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	return body, err
}

func (k *KevClient) Analyze(ctx context.Context, x Shipment, c Config) (AnalysisResult, error) {
	lossKg := x.NetOutKg - x.NetInKg
	lossPct := 0.0
	if x.NetOutKg != 0 {
		lossPct = lossKg / x.NetOutKg * 100
	}
	transitOverPct := 0.0
	if x.ExpectedHours > 0 {
		transitOverPct = (x.ActualHours - x.ExpectedHours) / x.ExpectedHours * 100
	}
	tareDiff := x.TareInKg - x.TareOutKg
	state := map[string]any{
		"domain": "Palm crude oil bulk transport with independent weighbridges at origin and destination.",
		"policy": map[string]any{
			"maximum_expected_loss_percent":            c.LossTolerancePct,
			"maximum_expected_transit_overrun_percent": c.MaxTransitOverPct,
			"maximum_expected_weight_gain_percent":     c.MaxWeightGainPct,
			"rule":                                     "Calculations are supplied as facts. Judge the operational pattern and evidence. Do not invent missing facts.",
		},
		"trip": map[string]any{
			"id": x.TripID, "vehicle": x.Vehicle, "driver": x.Driver, "origin": x.Origin, "destination": x.Destination,
			"departure_utc": x.DepartureAt.Format(time.RFC3339), "arrival_utc": x.ArrivalAt.Format(time.RFC3339),
			"net_departure_kg": round(x.NetOutKg, 2), "net_arrival_kg": round(x.NetInKg, 2),
			"gross_departure_kg": round(x.GrossOutKg, 2), "tare_departure_kg": round(x.TareOutKg, 2),
			"gross_arrival_kg": round(x.GrossInKg, 2), "tare_arrival_kg": round(x.TareInKg, 2),
			"loss_kg": round(lossKg, 2), "loss_percent": round(lossPct, 4), "tare_difference_kg": round(tareDiff, 2),
			"distance_km": round(x.DistanceKm, 2), "expected_hours": round(x.ExpectedHours, 2),
			"actual_hours": round(x.ActualHours, 2), "transit_overrun_percent": round(transitOverPct, 2),
		},
	}
	questions := map[string]any{
		"weight_anomaly": map[string]any{"type": "noul", "instructions": "Is the measured net weight difference outside the supplied policy or physically suspicious?"},
		"data_quality": map[string]any{"type": "choice", "instructions": "Classify the record's measurement consistency.",
			"criteria": map[string]any{"consistent": "Arithmetic and measurements are plausible.", "weight_gain": "Arrival net weight is suspiciously above departure.",
				"tare_mismatch": "Tare change could materially distort the comparison.", "invalid_measurement": "Gross, tare, net, dates, or durations are internally invalid."}},
		"transit_anomaly": map[string]any{"type": "noul", "instructions": "Is travel time materially abnormal under the supplied policy?"},
		"fraud_risk": map[string]any{"type": "score", "instructions": "Rate evidence of manipulation, diversion, collusion, or weighbridge irregularity. High scores require evidence in this record.",
			"criteria": []string{"No material evidence", "Weak signal; monitor", "Strong signal; human investigation", "Critical evidence; block and investigate"}},
		"review_priority": map[string]any{"type": "score", "instructions": "Rate operational review priority considering weight, time, data quality, and financial exposure.",
			"criteria": []string{"Routine", "Review when convenient", "Review today", "Immediate escalation"}},
	}
	payload, _ := json.Marshal(kevRequest{Model: "kev-latest", State: state, Questions: questions})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(k.BaseURL, "/")+"/v1/systemone", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if k.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+k.APIKey)
	}
	started := time.Now()
	resp, err := k.HTTP.Do(req)
	if err != nil {
		return AnalysisResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var e any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return AnalysisResult{}, fmt.Errorf("KEV HTTP %d: %v", resp.StatusCode, e)
	}
	var kr kevResponse
	if err := json.NewDecoder(resp.Body).Decode(&kr); err != nil {
		return AnalysisResult{}, err
	}
	if kr.LatencyMS == 0 {
		kr.LatencyMS = float64(time.Since(started).Microseconds()) / 1000
	}
	rawAnswers, _ := json.Marshal(kr.Answers)

	var weight, transit noulAnswer
	var quality choiceAnswer
	var fraud, priority scoreAnswer
	_ = json.Unmarshal(kr.Answers["weight_anomaly"], &weight)
	_ = json.Unmarshal(kr.Answers["transit_anomaly"], &transit)
	_ = json.Unmarshal(kr.Answers["data_quality"], &quality)
	_ = json.Unmarshal(kr.Answers["fraud_risk"], &fraud)
	_ = json.Unmarshal(kr.Answers["review_priority"], &priority)

	flags := deterministicFlags(x, c, lossPct, transitOverPct, tareDiff)
	fraudProbability := clamp(fraud.Score/3, 0, 1)
	risk := clamp(0.30*weight.Noul+0.20*transit.Noul+0.25*fraudProbability+0.25*deterministicRisk(flags, lossPct, c), 0, 1)
	severity := severityFor(risk, flags, lossPct, c)
	review := scoreLabel(priority.Score, []string{"routine", "convenient", "today", "immediate"})
	if review == "routine" && severity != "normal" {
		review = "today"
	}
	impact := math.Max(lossKg, 0) * x.OilPriceUSDPerKg
	input := kr.Usage.InputTokens
	gptOutput := c.GPTOutputTokens
	return AnalysisResult{
		ShipmentID: x.ID, LossKg: round(lossKg, 2), LossPct: round(lossPct, 4), FinancialImpactUSD: round(impact, 2),
		DeterministicFlags: flags, KevAnswers: rawAnswers, RiskScore: round(risk, 4), FraudProbability: round(fraudProbability, 4),
		DataQuality: quality.Choice, ReviewPriority: review, Severity: severity, InputTokens: input, OutputTokens: kr.Usage.OutputTokens,
		KevLatencyMS: kr.LatencyMS, JevCostUSD: float64(input) * jevInputUSDPerMillion / 1_000_000,
		GPTCostUSD:     (float64(input)*gptInputUSDPerMillion + float64(gptOutput)*gptOutputUSDPerMillion) / 1_000_000,
		JevSimulatedMS: float64(c.JevReferenceMS), GPTSimulatedMS: float64(c.GPTReferenceMS), AnalyzedAt: time.Now().UTC(),
	}, nil
}

func deterministicFlags(x Shipment, c Config, lossPct, transitOverPct, tareDiff float64) []string {
	var f []string
	if x.GrossOutKg <= 0 || x.TareOutKg <= 0 || x.NetOutKg <= 0 || x.GrossInKg <= 0 || x.TareInKg <= 0 || x.NetInKg <= 0 {
		f = append(f, "non_positive_weight")
	}
	if math.Abs((x.GrossOutKg-x.TareOutKg)-x.NetOutKg) > 20 || math.Abs((x.GrossInKg-x.TareInKg)-x.NetInKg) > 20 {
		f = append(f, "weighbridge_arithmetic_mismatch")
	}
	if lossPct > c.LossTolerancePct {
		f = append(f, "loss_above_tolerance")
	}
	if lossPct < -c.MaxWeightGainPct {
		f = append(f, "unexpected_weight_gain")
	}
	if transitOverPct > c.MaxTransitOverPct {
		f = append(f, "transit_delay")
	}
	if x.ArrivalAt.Before(x.DepartureAt) || x.ActualHours <= 0 {
		f = append(f, "invalid_timeline")
	}
	if math.Abs(tareDiff) > 500 {
		f = append(f, "tare_shift_over_500kg")
	}
	if x.DistanceKm <= 0 || x.ExpectedHours <= 0 {
		f = append(f, "invalid_route_metrics")
	}
	return f
}

func deterministicRisk(flags []string, lossPct float64, c Config) float64 {
	r := float64(len(flags)) * 0.16
	if lossPct > c.LossTolerancePct {
		r += math.Min(lossPct/(c.LossTolerancePct*5), 0.55)
	}
	return clamp(r, 0, 1)
}

func severityFor(risk float64, flags []string, lossPct float64, c Config) string {
	for _, f := range flags {
		if f == "invalid_timeline" || f == "weighbridge_arithmetic_mismatch" || f == "non_positive_weight" {
			return "critical"
		}
	}
	if lossPct > c.LossTolerancePct*4 || risk >= .8 {
		return "critical"
	}
	if lossPct > c.LossTolerancePct*2 || risk >= .62 {
		return "high"
	}
	if len(flags) > 0 || risk >= .38 {
		return "warning"
	}
	return "normal"
}

func scoreLabel(score float64, labels []string) string {
	i := int(math.Round(score))
	if i < 0 {
		i = 0
	}
	if i >= len(labels) {
		i = len(labels) - 1
	}
	return labels[i]
}
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func round(v float64, n int) float64 { p := math.Pow10(n); return math.Round(v*p) / p }
