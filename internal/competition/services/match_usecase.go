package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/service"
)

type MatchUsecase struct {
	matchRepo  repository.MatchRepository
	scoreRepo  repository.ScoreRepository
	botRepo    repository.BotRepository
	leagueRepo repository.LeagueRepository
	judge      *JudgeService
}

func NewMatchUsecase(mr repository.MatchRepository, sr repository.ScoreRepository, br repository.BotRepository, lr repository.LeagueRepository, j *JudgeService) *MatchUsecase {
	return &MatchUsecase{
		matchRepo:  mr,
		scoreRepo:  sr,
		botRepo:    br,
		leagueRepo: lr,
		judge:      j,
	}
}

func (u *MatchUsecase) CreateMatch(challengeID string, participants []string) (*entity.Match, error) {
	match := &entity.Match{
		ID:           fmt.Sprintf("m_%d", time.Now().UnixNano()),
		ChallengeID:  challengeID,
		Status:       entity.MatchReady,
		Participants: participants,
		StartedAt:    time.Now().Unix(),
	}

	if err := u.matchRepo.Save(match); err != nil {
		return nil, err
	}

	return match, nil
}

func (u *MatchUsecase) StartMatch(ctx context.Context, matchID string) error {
	match, err := u.matchRepo.GetByID(matchID)
	if err != nil {
		return err
	}

	if match.Status != entity.MatchReady {
		return fmt.Errorf("match is not ready")
	}

	// 1. Run the match using the Judge
	if err := u.judge.RunMatch(ctx, match); err != nil {
		return fmt.Errorf("judge failed to start match: %w", err)
	}

	return nil
}

func (u *MatchUsecase) CompleteMatch(matchID string, winnerID string, metrics map[string]any) error {
	match, err := u.matchRepo.GetByID(matchID)
	if err != nil {
		return err
	}

	match.Status = entity.MatchFinished
	match.EndedAt = time.Now().Unix()
	match.Metadata = metrics
	u.matchRepo.Save(match)

	// Update ELO and Standings
	if len(match.Participants) > 0 {
		var ratings []float64
		winnerIndex := -1
		for i, pID := range match.Participants {
			bot, _ := u.botRepo.GetByID(pID)
			ratings = append(ratings, bot.Rating)
			if pID == winnerID {
				winnerIndex = i
			}
		}

		// 1. Calculate and Update ELO
		newRatings := service.CalculateMultiELO(ratings, winnerIndex)
		for i, pID := range match.Participants {
			u.botRepo.UpdateRating(pID, newRatings[i])
		}

		// 2. Update League Standings if DivisionID is present
		if match.DivisionID != "" {
			standings, _ := u.leagueRepo.GetStandings(match.DivisionID)
			for _, pID := range match.Participants {
				var standing *entity.Standing
				for _, s := range standings {
					if s.BotID == pID {
						standing = s
						break
					}
				}

				if standing == nil {
					standing = &entity.Standing{
						DivisionID: match.DivisionID,
						BotID:      pID,
					}
				}

				standing.Matches++
				if pID == winnerID {
					standing.Points += 3
					standing.Wins++
				} else if winnerID == "" {
					standing.Points += 1
					standing.Draws++
				} else {
					standing.Losses++
				}

				// Update metrics if available
				if m, ok := match.Metadata[pID].(map[string]any); ok {
					if d, ok := m["damage"].(float64); ok {
						standing.DanoCausado += d
					}
				}

				u.leagueRepo.SaveStanding(standing)
			}
		}
	}

	return nil
}
func (u *MatchUsecase) RegisterBot(userID, name, version, sourceCode, language string) (*entity.Bot, *entity.BotVersion, error) {
	bot := &entity.Bot{
		ID:        fmt.Sprintf("bot_%d", time.Now().UnixNano()),
		UserID:    userID,
		Name:      name,
		Rating:    1200, // Default ranking
		CreatedAt: time.Now().Unix(),
	}

	if err := u.botRepo.Save(bot); err != nil {
		return nil, nil, err
	}

	botVersion := &entity.BotVersion{
		ID:        fmt.Sprintf("v_%d", time.Now().UnixNano()),
		BotID:     bot.ID,
		Version:   version,
		SourceURL: sourceCode, // In v1 we store source directly or mock S3 URL
		CreatedAt: time.Now().Unix(),
	}

	// Assuming we might need a separate repository for versions,
	// or we can add it to BotRepository if it handles both.
	// For now, let's assume Save in BotRepository handles associated entity or just mock it.

	return bot, botVersion, nil
}
