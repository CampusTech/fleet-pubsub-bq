package ingest

import "testing"

func TestSubscriptionSuffix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"projects/my-project/subscriptions/fleet-result-logs-sub", "fleet-result-logs-sub"},
		{"fleet-result-logs-sub", "fleet-result-logs-sub"},
		{"", ""},
	}
	for _, c := range cases {
		if got := subscriptionSuffix(c.in); got != c.want {
			t.Errorf("subscriptionSuffix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
