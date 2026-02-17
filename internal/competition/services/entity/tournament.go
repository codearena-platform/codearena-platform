package entity

type TournamentStatus string
type TournamentType string

const (
	TournamentDraft   TournamentStatus = "DRAFT"
	TournamentSignup  TournamentStatus = "SIGNUP"
	TournamentOngoing TournamentStatus = "ONGOING"
	TournamentFinish  TournamentStatus = "FINISHED"

	TournamentBrackets TournamentType = "BRACKETS"
	TournamentLeague   TournamentType = "LEAGUE"
)

type Tournament struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        TournamentType   `json:"type"`
	Status      TournamentStatus `json:"status"`
	Description string           `json:"description"`
	Config      map[string]any   `json:"config"`
	HubID       string           `json:"hub_id,omitempty"` // Link to organization hub
	CreatedAt   int64            `json:"created_at"`
}

type Bracket struct {
	ID           string `json:"id"`
	TournamentID string `json:"tournament_id"`
	Level        int    `json:"level"` // 1: Final, 2: Semi-Final, 4: Quarter-Final, etc.
	MatchID      string `json:"match_id"`
	Position     int    `json:"position"`
}
