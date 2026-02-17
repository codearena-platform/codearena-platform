package repository

import "github.com/codearena-platform/codearena-platform/internal/competition/services/entity"

type MatchRepository interface {
	Save(match *entity.Match) error
	GetByID(id string) (*entity.Match, error)
	ListActive() ([]*entity.Match, error)
	ListFinished(limit int) ([]*entity.Match, error)
	GetByBotID(botID string) ([]*entity.Match, error)
	GetByTournamentID(tournamentID string) ([]*entity.Match, error)
	GetAggregateStats(query map[string]any) (map[string]any, error)
}

type HubRepository interface {
	Save(hub *entity.Hub) error
	GetByID(id string) (*entity.Hub, error)
	ListPublic() ([]*entity.Hub, error)
	ListByUser(userID string) ([]*entity.Hub, error)
	SaveMembership(m *entity.HubMembership) error
	GetMemberships(hubID string) ([]*entity.HubMembership, error)
}

type ScoreRepository interface {
	UpdateScore(matchID string, participantID string, points int) error
	GetLeaderboard(challengeID string) ([]*entity.Participant, error)
}

type TournamentRepository interface {
	Save(tournament *entity.Tournament) error
	GetByID(id string) (*entity.Tournament, error)
	ListByType(t entity.TournamentType) ([]*entity.Tournament, error)
	ListByHub(hubID string) ([]*entity.Tournament, error)
	SaveBracket(bracket *entity.Bracket) error
	GetBrackets(tournamentID string) ([]*entity.Bracket, error)
}

type LeagueRepository interface {
	Save(league *entity.League) error
	GetByTournament(tournamentID string) (*entity.League, error)
	SaveDivision(division *entity.Division) error
	GetDivisions(leagueID string) ([]*entity.Division, error)
	SaveStanding(standing *entity.Standing) error
	GetStandings(divisionID string) ([]*entity.Standing, error)
}
