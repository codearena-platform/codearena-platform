package entity

type MatchStatus string

const (
	MatchReady     MatchStatus = "READY"
	MatchRunning   MatchStatus = "RUNNING"
	MatchFinished  MatchStatus = "FINISHED"
	MatchCompleted MatchStatus = "COMPLETED"
	MatchFailed    MatchStatus = "FAILED"
)

// Match is persisted in MongoDB. BSON tags are the source of truth for query
// field names; keep the mongo repository constants in sync with them.
type Match struct {
	ID              string         `bson:"id"`
	TournamentID    string         `bson:"tournament_id"` // Optional: link to a tournament
	DivisionID      string         `bson:"division_id"`   // Optional: link to a league division
	ChallengeID     string         `bson:"challenge_id"`
	Status          MatchStatus    `bson:"status"`
	Participants    []string       `bson:"participants"`
	WinnerID        string         `bson:"winner_id"`
	StartedAt       int64          `bson:"started_at"`
	EndedAt         int64          `bson:"ended_at"`
	FinishedAt      int64          `bson:"finished_at"`
	EngineURL       string         `bson:"engine_url"`       // The URL of the simulation engine that ran the match
	ReplayAvailable bool           `bson:"replay_available"` // Whether a detailed replay is available
	Metadata        map[string]any `bson:"metadata"`         // For telemetry (energy, damage, etc.)
}

type Participant struct {
	ID     string `bson:"id" json:"id"`
	UserID string `bson:"user_id" json:"user_id"`
	Name   string `bson:"name" json:"name"`
	Score  int    `bson:"score" json:"score"`
}
