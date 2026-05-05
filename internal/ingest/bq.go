package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sirupsen/logrus"
)

// Handler holds the BigQuery client and table routing config.
type Handler struct {
	client          *bigquery.Client
	dataset         string
	resultSubSuffix string
	statusSubSuffix string
	auditSubSuffix  string
	log             *logrus.Logger
}

// NewHandler creates a Handler.
func NewHandler(ctx context.Context, bqProjectID, datasetID, resultSub, statusSub, auditSub string, log *logrus.Logger) (*Handler, error) {
	client, err := bigquery.NewClient(ctx, bqProjectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}
	return &Handler{
		client:          client,
		dataset:         datasetID,
		resultSubSuffix: resultSub,
		statusSubSuffix: statusSub,
		auditSubSuffix:  auditSub,
		log:             log,
	}, nil
}

func (h *Handler) Close() { h.client.Close() }

func (h *Handler) inserter(table string) *bigquery.Inserter {
	return h.client.Dataset(h.dataset).Table(table).Inserter()
}

func (h *Handler) InsertResult(ctx context.Context, msg PubSubMessage, insertedAt time.Time) error {
	rows, err := BuildResultRows(msg.Message.Data, insertedAt)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		h.log.Warn("result message produced no rows; skipping")
		return nil
	}
	items := make([]interface{}, len(rows))
	for i, r := range rows {
		items[i] = r
	}
	if err := h.inserter("result_logs").Put(ctx, items); err != nil {
		return fmt.Errorf("BQ insert result_logs: %w", err)
	}
	h.log.WithField("rows", len(rows)).Debug("inserted result rows")
	return nil
}

func (h *Handler) InsertStatus(ctx context.Context, msg PubSubMessage, insertedAt time.Time) error {
	row, err := BuildStatusRow(msg.Message.Data, insertedAt)
	if err != nil {
		return err
	}
	if err := h.inserter("status_logs").Put(ctx, row); err != nil {
		return fmt.Errorf("BQ insert status_logs: %w", err)
	}
	return nil
}

func (h *Handler) InsertAudit(ctx context.Context, msg PubSubMessage, insertedAt time.Time) error {
	row, err := BuildAuditRow(msg.Message.Data, insertedAt)
	if err != nil {
		return err
	}
	if err := h.inserter("audit_logs").Put(ctx, row); err != nil {
		return fmt.Errorf("BQ insert audit_logs: %w", err)
	}
	return nil
}

func rawToString(b []byte) string {
	if len(b) == 0 || string(b) == "null" {
		return ""
	}
	return string(b)
}

func extractHostUUID(decorations []byte) string {
	if len(decorations) == 0 {
		return ""
	}
	var m map[string]string
	if err := json.Unmarshal(decorations, &m); err != nil {
		return ""
	}
	return m["host_uuid"]
}
