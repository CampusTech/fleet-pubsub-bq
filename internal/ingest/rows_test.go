package ingest

import (
	"encoding/json"
	"testing"
	"time"
)

var ts = time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

func TestBuildResultRows_Snapshot(t *testing.T) {
	payload := []byte(`{
		"name": "os_version",
		"hostIdentifier": "host-abc",
		"calendarTime": "Mon May  4 12:00:00 2026 UTC",
		"unixTime": 1746360000,
		"action": "snapshot",
		"epoch": 0,
		"counter": 1,
		"decorations": {"host_uuid": "uuid-123"},
		"snapshot": [
			{"name": "Ubuntu", "version": "22.04"},
			{"name": "Ubuntu", "version": "22.04"}
		]
	}`)

	rows, err := BuildResultRows(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Action != "snapshot" {
			t.Errorf("want action=snapshot, got %q", r.Action)
		}
		if r.QueryName != "os_version" {
			t.Errorf("want query_name=os_version, got %q", r.QueryName)
		}
		if r.HostUUID != "uuid-123" {
			t.Errorf("want host_uuid=uuid-123, got %q", r.HostUUID)
		}
		if r.InsertedAt != ts {
			t.Errorf("want inserted_at=%v, got %v", ts, r.InsertedAt)
		}
		// Row should be valid JSON
		var m map[string]string
		if err := json.Unmarshal([]byte(r.Row), &m); err != nil {
			t.Errorf("row is not valid JSON: %v", err)
		}
	}
}

func TestBuildResultRows_DiffResults(t *testing.T) {
	payload := []byte(`{
		"name": "listening_ports",
		"hostIdentifier": "host-def",
		"calendarTime": "Mon May  4 12:00:00 2026 UTC",
		"unixTime": 1746360000,
		"epoch": 0,
		"counter": 2,
		"diffResults": {
			"added":   [{"port": "80"}],
			"removed": [{"port": "8080"}, {"port": "9090"}]
		}
	}`)

	rows, err := BuildResultRows(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("want 3 rows (1 added + 2 removed), got %d", len(rows))
	}
	if rows[0].Action != "added" {
		t.Errorf("first row action want added, got %q", rows[0].Action)
	}
	if rows[1].Action != "removed" || rows[2].Action != "removed" {
		t.Errorf("last two rows want removed, got %q and %q", rows[1].Action, rows[2].Action)
	}
}

func TestBuildResultRows_Differential(t *testing.T) {
	payload := []byte(`{
		"name": "processes",
		"hostIdentifier": "host-ghi",
		"calendarTime": "Mon May  4 12:00:00 2026 UTC",
		"unixTime": 1746360000,
		"action": "added",
		"epoch": 0,
		"counter": 1,
		"columns": {"pid": "1234", "name": "nginx"}
	}`)

	rows, err := BuildResultRows(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].Action != "added" {
		t.Errorf("want action=added, got %q", rows[0].Action)
	}
}

func TestBuildResultRows_EmptySnapshot(t *testing.T) {
	payload := []byte(`{
		"name": "empty_query",
		"hostIdentifier": "host-xyz",
		"unixTime": 1746360000,
		"action": "snapshot",
		"snapshot": []
	}`)

	rows, err := BuildResultRows(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("want 0 rows for empty snapshot, got %d", len(rows))
	}
}

func TestBuildStatusRow(t *testing.T) {
	payload := []byte(`{
		"severity": "1",
		"filename": "scheduler.cpp",
		"line": 42,
		"message": "query timeout",
		"version": "5.12.1",
		"decorations": {"host_uuid": "uuid-abc"}
	}`)

	row, err := BuildStatusRow(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.Severity != 1 {
		t.Errorf("want severity=1, got %d", row.Severity)
	}
	if row.Filename != "scheduler.cpp" {
		t.Errorf("want filename=scheduler.cpp, got %q", row.Filename)
	}
	if row.HostUUID != "uuid-abc" {
		t.Errorf("want host_uuid=uuid-abc, got %q", row.HostUUID)
	}
	if row.InsertedAt != ts {
		t.Errorf("want inserted_at=%v, got %v", ts, row.InsertedAt)
	}
}

func TestBuildAuditRow(t *testing.T) {
	payload := []byte(`{
		"id": 99,
		"uuid": "aud-uuid-1",
		"created_at": "2026-05-04T12:00:00Z",
		"type": "created_user",
		"actor_id": 7,
		"actor_full_name": "Robbie T",
		"actor_email": "robbie@campus.edu",
		"actor_api_only": false,
		"fleet_initiated": false,
		"details": {"target_email": "new@campus.edu"}
	}`)

	row, err := BuildAuditRow(payload, ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.Type != "created_user" {
		t.Errorf("want type=created_user, got %q", row.Type)
	}
	if row.ActorEmail != "robbie@campus.edu" {
		t.Errorf("want actor_email=robbie@campus.edu, got %q", row.ActorEmail)
	}
	if row.Details == "" {
		t.Error("want details non-empty")
	}
	if row.CreatedAt.IsZero() {
		t.Error("want created_at parsed, got zero")
	}
}

func TestExtractHostUUID(t *testing.T) {
	cases := []struct {
		in   []byte
		want string
	}{
		{[]byte(`{"host_uuid":"abc-123","hostname":"myhost"}`), "abc-123"},
		{[]byte(`{"hostname":"myhost"}`), ""},
		{nil, ""},
		{[]byte(`null`), ""},
	}
	for _, c := range cases {
		if got := extractHostUUID(c.in); got != c.want {
			t.Errorf("extractHostUUID(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}
