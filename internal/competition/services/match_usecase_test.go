package services

import (
	"errors"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMatchRepo struct {
	match   *entity.Match
	getErr  error
	saveErr error
	saved   []*entity.Match
}

func (r *fakeMatchRepo) Save(match *entity.Match) error {
	r.saved = append(r.saved, match)
	return r.saveErr
}
func (r *fakeMatchRepo) GetByID(string) (*entity.Match, error) { return r.match, r.getErr }
func (r *fakeMatchRepo) ListActive() ([]*entity.Match, error)  { return nil, nil }
func (r *fakeMatchRepo) ListFinished(int) ([]*entity.Match, error) {
	return nil, nil
}
func (r *fakeMatchRepo) GetByBotID(string) ([]*entity.Match, error) { return nil, nil }
func (r *fakeMatchRepo) GetByTournamentID(string) ([]*entity.Match, error) {
	return nil, nil
}
func (r *fakeMatchRepo) GetAggregateStats(map[string]any) (map[string]any, error) {
	return nil, nil
}

type fakeBotRepo struct {
	bots        map[string]*entity.Bot
	getErr      error
	updateErr   error
	ratings     map[string]float64
	updatesSeen []string
}

func (r *fakeBotRepo) Save(bot *entity.Bot) error {
	if r.bots == nil {
		r.bots = map[string]*entity.Bot{}
	}
	r.bots[bot.ID] = bot
	return nil
}
func (r *fakeBotRepo) GetByID(id string) (*entity.Bot, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.bots[id], nil
}
func (r *fakeBotRepo) GetByUserID(string) ([]*entity.Bot, error) { return nil, nil }
func (r *fakeBotRepo) GetTop(int) ([]*entity.Bot, error)         { return nil, nil }
func (r *fakeBotRepo) UpdateRating(botID string, rating float64) error {
	if r.ratings == nil {
		r.ratings = map[string]float64{}
	}
	r.ratings[botID] = rating
	r.updatesSeen = append(r.updatesSeen, botID)
	return r.updateErr
}

type fakeScoreRepo struct {
	updates map[string]int
	err     error
}

func (r *fakeScoreRepo) UpdateScore(_ string, participantID string, points int) error {
	if r.updates == nil {
		r.updates = map[string]int{}
	}
	r.updates[participantID] = points
	return r.err
}
func (r *fakeScoreRepo) GetLeaderboard(string) ([]*entity.Participant, error) {
	return nil, nil
}

type fakeLeagueRepo struct {
	standings []*entity.Standing
	getErr    error
	saveErr   error
	saved     []*entity.Standing
}

func (r *fakeLeagueRepo) Save(*entity.League) error                       { return nil }
func (r *fakeLeagueRepo) GetByTournament(string) (*entity.League, error)  { return nil, nil }
func (r *fakeLeagueRepo) SaveDivision(*entity.Division) error             { return nil }
func (r *fakeLeagueRepo) GetDivisions(string) ([]*entity.Division, error) { return nil, nil }
func (r *fakeLeagueRepo) SaveStanding(standing *entity.Standing) error {
	r.saved = append(r.saved, standing)
	return r.saveErr
}
func (r *fakeLeagueRepo) GetStandings(string) ([]*entity.Standing, error) {
	return r.standings, r.getErr
}

func readyMatch(participants ...string) *entity.Match {
	return &entity.Match{
		ID:           "match-1",
		Status:       entity.MatchReady,
		Participants: participants,
	}
}

func botsWithRatings(ratings map[string]float64) *fakeBotRepo {
	bots := map[string]*entity.Bot{}
	for id, rating := range ratings {
		bots[id] = &entity.Bot{ID: id, Rating: rating}
	}
	return &fakeBotRepo{bots: bots}
}

func TestCompleteMatchLoadErrorIsReturned(t *testing.T) {
	repo := &fakeMatchRepo{getErr: errors.New("database down")}
	uc := NewMatchUsecase(repo, nil, &fakeBotRepo{}, &fakeLeagueRepo{}, nil)

	err := uc.CompleteMatch("match-1", "bot-1", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database down")
	assert.Empty(t, repo.saved)
}

func TestCompleteMatchRejectsWinnerOutsideParticipants(t *testing.T) {
	repo := &fakeMatchRepo{match: readyMatch("bot-1", "bot-2")}
	botRepo := botsWithRatings(map[string]float64{"bot-1": 1200, "bot-2": 1200})
	uc := NewMatchUsecase(repo, nil, botRepo, &fakeLeagueRepo{}, nil)

	err := uc.CompleteMatch("match-1", "bot-3", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a participant")
	assert.Empty(t, repo.saved)
	assert.Empty(t, botRepo.updatesSeen)
}

func TestCompleteMatchPersistsWinnerAndUpdatesRatingsScoresStandings(t *testing.T) {
	repo := &fakeMatchRepo{match: readyMatch("bot-1", "bot-2")}
	repo.match.DivisionID = "division-1"
	botRepo := botsWithRatings(map[string]float64{"bot-1": 1200, "bot-2": 1200})
	scoreRepo := &fakeScoreRepo{}
	leagueRepo := &fakeLeagueRepo{}
	uc := NewMatchUsecase(repo, scoreRepo, botRepo, leagueRepo, nil)

	metrics := map[string]any{
		"bot-1": map[string]any{"damage": 12.5},
	}
	require.NoError(t, uc.CompleteMatch("match-1", "bot-1", metrics))

	require.Len(t, repo.saved, 1)
	assert.Equal(t, entity.MatchFinished, repo.saved[0].Status)
	assert.Equal(t, "bot-1", repo.saved[0].WinnerID)
	assert.NotZero(t, repo.saved[0].EndedAt)
	assert.Equal(t, metrics, repo.saved[0].Metadata)

	assert.Greater(t, botRepo.ratings["bot-1"], float64(1200))
	assert.Less(t, botRepo.ratings["bot-2"], float64(1200))
	assert.Equal(t, map[string]int{"bot-1": 3}, scoreRepo.updates)

	require.Len(t, leagueRepo.saved, 2)
	byBot := map[string]*entity.Standing{}
	for _, standing := range leagueRepo.saved {
		byBot[standing.BotID] = standing
		assert.Equal(t, 1, standing.Matches)
		assert.Equal(t, "division-1", standing.DivisionID)
	}
	assert.Equal(t, 1, byBot["bot-1"].Wins)
	assert.Equal(t, 3, byBot["bot-1"].Points)
	assert.InDelta(t, 12.5, byBot["bot-1"].DanoCausado, 0.001)
	assert.Equal(t, 1, byBot["bot-2"].Losses)
	assert.Equal(t, 0, byBot["bot-2"].Points)
}

func TestCompleteMatchDraw(t *testing.T) {
	repo := &fakeMatchRepo{match: readyMatch("bot-1", "bot-2")}
	repo.match.DivisionID = "division-1"
	botRepo := botsWithRatings(map[string]float64{"bot-1": 1000, "bot-2": 1400})
	scoreRepo := &fakeScoreRepo{}
	leagueRepo := &fakeLeagueRepo{}
	uc := NewMatchUsecase(repo, scoreRepo, botRepo, leagueRepo, nil)

	require.NoError(t, uc.CompleteMatch("match-1", "", nil))

	assert.Equal(t, "", repo.saved[0].WinnerID)
	assert.Greater(t, botRepo.ratings["bot-1"], float64(1000))
	assert.Less(t, botRepo.ratings["bot-2"], float64(1400))
	assert.Equal(t, map[string]int{"bot-1": 1, "bot-2": 1}, scoreRepo.updates)
	for _, standing := range leagueRepo.saved {
		assert.Equal(t, 1, standing.Draws)
		assert.Equal(t, 1, standing.Points)
		assert.Equal(t, 0, standing.Wins)
		assert.Equal(t, 0, standing.Losses)
	}
}

func TestCompleteMatchPropagatesRepositoryErrors(t *testing.T) {
	t.Run("bot lookup", func(t *testing.T) {
		repo := &fakeMatchRepo{match: readyMatch("bot-1")}
		botRepo := botsWithRatings(map[string]float64{"bot-1": 1200})
		botRepo.getErr = errors.New("bot store down")
		uc := NewMatchUsecase(repo, nil, botRepo, &fakeLeagueRepo{}, nil)

		err := uc.CompleteMatch("match-1", "bot-1", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "bot store down")
		assert.Empty(t, repo.saved)
	})

	t.Run("match save", func(t *testing.T) {
		repo := &fakeMatchRepo{match: readyMatch("bot-1"), saveErr: errors.New("write failed")}
		uc := NewMatchUsecase(repo, nil, botsWithRatings(map[string]float64{"bot-1": 1200}), &fakeLeagueRepo{}, nil)

		err := uc.CompleteMatch("match-1", "bot-1", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "write failed")
	})

	t.Run("standings lookup", func(t *testing.T) {
		repo := &fakeMatchRepo{match: readyMatch("bot-1")}
		repo.match.DivisionID = "division-1"
		leagueRepo := &fakeLeagueRepo{getErr: errors.New("standings unavailable")}
		uc := NewMatchUsecase(repo, nil, botsWithRatings(map[string]float64{"bot-1": 1200}), leagueRepo, nil)

		err := uc.CompleteMatch("match-1", "bot-1", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "standings unavailable")
	})

	t.Run("standing save", func(t *testing.T) {
		repo := &fakeMatchRepo{match: readyMatch("bot-1")}
		repo.match.DivisionID = "division-1"
		leagueRepo := &fakeLeagueRepo{saveErr: errors.New("standing write failed")}
		uc := NewMatchUsecase(repo, nil, botsWithRatings(map[string]float64{"bot-1": 1200}), leagueRepo, nil)

		err := uc.CompleteMatch("match-1", "bot-1", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "standing write failed")
	})
}

func TestCompleteMatchIsIdempotent(t *testing.T) {
	match := readyMatch("bot-1")
	match.Status = entity.MatchFinished
	repo := &fakeMatchRepo{match: match}
	botRepo := botsWithRatings(map[string]float64{"bot-1": 1200})
	uc := NewMatchUsecase(repo, nil, botRepo, &fakeLeagueRepo{}, nil)

	require.NoError(t, uc.CompleteMatch("match-1", "bot-1", nil))
	assert.Empty(t, repo.saved)
	assert.Empty(t, botRepo.updatesSeen)
}
