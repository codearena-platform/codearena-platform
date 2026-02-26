package services

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/stretchr/testify/assert"
)

func TestPlagiarismService_CheckCode(t *testing.T) {
	service := NewPlagiarismService()

	originalCode := `
	function update() {
		// bot logic
		let radar = getRadar();
		if (radar.enemy) {
			fire(radar.enemy.distance);
		}
		moveForward();
	}
	`

	// Complete copy with just different variable names and comments
	plagiarizedCode := `
	/* 
	   My awesome bot 
	*/
	function update() {
		let scanner = getRadar();
		if (scanner.enemy) {
			fire(scanner.enemy.distance);
		}
		moveForward();
	}
	`

	// Completely different logic
	differentCode := `
	function update() {
		turnLeft(90);
		fire(100);
		turnRight(45);
	}
	`

	originalSignature := service.GenerateSignature(originalCode)
	originalSignature.BotID = "bot-001"

	ctx := context.Background()

	t.Run("Should detect plagiarized code", func(t *testing.T) {
		valid, report := service.CheckCode(ctx, plagiarizedCode, []entity.BotSignature{originalSignature})

		assert.False(t, valid, "Plagiarized code should be rejected")
		assert.True(t, report.IsPlagiarized)
		assert.Equal(t, "bot-001", report.MatchedBotID)
		assert.Greater(t, report.Similarity, PlagiarismThreshold, "Similarity should be higher than threshold")
	})

	t.Run("Should accept completely different code", func(t *testing.T) {
		valid, report := service.CheckCode(ctx, differentCode, []entity.BotSignature{originalSignature})

		assert.True(t, valid, "Different code should be accepted")
		assert.False(t, report.IsPlagiarized)
		assert.Less(t, report.Similarity, PlagiarismThreshold, "Similarity should be lower than threshold")
	})
}
