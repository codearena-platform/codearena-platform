package entity

type ClassName string

const (
	ClassTank   ClassName = "Tank"
	ClassScout  ClassName = "Scout"
	ClassSniper ClassName = "Sniper"
)

type ClassStats struct {
	Class    ClassName `json:"class"`
	WinRate  float64   `json:"win_rate"`
	PickRate float64   `json:"pick_rate"`
	AvgADR   float64   `json:"avg_adr"` // Average Damage per Round
}

type GlobalStats struct {
	TotalMatches int                       `json:"total_matches"`
	Classes      map[ClassName]*ClassStats `json:"classes"`
	Trend        []float64                 `json:"trend"` // Successive ADRs for trend analysis
}

type BotAnalytics struct {
	BotID           string  `json:"bot_id"`
	TotalMatches    int     `json:"total_matches"`
	WinCount        int     `json:"win_count"`
	WinRate         float64 `json:"win_rate"`
	ADR             float64 `json:"adr"`
	AvgSurvivalTime float64 `json:"avg_survival_time"`
	Accuracy        float64 `json:"accuracy"`
}
