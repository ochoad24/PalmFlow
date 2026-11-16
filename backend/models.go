package main

import (
	"encoding/json"
	"time"
)

type Config struct {
	LossTolerancePct   float64 `json:"lossTolerancePct"`
	MaxTransitOverPct  float64 `json:"maxTransitOverPct"`
	MaxWeightGainPct   float64 `json:"maxWeightGainPct"`
	DefaultOilPriceUSD float64 `json:"defaultOilPriceUsd"`
	GPTOutputTokens    int     `json:"gptOutputTokens"`
	JevReferenceMS     int     `json:"jevReferenceMs"`
	GPTReferenceMS     int     `json:"gptReferenceMs"`
}

type Shipment struct {
	ID               int64           `json:"id"`
	TripID           string          `json:"tripId"`
	Vehicle          string          `json:"vehicle"`
	Driver           string          `json:"driver"`
	Origin           string          `json:"origin"`
	Destination      string          `json:"destination"`
	DepartureAt      time.Time       `json:"departureAt"`
	ArrivalAt        time.Time       `json:"arrivalAt"`
	GrossOutKg       float64         `json:"grossOutKg"`
	TareOutKg        float64         `json:"tareOutKg"`
	NetOutKg         float64         `json:"netOutKg"`
	GrossInKg        float64         `json:"grossInKg"`
	TareInKg         float64         `json:"tareInKg"`
	NetInKg          float64         `json:"netInKg"`
	DistanceKm       float64         `json:"distanceKm"`
	ExpectedHours    float64         `json:"expectedHours"`
	ActualHours      float64         `json:"actualHours"`
	OilPriceUSDPerKg float64         `json:"oilPriceUsdPerKg"`
	CreatedAt        time.Time       `json:"createdAt"`
	Analysis         *AnalysisResult `json:"analysis,omitempty"`
}

type AnalysisResult struct {
	ShipmentID         int64           `json:"shipmentId"`
	LossKg             float64         `json:"lossKg"`
	LossPct            float64         `json:"lossPct"`
	FinancialImpactUSD float64         `json:"financialImpactUsd"`
	DeterministicFlags []string        `json:"deterministicFlags"`
	KevAnswers         json.RawMessage `json:"kevAnswers"`
	RiskScore          float64         `json:"riskScore"`
	FraudProbability   float64         `json:"fraudProbability"`
	DataQuality        string          `json:"dataQuality"`
	ReviewPriority     string          `json:"reviewPriority"`
	Severity           string          `json:"severity"`
	InputTokens        int             `json:"inputTokens"`
	OutputTokens       int             `json:"outputTokens"`
	KevLatencyMS       float64         `json:"kevLatencyMs"`
	JevCostUSD         float64         `json:"jevCostUsd"`
	GPTCostUSD         float64         `json:"gptCostUsd"`
	JevSimulatedMS     float64         `json:"jevSimulatedMs"`
	GPTSimulatedMS     float64         `json:"gptSimulatedMs"`
	AnalyzedAt         time.Time       `json:"analyzedAt"`
}

type Dashboard struct {
	TotalRecords       int64            `json:"totalRecords"`
	AnalyzedRecords    int64            `json:"analyzedRecords"`
	AnomalyRecords     int64            `json:"anomalyRecords"`
	CriticalRecords    int64            `json:"criticalRecords"`
	TotalLossKg        float64          `json:"totalLossKg"`
	FinancialImpactUSD float64          `json:"financialImpactUsd"`
	AvgLossPct         float64          `json:"avgLossPct"`
	AvgKevLatencyMS    float64          `json:"avgKevLatencyMs"`
	KevTotalCostUSD    float64          `json:"kevTotalCostUsd"`
	JevTotalCostUSD    float64          `json:"jevTotalCostUsd"`
	GPTTotalCostUSD    float64          `json:"gptTotalCostUsd"`
	JevTotalTimeMS     float64          `json:"jevTotalTimeMs"`
	GPTTotalTimeMS     float64          `json:"gptTotalTimeMs"`
	SeverityCounts     map[string]int64 `json:"severityCounts"`
	Config             Config           `json:"config"`
}

type AnalysisProgress struct {
	Status     string  `json:"status"`
	Processed  int     `json:"processed"`
	Total      int     `json:"total"`
	Percent    float64 `json:"percent"`
	TripID     string  `json:"tripId,omitempty"`
	Error      string  `json:"error,omitempty"`
	StartedAt  string  `json:"startedAt,omitempty"`
	FinishedAt string  `json:"finishedAt,omitempty"`
}
