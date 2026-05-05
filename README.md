# fleet-pubsub-bq

Cloud Run service that receives Fleet osquery log messages from GCP PubSub push subscriptions and writes them to BigQuery.

## Environment variables

| Variable | Description |
|---|---|
| `BQ_PROJECT_ID` | GCP project hosting the BigQuery dataset |
| `BQ_DATASET_ID` | BigQuery dataset ID (e.g. `fleet_logs`) |
| `RESULT_SUBSCRIPTION` | PubSub subscription ID for osquery result logs |
| `STATUS_SUBSCRIPTION` | PubSub subscription ID for osquery status logs |
| `AUDIT_SUBSCRIPTION` | PubSub subscription ID for Fleet audit logs |
| `PORT` | HTTP port (default `8080`) |

## BigQuery tables

- `result_logs` — One row per osquery result row. Snapshot arrays are exploded; diffResults are split into `added`/`removed` rows.
- `status_logs` — Osquery agent status/error logs.
- `audit_logs` — Fleet user/automation audit activity.

## Building

```sh
docker build -t fleet-pubsub-bq .
```

## Testing

```sh
go test ./...
```

## Authorship

This code was written by Claude (Anthropic) and reviewed by a human before submission.
