package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func GenerateDemoData(n int, c Config) []Shipment {
	if n < 1 {
		n = 1000
	}
	r := rand.New(rand.NewSource(42))
	routes := []struct {
		origin        string
		distanceKm    float64
		expectedHours float64
	}{
		{origin: "El Triunfo", distanceKm: 305, expectedHours: 6.5},
		{origin: "Santa Rosa", distanceKm: 245, expectedHours: 5.5},
		{origin: "Repsa I", distanceKm: 330, expectedHours: 7},
		{origin: "Repsa II", distanceKm: 345, expectedHours: 7.3},
	}
	drivers := []string{"Ana Torres", "Luis Rojas", "Carlos Díaz", "Marta León", "Diego Pérez", "Sofía Ruiz"}
	start := time.Now().UTC().AddDate(0, 0, -31).Truncate(time.Hour)
	stepMinutes := (30 * 24 * 60) / n
	if stepMinutes < 1 {
		stepMinutes = 1
	}
	out := make([]Shipment, 0, n)
	for i := 0; i < n; i++ {
		route := routes[i%len(routes)]
		netOut := 24000 + r.Float64()*9000
		tareOut := 10800 + r.Float64()*2600
		lossPct := math.Max(0, r.NormFloat64()*0.12+0.22)
		expected := route.expectedHours * (0.94 + r.Float64()*0.12)
		actual := expected * (0.9 + r.Float64()*0.25)
		tareIn := tareOut + r.NormFloat64()*45
		distance := route.distanceKm * (0.97 + r.Float64()*0.06)
		switch {
		case i%29 == 0:
			lossPct = 2.5 + r.Float64()*2.5
		case i%23 == 0:
			lossPct = -(0.5 + r.Float64()*1.2)
		case i%19 == 0:
			actual = expected * (1.7 + r.Float64())
		case i%17 == 0:
			tareIn = tareOut + 900 + r.Float64()*500
		case i%13 == 0:
			lossPct = 0.8 + r.Float64()*1.0
		}
		netIn := netOut * (1 - lossPct/100)
		departure := start.Add(time.Duration(i*stepMinutes+r.Intn(stepMinutes)) * time.Minute)
		arrival := departure.Add(time.Duration(actual*60) * time.Minute)
		out = append(out, Shipment{
			TripID: fmt.Sprintf("PALM-%04d", i+1), Vehicle: fmt.Sprintf("TRK-%03d", i%18+1),
			Driver: drivers[i%len(drivers)], Origin: route.origin,
			Destination: "Zolic", DepartureAt: departure, ArrivalAt: arrival,
			GrossOutKg: netOut + tareOut, TareOutKg: tareOut, NetOutKg: netOut,
			GrossInKg: netIn + tareIn, TareInKg: tareIn, NetInKg: netIn,
			DistanceKm: distance, ExpectedHours: expected, ActualHours: actual,
			OilPriceUSDPerKg: c.DefaultOilPriceUSD,
		})
	}
	return out
}

func ParseCSV(r io.Reader, defaultPrice float64) ([]Shipment, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("el CSV no contiene registros")
	}
	idx := map[string]int{}
	for i, h := range rows[0] {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	required := []string{"trip_id", "vehicle", "driver", "origin", "destination", "departure_at", "arrival_at",
		"distance_km", "expected_hours", "actual_hours"}
	for _, k := range required {
		if _, ok := idx[k]; !ok {
			return nil, fmt.Errorf("falta columna requerida %q", k)
		}
	}
	weightColumns := []string{"gross_out", "tare_out", "net_out", "gross_in", "tare_in", "net_in"}
	weightSuffix := ""
	weightMultiplier := 1.0
	for _, candidate := range []struct {
		suffix     string
		multiplier float64
	}{{"_t", 1000}, {"_kg", 1}} {
		complete := true
		for _, k := range weightColumns {
			if _, ok := idx[k+candidate.suffix]; !ok {
				complete = false
				break
			}
		}
		if complete {
			weightSuffix = candidate.suffix
			weightMultiplier = candidate.multiplier
			break
		}
	}
	if weightSuffix == "" {
		return nil, fmt.Errorf("faltan columnas de peso; use gross_out_t...net_in_t o las columnas compatibles en kg")
	}
	get := func(row []string, key string) string { return strings.TrimSpace(row[idx[key]]) }
	num := func(row []string, key string) (float64, error) { return strconv.ParseFloat(get(row, key), 64) }
	var out []Shipment
	for line, row := range rows[1:] {
		if len(row) < len(rows[0]) {
			return nil, fmt.Errorf("fila %d incompleta", line+2)
		}
		departure, err := time.Parse(time.RFC3339, get(row, "departure_at"))
		if err != nil {
			return nil, fmt.Errorf("fila %d departure_at: %w", line+2, err)
		}
		arrival, err := time.Parse(time.RFC3339, get(row, "arrival_at"))
		if err != nil {
			return nil, fmt.Errorf("fila %d arrival_at: %w", line+2, err)
		}
		x := Shipment{TripID: get(row, "trip_id"), Vehicle: get(row, "vehicle"), Driver: get(row, "driver"),
			Origin: get(row, "origin"), Destination: get(row, "destination"), DepartureAt: departure, ArrivalAt: arrival, OilPriceUSDPerKg: defaultPrice}
		fields := []struct {
			k string
			p *float64
		}{
			{"gross_out" + weightSuffix, &x.GrossOutKg}, {"tare_out" + weightSuffix, &x.TareOutKg}, {"net_out" + weightSuffix, &x.NetOutKg},
			{"gross_in" + weightSuffix, &x.GrossInKg}, {"tare_in" + weightSuffix, &x.TareInKg}, {"net_in" + weightSuffix, &x.NetInKg},
			{"distance_km", &x.DistanceKm}, {"expected_hours", &x.ExpectedHours}, {"actual_hours", &x.ActualHours},
		}
		for _, f := range fields {
			v, e := num(row, f.k)
			if e != nil {
				return nil, fmt.Errorf("fila %d %s: %w", line+2, f.k, e)
			}
			if strings.HasSuffix(f.k, weightSuffix) && f.k != "distance_km" && f.k != "expected_hours" && f.k != "actual_hours" {
				v *= weightMultiplier
			}
			*f.p = v
		}
		priceKey := "oil_price_usd_per_t"
		priceDivisor := 1000.0
		if _, ok := idx[priceKey]; !ok {
			priceKey = "oil_price_usd_per_kg"
			priceDivisor = 1
		}
		if p, ok := idx[priceKey]; ok && p < len(row) && strings.TrimSpace(row[p]) != "" {
			x.OilPriceUSDPerKg, err = strconv.ParseFloat(strings.TrimSpace(row[p]), 64)
			if err != nil {
				return nil, fmt.Errorf("fila %d %s: %w", line+2, priceKey, err)
			}
			x.OilPriceUSDPerKg /= priceDivisor
		}
		out = append(out, x)
	}
	return out, nil
}

func DemoCSV(w io.Writer, shipments []Shipment) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	headers := []string{"trip_id", "vehicle", "driver", "origin", "destination", "departure_at", "arrival_at", "gross_out_t",
		"tare_out_t", "net_out_t", "gross_in_t", "tare_in_t", "net_in_t", "distance_km", "expected_hours", "actual_hours", "oil_price_usd_per_t"}
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, x := range shipments {
		row := []string{x.TripID, x.Vehicle, x.Driver, x.Origin, x.Destination, x.DepartureAt.Format(time.RFC3339), x.ArrivalAt.Format(time.RFC3339),
			fmt.Sprintf("%.3f", x.GrossOutKg/1000), fmt.Sprintf("%.3f", x.TareOutKg/1000), fmt.Sprintf("%.3f", x.NetOutKg/1000),
			fmt.Sprintf("%.3f", x.GrossInKg/1000), fmt.Sprintf("%.3f", x.TareInKg/1000), fmt.Sprintf("%.3f", x.NetInKg/1000),
			fmt.Sprintf("%.2f", x.DistanceKm), fmt.Sprintf("%.2f", x.ExpectedHours), fmt.Sprintf("%.2f", x.ActualHours),
			fmt.Sprintf("%.2f", x.OilPriceUSDPerKg*1000)}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return cw.Error()
}
