package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
	"github.com/google/uuid"
)

type BotUsecase struct {
	botRepo         repository.BotRepository
	versionRepo     repository.BotVersionRepository
	plagiarismSvc   *PlagiarismService
	signatureMemory []entity.BotSignature // In memory cache of signatures for demo
}

// NewBotUsecase creates a new service that manages bots and validates plagiarism
func NewBotUsecase(br repository.BotRepository, vr repository.BotVersionRepository, pSvc *PlagiarismService) *BotUsecase {
	return &BotUsecase{
		botRepo:         br,
		versionRepo:     vr,
		plagiarismSvc:   pSvc,
		signatureMemory: []entity.BotSignature{},
	}
}

// SubmitBot Creates or Updates a bot with a new piece of code
func (u *BotUsecase) SubmitBot(ctx context.Context, userID, botName, chassis, sourceCode string) (*entity.Bot, error) {
	// 1. Check Plagiarism
	valid, report := u.plagiarismSvc.CheckCode(ctx, sourceCode, u.signatureMemory)
	if !valid {
		return nil, fmt.Errorf("plagiarism detected (%.2f%% similarity with bot %s). Submission rejected", report.Similarity*100, report.MatchedBotID)
	}

	botID := uuid.New().String()

	// 2. Generate and save signature for future checks
	newSignature := u.plagiarismSvc.GenerateSignature(sourceCode)
	newSignature.ID = uuid.New().String()
	newSignature.BotID = botID
	newSignature.CreatedAt = time.Now()

	u.signatureMemory = append(u.signatureMemory, newSignature)

	// 3. Persist Bot
	bot := &entity.Bot{
		ID:        botID,
		Name:      botName,
		UserID:    userID,
		Rating:    1200.0, // Default starting Elo
		Matches:   0,
		Wins:      0,
		CreatedAt: time.Now().Unix(),
	}

	if err := u.botRepo.Save(bot); err != nil {
		return nil, fmt.Errorf("failed to save bot profile: %w", err)
	}

	// 4. Persist Version
	version := &entity.BotVersion{
		ID:        uuid.New().String(),
		BotID:     botID,
		Version:   "v1",
		SourceURL: "source_attached",                           // In a real system, we'd upload to S3 here
		Hash:      fmt.Sprintf("%d", newSignature.HashTree[0]), // simple hash representation
		CreatedAt: time.Now().Unix(),
	}

	if err := u.versionRepo.Save(version); err != nil {
		return nil, fmt.Errorf("failed to save bot version: %w", err)
	}

	return bot, nil
}
