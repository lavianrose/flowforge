package models

type HealthStats struct {
	ActiveRuns      int64              `json:"active_runs"`
	SuccessRate     float64            `json:"success_rate"`
	FailureRate     float64            `json:"failure_rate"`
	AvgDuration     float64            `json:"avg_duration_seconds"`
	TotalRuns24h    int64              `json:"total_runs_24h"`
	SuccessRuns24h  int64              `json:"success_runs_24h"`
	FailedRuns24h   int64              `json:"failed_runs_24h"`
	HourlyStats     []HourlyStats      `json:"hourly_stats"`
}

type HourlyStats struct {
	Hour         int     `json:"hour"`         // 0-23
	TotalRuns    int64   `json:"total_runs"`
	SuccessRuns  int64   `json:"success_runs"`
	FailedRuns   int64   `json:"failed_runs"`
	AvgDuration  float64 `json:"avg_duration"`
}
