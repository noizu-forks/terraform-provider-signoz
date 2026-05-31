package model

// DataRetention maps the GET response from /api/v1/settings/ttl.
// Uses the expected_* fields to avoid drift during TTL change processing.
type DataRetention struct {
	MetricsTTLDurationHrs     int `json:"expected_metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs int `json:"expected_metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        int `json:"expected_logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    int `json:"expected_logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      int `json:"expected_traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  int `json:"expected_traces_move_ttl_duration_hrs"`
}

// DataRetentionUpdate is the POST payload for /api/v1/settings/ttl.
// One POST per signal type (metrics, logs, traces).
type DataRetentionUpdate struct {
	Type                string `json:"type"`
	Duration            string `json:"duration"`
	ColdStorageDuration string `json:"cold_storage_duration"`
}
