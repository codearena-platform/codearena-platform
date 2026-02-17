package entity

type DivisionName string

const (
	DivisionSerieA  DivisionName = "SERIE_A"
	DivisionSerieB  DivisionName = "SERIE_B"
	DivisionSerieC  DivisionName = "SERIE_C"
	DivisionAmateur DivisionName = "AMATEUR"
)

type League struct {
	ID           string `json:"id"`
	TournamentID string `json:"tournament_id"` // Link to main tournament
	Season       string `json:"season"`        // e.g., "2026-Q1"
}

type Division struct {
	ID              string       `json:"id"`
	LeagueID        string       `json:"league_id"`
	Name            DivisionName `json:"name"`
	Capacity        int          `json:"capacity"`
	PromotionSlots  int          `json:"promotion_slots"`
	RelegationSlots int          `json:"relegation_slots"`
}

type Standing struct {
	DivisionID  string  `json:"division_id"`
	BotID       string  `json:"bot_id"`
	Points      int     `json:"points"`
	Wins        int     `json:"wins"`
	Draws       int     `json:"draws"`
	Losses      int     `json:"losses"`
	Matches     int     `json:"matches"`
	DanoCausado float64 `json:"dano_causado"`
}
