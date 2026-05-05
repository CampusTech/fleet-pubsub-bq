package ingest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ResultRow is a single BQ row for the result_logs table.
type ResultRow struct {
	InsertedAt     time.Time `bigquery:"inserted_at"`
	QueryName      string    `bigquery:"query_name"`
	QueryID        *int64    `bigquery:"query_id"`
	HostIdentifier string    `bigquery:"host_identifier"`
	CalendarTime   string    `bigquery:"calendar_time"`
	UnixTime       time.Time `bigquery:"unix_time"`
	Action         string    `bigquery:"action"`
	Epoch          int64     `bigquery:"epoch"`
	Counter        int64     `bigquery:"counter"`
	HostUUID       string    `bigquery:"host_uuid"`
	Decorations    string    `bigquery:"decorations"`
	Row            string    `bigquery:"row"`
}

// StatusRow is a single BQ row for the status_logs table.
type StatusRow struct {
	InsertedAt  time.Time `bigquery:"inserted_at"`
	Severity    int64     `bigquery:"severity"`
	Filename    string    `bigquery:"filename"`
	Line        int64     `bigquery:"line"`
	Message     string    `bigquery:"message"`
	Version     string    `bigquery:"version"`
	HostUUID    string    `bigquery:"host_uuid"`
	Decorations string    `bigquery:"decorations"`
}

// AuditRow is a single BQ row for the audit_logs table.
type AuditRow struct {
	InsertedAt     time.Time `bigquery:"inserted_at"`
	ID             *int64    `bigquery:"id"`
	UUID           string    `bigquery:"uuid"`
	CreatedAt      time.Time `bigquery:"created_at"`
	Type           string    `bigquery:"type"`
	ActorID        *int64    `bigquery:"actor_id"`
	ActorFullName  string    `bigquery:"actor_full_name"`
	ActorEmail     string    `bigquery:"actor_email"`
	ActorAPIOnly   *bool     `bigquery:"actor_api_only"`
	FleetInitiated *bool     `bigquery:"fleet_initiated"`
	Details        string    `bigquery:"details"`
}

// BuildResultRows parses a raw result message payload and returns BQ rows.
func BuildResultRows(data []byte, insertedAt time.Time) ([]*ResultRow, error) {
	var env resultEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	decorationsStr := rawToString(env.Decorations)
	hostUUID := extractHostUUID(env.Decorations)
	unixTime := time.Unix(env.UnixTime, 0).UTC()

	base := func(action string, rowJSON json.RawMessage) *ResultRow {
		return &ResultRow{
			InsertedAt:     insertedAt,
			QueryName:      env.Name,
			QueryID:        env.QueryID,
			HostIdentifier: env.HostIdentifier,
			CalendarTime:   env.CalendarTime,
			UnixTime:       unixTime,
			Action:         action,
			Epoch:          env.Epoch,
			Counter:        env.Counter,
			HostUUID:       hostUUID,
			Decorations:    decorationsStr,
			Row:            string(rowJSON),
		}
	}

	var rows []*ResultRow

	switch {
	case env.Action == "snapshot" && len(env.Snapshot) > 0:
		for _, r := range env.Snapshot {
			rows = append(rows, base("snapshot", r))
		}

	case env.DiffResults != nil:
		for _, r := range env.DiffResults.Added {
			rows = append(rows, base("added", r))
		}
		for _, r := range env.DiffResults.Removed {
			rows = append(rows, base("removed", r))
		}

	case env.Columns != nil:
		rows = append(rows, base(env.Action, env.Columns))

	default:
		// no rows to insert (empty snapshot, etc.)
	}

	return rows, nil
}

// BuildStatusRow parses a raw status message payload and returns a BQ row.
func BuildStatusRow(data []byte, insertedAt time.Time) (*StatusRow, error) {
	var env statusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal status: %w", err)
	}
	sev, _ := strconv.ParseInt(env.Severity, 10, 64)
	return &StatusRow{
		InsertedAt:  insertedAt,
		Severity:    sev,
		Filename:    env.Filename,
		Line:        env.Line,
		Message:     env.Message,
		Version:     env.Version,
		HostUUID:    extractHostUUID(env.Decorations),
		Decorations: rawToString(env.Decorations),
	}, nil
}

// BuildAuditRow parses a raw audit message payload and returns a BQ row.
func BuildAuditRow(data []byte, insertedAt time.Time) (*AuditRow, error) {
	var env auditEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal audit: %w", err)
	}

	createdAt := time.Time{}
	if env.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, env.CreatedAt); err == nil {
			createdAt = t.UTC()
		}
	}

	return &AuditRow{
		InsertedAt:     insertedAt,
		ID:             env.ID,
		UUID:           env.UUID,
		CreatedAt:      createdAt,
		Type:           env.Type,
		ActorID:        env.ActorID,
		ActorFullName:  env.ActorFullName,
		ActorEmail:     env.ActorEmail,
		ActorAPIOnly:   env.ActorAPIOnly,
		FleetInitiated: env.FleetInitiated,
		Details:        rawToString(env.Details),
	}, nil
}
