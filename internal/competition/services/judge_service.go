package services

import (
	"context"
	"fmt"
	"log"

	pb "github.com/codearena-platform/codearena-core/pkg/api/v1"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type JudgeService struct {
	matchRepo repository.MatchRepository
	botRepo   repository.BotRepository
	simClient pb.SimulationServiceClient
	simAddr   string
}

func NewJudgeService(mr repository.MatchRepository, br repository.BotRepository, simAddr string) *JudgeService {
	conn, err := grpc.Dial(simAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("WARNING: Judge could not connect to Simulation Engine at %s: %v", simAddr, err)
	}

	return &JudgeService{
		matchRepo: mr,
		botRepo:   br,
		simClient: pb.NewSimulationServiceClient(conn),
		simAddr:   simAddr,
	}
}

func (s *JudgeService) RunMatch(ctx context.Context, match *entity.Match) error {
	log.Printf("Judge: Preparing execution for match %s", match.ID)

	if s.simClient == nil {
		return fmt.Errorf("simulation client not initialized")
	}

	// 1. Transform Entity Match to Protobuf ArenaConfig
	config := &pb.ArenaConfig{
		Id:   match.ID,
		Name: fmt.Sprintf("Match_%s", match.ID),
		// In a real scenario, we would populate participants, bots, and rules here
		// For now, using match ID to bridge the session
	}

	// 2. Trigger Simulation
	resp, err := s.simClient.StartSimulation(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to start simulation: %w", err)
	}

	log.Printf("Judge: Simulation triggered for match %s. Status: %v", match.ID, resp.Status)

	// 3. Mark Match as Running
	match.Status = entity.MatchRunning
	match.EngineURL = s.simAddr
	return s.matchRepo.Save(match)
}

func (s *JudgeService) ProcessResult(matchID string, winnerID string, metrics map[string]any) error {
	log.Printf("Judge: Processing result for match %s. Winner: %s", matchID, winnerID)

	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return err
	}

	match.Status = entity.MatchFinished
	match.WinnerID = winnerID
	match.Metadata = metrics
	match.EndedAt = 0 // Update time here if needed

	// Basic ELO calculation
	// In a real scenario, we'll fetch both bots and calculate for both
	// For MVP: Simple K-factor increment if we find the winner
	bot, err := s.botRepo.GetByID(winnerID)
	if err == nil && bot != nil {
		K := 32.0
		// Assuming expected outcome was 0.5 (equal level)
		expected := 0.5
		newRating := bot.Rating + K*(1.0-expected)

		s.botRepo.UpdateRating(bot.ID, newRating)
		log.Printf("Judge: Updated ELO for %s: %.2f -> %.2f", winnerID, bot.Rating, newRating)
	}

	return s.matchRepo.Save(match)
}
