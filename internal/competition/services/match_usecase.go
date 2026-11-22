package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/service"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("competition-service")

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
	_, span := tracer.Start(context.Background(), "CreateMatch")
	defer span.End()

	match := &entity.Match{
		ID:           "m_" + uuid.New().String(),
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
	ctx, span := tracer.Start(ctx, "StartMatch")
	defer span.End()

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
		return fmt.Errorf("load match %s: %w", matchID, err)
	}
	if match == nil {
		return fmt.Errorf("match %s not found", matchID)
	}
	// Completion is idempotent: a retry after a successful save must not apply
	// rating or standing changes a second time.
	if match.Status == entity.MatchFinished || match.Status == entity.MatchCompleted {
		return nil
	}

	winnerIndex := -1
	seen := make(map[string]bool, len(match.Participants))
	for i, participantID := range match.Participants {
		if participantID == "" || seen[participantID] {
			return fmt.Errorf("match %s has invalid participants", matchID)
		}
		seen[participantID] = true
		if participantID == winnerID {
			winnerIndex = i
		}
	}
	if winnerID != "" && winnerIndex < 0 {
		return fmt.Errorf("winner %s is not a participant in match %s", winnerID, matchID)
	}
	draw := winnerID == ""

	ratings := make([]float64, 0, len(match.Participants))
	for _, participantID := range match.Participants {
		bot, err := u.botRepo.GetByID(participantID)
		if err != nil {
			return fmt.Errorf("load bot %s: %w", participantID, err)
		}
		if bot == nil {
			return fmt.Errorf("bot %s not found", participantID)
		}
		ratings = append(ratings, bot.Rating)
	}

	match.Status = entity.MatchFinished
	match.EndedAt = time.Now().Unix()
	match.Metadata = metrics
	match.WinnerID = winnerID
	if err := u.matchRepo.Save(match); err != nil {
		return fmt.Errorf("save match %s: %w", matchID, err)
	}

	if len(match.Participants) == 0 {
		return nil
	}

	var newRatings []float64
	if draw {
		newRatings = service.CalculateMultiELODraw(ratings)
	} else {
		newRatings = service.CalculateMultiELO(ratings, winnerIndex)
	}
	for i, participantID := range match.Participants {
		if err := u.botRepo.UpdateRating(participantID, newRatings[i]); err != nil {
			return fmt.Errorf("update rating for bot %s: %w", participantID, err)
		}
	}

	if u.scoreRepo != nil {
		for _, participantID := range match.Participants {
			points := 0
			switch {
			case draw:
				points = 1
			case participantID == winnerID:
				points = 3
			}
			if points == 0 {
				continue
			}
			if err := u.scoreRepo.UpdateScore(matchID, participantID, points); err != nil {
				return fmt.Errorf("update score for bot %s: %w", participantID, err)
			}
		}
	}

	if match.DivisionID != "" {
		standings, err := u.leagueRepo.GetStandings(match.DivisionID)
		if err != nil {
			return fmt.Errorf("load standings for division %s: %w", match.DivisionID, err)
		}
		byBot := make(map[string]*entity.Standing, len(standings))
		for _, standing := range standings {
			if standing != nil {
				byBot[standing.BotID] = standing
			}
		}

		for _, participantID := range match.Participants {
			standing := byBot[participantID]
			if standing == nil {
				standing = &entity.Standing{DivisionID: match.DivisionID, BotID: participantID}
				byBot[participantID] = standing
			}
			standing.Matches++
			switch {
			case draw:
				standing.Points++
				standing.Draws++
			case participantID == winnerID:
				standing.Points += 3
				standing.Wins++
			default:
				standing.Losses++
			}

			if detail, ok := metrics[participantID].(map[string]any); ok {
				if damage, ok := detail["damage"].(float64); ok {
					standing.DanoCausado += damage
				}
			}

			if err := u.leagueRepo.SaveStanding(standing); err != nil {
				return fmt.Errorf("save standing for bot %s: %w", participantID, err)
			}
		}
	}

	return nil
}
func (u *MatchUsecase) RegisterBot(userID, name, version, sourceCode, language string) (*entity.Bot, *entity.BotVersion, error) {
	bot := &entity.Bot{
		ID:        "bot_" + uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Rating:    1200, // Default ranking
		CreatedAt: time.Now().Unix(),
	}

	if err := u.botRepo.Save(bot); err != nil {
		return nil, nil, err
	}

	botVersion := &entity.BotVersion{
		ID:        "v_" + uuid.New().String(),
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
