package ingest

import "encoding/json"

// PubSubMessage is the payload from a PubSub push subscription.
type PubSubMessage struct {
	Message struct {
		Data        []byte            `json:"data"`
		Attributes  map[string]string `json:"attributes"`
		MessageID   string            `json:"messageId"`
		PublishTime string            `json:"publishTime"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// subscriptionSuffix returns the last segment after the final "/" in the
// subscription resource name, which equals the subscription ID we set in TF.
func subscriptionSuffix(sub string) string {
	for i := len(sub) - 1; i >= 0; i-- {
		if sub[i] == '/' {
			return sub[i+1:]
		}
	}
	return sub
}

// ---- Fleet result log shapes ----

type resultEnvelope struct {
	Name           string          `json:"name"`
	QueryID        *int64          `json:"query_id"`
	HostIdentifier string          `json:"hostIdentifier"`
	CalendarTime   string          `json:"calendarTime"`
	UnixTime       int64           `json:"unixTime"`
	Epoch          int64           `json:"epoch"`
	Counter        int64           `json:"counter"`
	Decorations    json.RawMessage `json:"decorations"`

	// snapshot query
	Action   string            `json:"action"`
	Snapshot []json.RawMessage `json:"snapshot"`

	// differential query
	Columns json.RawMessage `json:"columns"`

	// batch-differential
	DiffResults *struct {
		Added   []json.RawMessage `json:"added"`
		Removed []json.RawMessage `json:"removed"`
	} `json:"diffResults"`
}

// ---- Fleet status log shape ----

type statusEnvelope struct {
	Severity     string          `json:"severity"`
	Filename     string          `json:"filename"`
	Line         int64           `json:"line"`
	Message      string          `json:"message"`
	Version      string          `json:"version"`
	Decorations  json.RawMessage `json:"decorations"`
}

// ---- Fleet audit log shape ----

type auditEnvelope struct {
	ID             *int64          `json:"id"`
	UUID           string          `json:"uuid"`
	CreatedAt      string          `json:"created_at"`
	Type           string          `json:"type"`
	ActorID        *int64          `json:"actor_id"`
	ActorFullName  string          `json:"actor_full_name"`
	ActorEmail     string          `json:"actor_email"`
	ActorAPIOnly   *bool           `json:"actor_api_only"`
	FleetInitiated *bool           `json:"fleet_initiated"`
	Details        json.RawMessage `json:"details"`
}
