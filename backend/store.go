package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ db *pgxpool.Pool }

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	s := &Store{db: db}
	return s, s.migrate(ctx)
}

func (s *Store) Close() { s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS app_config (
  id smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  loss_tolerance_pct double precision NOT NULL DEFAULT 0.5,
  max_transit_over_pct double precision NOT NULL DEFAULT 35,
  max_weight_gain_pct double precision NOT NULL DEFAULT 0.2,
  default_oil_price_usd double precision NOT NULL DEFAULT 1.05,
  gpt_output_tokens integer NOT NULL DEFAULT 220,
  jev_reference_ms integer NOT NULL DEFAULT 100,
  gpt_reference_ms integer NOT NULL DEFAULT 8311
);
INSERT INTO app_config(id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS shipments (
  id bigserial PRIMARY KEY,
  trip_id text NOT NULL UNIQUE,
  vehicle text NOT NULL,
  driver text NOT NULL,
  origin text NOT NULL,
  destination text NOT NULL,
  departure_at timestamptz NOT NULL,
  arrival_at timestamptz NOT NULL,
  gross_out_kg double precision NOT NULL,
  tare_out_kg double precision NOT NULL,
  net_out_kg double precision NOT NULL,
  gross_in_kg double precision NOT NULL,
  tare_in_kg double precision NOT NULL,
  net_in_kg double precision NOT NULL,
  distance_km double precision NOT NULL,
  expected_hours double precision NOT NULL,
  actual_hours double precision NOT NULL,
  oil_price_usd_per_kg double precision NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS analyses (
  shipment_id bigint PRIMARY KEY REFERENCES shipments(id) ON DELETE CASCADE,
  loss_kg double precision NOT NULL,
  loss_pct double precision NOT NULL,
  financial_impact_usd double precision NOT NULL,
  deterministic_flags jsonb NOT NULL,
  kev_answers jsonb NOT NULL,
  risk_score double precision NOT NULL,
  fraud_probability double precision NOT NULL,
  data_quality text NOT NULL,
  review_priority text NOT NULL,
  severity text NOT NULL,
  input_tokens integer NOT NULL,
  output_tokens integer NOT NULL,
  kev_latency_ms double precision NOT NULL,
  jev_cost_usd double precision NOT NULL,
  gpt_cost_usd double precision NOT NULL,
  jev_simulated_ms double precision NOT NULL,
  gpt_simulated_ms double precision NOT NULL,
  analyzed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS analyses_severity_idx ON analyses(severity);
`)
	return err
}

func (s *Store) Config(ctx context.Context) (Config, error) {
	var c Config
	err := s.db.QueryRow(ctx, `SELECT loss_tolerance_pct,max_transit_over_pct,max_weight_gain_pct,
default_oil_price_usd,gpt_output_tokens,jev_reference_ms,gpt_reference_ms FROM app_config WHERE id=1`).
		Scan(&c.LossTolerancePct, &c.MaxTransitOverPct, &c.MaxWeightGainPct, &c.DefaultOilPriceUSD,
			&c.GPTOutputTokens, &c.JevReferenceMS, &c.GPTReferenceMS)
	return c, err
}

func (s *Store) SaveConfig(ctx context.Context, c Config) error {
	if c.LossTolerancePct < 0 || c.LossTolerancePct > 20 || c.MaxTransitOverPct < 0 ||
		c.MaxWeightGainPct < 0 || c.DefaultOilPriceUSD <= 0 || c.GPTOutputTokens < 1 {
		return fmt.Errorf("configuración fuera de rango")
	}
	_, err := s.db.Exec(ctx, `UPDATE app_config SET loss_tolerance_pct=$1,max_transit_over_pct=$2,
max_weight_gain_pct=$3,default_oil_price_usd=$4,gpt_output_tokens=$5,jev_reference_ms=$6,gpt_reference_ms=$7 WHERE id=1`,
		c.LossTolerancePct, c.MaxTransitOverPct, c.MaxWeightGainPct, c.DefaultOilPriceUSD,
		c.GPTOutputTokens, c.JevReferenceMS, c.GPTReferenceMS)
	return err
}

func (s *Store) ResetShipments(ctx context.Context, shipments []Shipment) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `TRUNCATE analyses, shipments RESTART IDENTITY`); err != nil {
		return err
	}
	for _, x := range shipments {
		_, err = tx.Exec(ctx, `INSERT INTO shipments(trip_id,vehicle,driver,origin,destination,departure_at,arrival_at,
gross_out_kg,tare_out_kg,net_out_kg,gross_in_kg,tare_in_kg,net_in_kg,distance_km,expected_hours,actual_hours,oil_price_usd_per_kg)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
			x.TripID, x.Vehicle, x.Driver, x.Origin, x.Destination, x.DepartureAt, x.ArrivalAt,
			x.GrossOutKg, x.TareOutKg, x.NetOutKg, x.GrossInKg, x.TareInKg, x.NetInKg,
			x.DistanceKm, x.ExpectedHours, x.ActualHours, x.OilPriceUSDPerKg)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) InsertShipments(ctx context.Context, shipments []Shipment) (int, error) {
	count := 0
	for _, x := range shipments {
		tag, err := s.db.Exec(ctx, `INSERT INTO shipments(trip_id,vehicle,driver,origin,destination,departure_at,arrival_at,
gross_out_kg,tare_out_kg,net_out_kg,gross_in_kg,tare_in_kg,net_in_kg,distance_km,expected_hours,actual_hours,oil_price_usd_per_kg)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) ON CONFLICT(trip_id) DO NOTHING`,
			x.TripID, x.Vehicle, x.Driver, x.Origin, x.Destination, x.DepartureAt, x.ArrivalAt,
			x.GrossOutKg, x.TareOutKg, x.NetOutKg, x.GrossInKg, x.TareInKg, x.NetInKg,
			x.DistanceKm, x.ExpectedHours, x.ActualHours, x.OilPriceUSDPerKg)
		if err != nil {
			return count, err
		}
		count += int(tag.RowsAffected())
	}
	return count, nil
}

func (s *Store) PendingShipments(ctx context.Context) ([]Shipment, error) {
	rows, err := s.db.Query(ctx, `SELECT s.id,s.trip_id,s.vehicle,s.driver,s.origin,s.destination,s.departure_at,s.arrival_at,
s.gross_out_kg,s.tare_out_kg,s.net_out_kg,s.gross_in_kg,s.tare_in_kg,s.net_in_kg,s.distance_km,
s.expected_hours,s.actual_hours,s.oil_price_usd_per_kg,s.created_at
FROM shipments s LEFT JOIN analyses a ON a.shipment_id=s.id WHERE a.shipment_id IS NULL ORDER BY s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Shipment
	for rows.Next() {
		var x Shipment
		if err := rows.Scan(&x.ID, &x.TripID, &x.Vehicle, &x.Driver, &x.Origin, &x.Destination, &x.DepartureAt,
			&x.ArrivalAt, &x.GrossOutKg, &x.TareOutKg, &x.NetOutKg, &x.GrossInKg, &x.TareInKg, &x.NetInKg,
			&x.DistanceKm, &x.ExpectedHours, &x.ActualHours, &x.OilPriceUSDPerKg, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) SaveAnalysis(ctx context.Context, a AnalysisResult) error {
	flags, _ := json.Marshal(a.DeterministicFlags)
	_, err := s.db.Exec(ctx, `INSERT INTO analyses(shipment_id,loss_kg,loss_pct,financial_impact_usd,
deterministic_flags,kev_answers,risk_score,fraud_probability,data_quality,review_priority,severity,
input_tokens,output_tokens,kev_latency_ms,jev_cost_usd,gpt_cost_usd,jev_simulated_ms,gpt_simulated_ms,analyzed_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
ON CONFLICT(shipment_id) DO UPDATE SET loss_kg=EXCLUDED.loss_kg,loss_pct=EXCLUDED.loss_pct,
financial_impact_usd=EXCLUDED.financial_impact_usd,deterministic_flags=EXCLUDED.deterministic_flags,
kev_answers=EXCLUDED.kev_answers,risk_score=EXCLUDED.risk_score,fraud_probability=EXCLUDED.fraud_probability,
data_quality=EXCLUDED.data_quality,review_priority=EXCLUDED.review_priority,severity=EXCLUDED.severity,
input_tokens=EXCLUDED.input_tokens,output_tokens=EXCLUDED.output_tokens,kev_latency_ms=EXCLUDED.kev_latency_ms,
jev_cost_usd=EXCLUDED.jev_cost_usd,gpt_cost_usd=EXCLUDED.gpt_cost_usd,
jev_simulated_ms=EXCLUDED.jev_simulated_ms,gpt_simulated_ms=EXCLUDED.gpt_simulated_ms,analyzed_at=EXCLUDED.analyzed_at`,
		a.ShipmentID, a.LossKg, a.LossPct, a.FinancialImpactUSD, flags, a.KevAnswers, a.RiskScore,
		a.FraudProbability, a.DataQuality, a.ReviewPriority, a.Severity, a.InputTokens, a.OutputTokens,
		a.KevLatencyMS, a.JevCostUSD, a.GPTCostUSD, a.JevSimulatedMS, a.GPTSimulatedMS, a.AnalyzedAt)
	return err
}

func (s *Store) Shipments(ctx context.Context, limit int) ([]Shipment, error) {
	if limit <= 0 || limit > 1000 {
		limit = 250
	}
	rows, err := s.db.Query(ctx, `SELECT s.id,s.trip_id,s.vehicle,s.driver,s.origin,s.destination,s.departure_at,s.arrival_at,
s.gross_out_kg,s.tare_out_kg,s.net_out_kg,s.gross_in_kg,s.tare_in_kg,s.net_in_kg,s.distance_km,
s.expected_hours,s.actual_hours,s.oil_price_usd_per_kg,s.created_at,
a.loss_kg,a.loss_pct,a.financial_impact_usd,a.deterministic_flags,a.kev_answers,a.risk_score,
a.fraud_probability,a.data_quality,a.review_priority,a.severity,a.input_tokens,a.output_tokens,
a.kev_latency_ms,a.jev_cost_usd,a.gpt_cost_usd,a.jev_simulated_ms,a.gpt_simulated_ms,a.analyzed_at
FROM shipments s LEFT JOIN analyses a ON a.shipment_id=s.id ORDER BY s.id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Shipment
	for rows.Next() {
		var x Shipment
		var a AnalysisResult
		var flags, answers []byte
		var lossKg, lossPct, impact, risk, fraud, kevMS, jevCost, gptCost, jevMS, gptMS *float64
		var quality, priority, severity *string
		var inTok, outTok *int
		var analyzed *time.Time
		if err := rows.Scan(&x.ID, &x.TripID, &x.Vehicle, &x.Driver, &x.Origin, &x.Destination, &x.DepartureAt,
			&x.ArrivalAt, &x.GrossOutKg, &x.TareOutKg, &x.NetOutKg, &x.GrossInKg, &x.TareInKg, &x.NetInKg,
			&x.DistanceKm, &x.ExpectedHours, &x.ActualHours, &x.OilPriceUSDPerKg, &x.CreatedAt,
			&lossKg, &lossPct, &impact, &flags, &answers, &risk, &fraud, &quality, &priority, &severity, &inTok, &outTok,
			&kevMS, &jevCost, &gptCost, &jevMS, &gptMS, &analyzed); err != nil {
			return nil, err
		}
		if analyzed != nil {
			a = AnalysisResult{ShipmentID: x.ID, LossKg: *lossKg, LossPct: *lossPct, FinancialImpactUSD: *impact,
				KevAnswers: answers, RiskScore: *risk, FraudProbability: *fraud, DataQuality: *quality,
				ReviewPriority: *priority, Severity: *severity, InputTokens: *inTok, OutputTokens: *outTok,
				KevLatencyMS: *kevMS, JevCostUSD: *jevCost, GPTCostUSD: *gptCost, JevSimulatedMS: *jevMS,
				GPTSimulatedMS: *gptMS, AnalyzedAt: *analyzed}
			_ = json.Unmarshal(flags, &a.DeterministicFlags)
			x.Analysis = &a
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) Dashboard(ctx context.Context) (Dashboard, error) {
	var d Dashboard
	d.SeverityCounts = map[string]int64{"normal": 0, "warning": 0, "high": 0, "critical": 0}
	err := s.db.QueryRow(ctx, `SELECT COUNT(*),COUNT(a.shipment_id),
COUNT(*) FILTER (WHERE a.severity <> 'normal'),COUNT(*) FILTER (WHERE a.severity='critical'),
COALESCE(SUM(GREATEST(a.loss_kg,0)),0),COALESCE(SUM(a.financial_impact_usd),0),
COALESCE(AVG(a.loss_pct),0),COALESCE(AVG(a.kev_latency_ms),0),
COALESCE(SUM(a.jev_cost_usd),0),COALESCE(SUM(a.gpt_cost_usd),0),
COALESCE(SUM(a.jev_simulated_ms),0),COALESCE(SUM(a.gpt_simulated_ms),0)
FROM shipments s LEFT JOIN analyses a ON a.shipment_id=s.id`).
		Scan(&d.TotalRecords, &d.AnalyzedRecords, &d.AnomalyRecords, &d.CriticalRecords, &d.TotalLossKg,
			&d.FinancialImpactUSD, &d.AvgLossPct, &d.AvgKevLatencyMS, &d.JevTotalCostUSD, &d.GPTTotalCostUSD,
			&d.JevTotalTimeMS, &d.GPTTotalTimeMS)
	if err != nil {
		return d, err
	}
	rows, err := s.db.Query(ctx, `SELECT severity,COUNT(*) FROM analyses GROUP BY severity`)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			return d, err
		}
		d.SeverityCounts[k] = v
	}
	d.Config, err = s.Config(ctx)
	return d, err
}
