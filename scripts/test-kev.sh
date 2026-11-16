#!/usr/bin/env bash
set -euo pipefail

curl -fsS http://127.0.0.1:8009/v1/systemone \
  -H "Authorization: Bearer ${KEV_API_KEY:-local-demo-key}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "kev-latest",
    "state": {
      "domain": "Palm crude oil transport",
      "net_departure_kg": 30000,
      "net_arrival_kg": 29100,
      "loss_percent": 3.0,
      "maximum_expected_loss_percent": 0.5
    },
    "questions": {
      "weight_anomaly": {
        "type": "noul",
        "instructions": "Is the measured weight difference outside policy or physically suspicious?"
      },
      "priority": {
        "type": "score",
        "instructions": "How urgently should a human review this trip?",
        "criteria": ["Routine", "Review later", "Review today", "Immediate escalation"]
      }
    }
  }'
echo
