package repository

import (
	"fmt"
	"sync"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type MemoryTournamentRepository struct {
	mu          sync.RWMutex
	tournaments map[string]*entity.Tournament
	brackets    map[string][]*entity.Bracket
}

func NewMemoryTournamentRepository() *MemoryTournamentRepository {
	return &MemoryTournamentRepository{
		tournaments: make(map[string]*entity.Tournament),
		brackets:    make(map[string][]*entity.Bracket),
	}
}

func (r *MemoryTournamentRepository) Save(t *entity.Tournament) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tournaments[t.ID] = t
	return nil
}

func (r *MemoryTournamentRepository) GetByID(id string) (*entity.Tournament, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tournaments[id]
	if !ok {
		return nil, fmt.Errorf("tournament not found")
	}
	return t, nil
}

func (r *MemoryTournamentRepository) ListByType(t entity.TournamentType) ([]*entity.Tournament, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Tournament
	for _, v := range r.tournaments {
		if v.Type == t {
			list = append(list, v)
		}
	}
	return list, nil
}

func (r *MemoryTournamentRepository) ListByHub(hubID string) ([]*entity.Tournament, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Tournament
	for _, v := range r.tournaments {
		if v.HubID == hubID {
			list = append(list, v)
		}
	}
	return list, nil
}

func (r *MemoryTournamentRepository) SaveBracket(b *entity.Bracket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.brackets[b.TournamentID] = append(r.brackets[b.TournamentID], b)
	return nil
}

func (r *MemoryTournamentRepository) GetBrackets(tournamentID string) ([]*entity.Bracket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.brackets[tournamentID], nil
}

type MemoryLeagueRepository struct {
	mu        sync.RWMutex
	leagues   map[string]*entity.League
	divisions map[string][]*entity.Division
	standings map[string][]*entity.Standing
}

func NewMemoryLeagueRepository() *MemoryLeagueRepository {
	return &MemoryLeagueRepository{
		leagues:   make(map[string]*entity.League),
		divisions: make(map[string][]*entity.Division),
		standings: make(map[string][]*entity.Standing),
	}
}

func (r *MemoryLeagueRepository) Save(l *entity.League) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leagues[l.ID] = l
	return nil
}

func (r *MemoryLeagueRepository) GetByTournament(tournamentID string) (*entity.League, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, l := range r.leagues {
		if l.TournamentID == tournamentID {
			return l, nil
		}
	}
	return nil, fmt.Errorf("league not found")
}

func (r *MemoryLeagueRepository) SaveDivision(d *entity.Division) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.divisions[d.LeagueID] = append(r.divisions[d.LeagueID], d)
	return nil
}

func (r *MemoryLeagueRepository) GetDivisions(leagueID string) ([]*entity.Division, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.divisions[leagueID], nil
}

func (r *MemoryLeagueRepository) SaveStanding(s *entity.Standing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Find and update or append
	standings := r.standings[s.DivisionID]
	for i, existing := range standings {
		if existing.BotID == s.BotID {
			standings[i] = s
			return nil
		}
	}
	r.standings[s.DivisionID] = append(r.standings[s.DivisionID], s)
	return nil
}

func (r *MemoryLeagueRepository) GetStandings(divisionID string) ([]*entity.Standing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.standings[divisionID], nil
}

type MemoryMatchRepository struct {
	mu      sync.RWMutex
	matches map[string]*entity.Match
}

func NewMemoryMatchRepository() *MemoryMatchRepository {
	return &MemoryMatchRepository{
		matches: make(map[string]*entity.Match),
	}
}

func (r *MemoryMatchRepository) Save(match *entity.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[match.ID] = match
	return nil
}

func (r *MemoryMatchRepository) GetByID(id string) (*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.matches[id]
	if !ok {
		return nil, fmt.Errorf("match not found")
	}
	return m, nil
}

func (r *MemoryMatchRepository) ListActive() ([]*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Match
	for _, m := range r.matches {
		if m.Status == entity.MatchRunning {
			list = append(list, m)
		}
	}
	return list, nil
}

func (r *MemoryMatchRepository) ListFinished(limit int) ([]*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Match
	for _, m := range r.matches {
		if m.Status == entity.MatchFinished || m.Status == entity.MatchCompleted {
			list = append(list, m)
			if limit > 0 && len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryMatchRepository) GetByBotID(botID string) ([]*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Match
	for _, m := range r.matches {
		for _, p := range m.Participants {
			if p == botID {
				list = append(list, m)
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryMatchRepository) GetByTournamentID(tournamentID string) ([]*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Match
	for _, m := range r.matches {
		if m.TournamentID == tournamentID {
			list = append(list, m)
		}
	}
	return list, nil
}

func (r *MemoryMatchRepository) GetAggregateStats(query map[string]any) (map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	winners := make([]string, 0)
	total := 0
	for _, m := range r.matches {
		if m.Status == entity.MatchCompleted {
			total++
			winners = append(winners, m.WinnerID)
		}
	}

	return map[string]any{
		"total_matches": total,
		"winners":       winners,
	}, nil
}

type MemoryBotRepository struct {
	mu   sync.RWMutex
	bots map[string]*entity.Bot
}

func NewMemoryBotRepository() *MemoryBotRepository {
	return &MemoryBotRepository{
		bots: make(map[string]*entity.Bot),
	}
}

func (r *MemoryBotRepository) Save(bot *entity.Bot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bots[bot.ID] = bot
	return nil
}

func (r *MemoryBotRepository) GetByID(id string) (*entity.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bots[id]
	if !ok {
		return nil, fmt.Errorf("bot not found")
	}
	return b, nil
}

func (r *MemoryBotRepository) GetByUserID(userID string) ([]*entity.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Bot
	for _, b := range r.bots {
		if b.UserID == userID {
			list = append(list, b)
		}
	}
	return list, nil
}

func (r *MemoryBotRepository) UpdateRating(botID string, newRating float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bots[botID]; ok {
		b.Rating = newRating
	}
	return nil
}

func (r *MemoryBotRepository) SaveVersion(v *entity.BotVersion) error {
	return nil
}

func (r *MemoryBotRepository) GetVersions(botID string) ([]*entity.BotVersion, error) {
	return nil, nil
}

type MemoryScoreRepository struct {
	mu     sync.RWMutex
	scores map[string]map[string]int // challengeID -> participantID -> points
}

func NewMemoryScoreRepository() *MemoryScoreRepository {
	return &MemoryScoreRepository{
		scores: make(map[string]map[string]int),
	}
}

func (r *MemoryScoreRepository) UpdateScore(matchID string, participantID string, points int) error {
	return nil
}

func (r *MemoryScoreRepository) GetLeaderboard(challengeID string) ([]*entity.Participant, error) {
	return nil, nil
}

type MemoryHubRepository struct {
	mu          sync.RWMutex
	hubs        map[string]*entity.Hub
	memberships map[string][]*entity.HubMembership
}

func NewMemoryHubRepository() *MemoryHubRepository {
	return &MemoryHubRepository{
		hubs:        make(map[string]*entity.Hub),
		memberships: make(map[string][]*entity.HubMembership),
	}
}

func (r *MemoryHubRepository) Save(hub *entity.Hub) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hubs[hub.ID] = hub
	return nil
}

func (r *MemoryHubRepository) GetByID(id string) (*entity.Hub, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hubs[id]
	if !ok {
		return nil, fmt.Errorf("hub not found")
	}
	return h, nil
}

func (r *MemoryHubRepository) ListPublic() ([]*entity.Hub, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Hub
	for _, h := range r.hubs {
		if h.Visibility == entity.HubPublic {
			list = append(list, h)
		}
	}
	return list, nil
}

func (r *MemoryHubRepository) ListByUser(userID string) ([]*entity.Hub, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Hub
	for hubID, members := range r.memberships {
		for _, m := range members {
			if m.UserID == userID {
				list = append(list, r.hubs[hubID])
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryHubRepository) SaveMembership(m *entity.HubMembership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberships[m.HubID] = append(r.memberships[m.HubID], m)
	return nil
}

func (r *MemoryHubRepository) GetMemberships(hubID string) ([]*entity.HubMembership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.memberships[hubID], nil
}
