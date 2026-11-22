package services

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeVersionRepo struct {
	versions []*entity.BotVersion
}

func (r *fakeVersionRepo) Save(version *entity.BotVersion) error {
	r.versions = append(r.versions, version)
	return nil
}
func (r *fakeVersionRepo) GetLatest(string) (*entity.BotVersion, error) { return nil, nil }
func (r *fakeVersionRepo) GetVersions(string) ([]*entity.BotVersion, error) {
	return nil, nil
}

func TestSignatureHash(t *testing.T) {
	assert.Empty(t, signatureHash(entity.BotSignature{}))
	assert.NotEmpty(t, signatureHash(entity.BotSignature{HashTree: []uint32{1, 2, 3}}))
	assert.Equal(t,
		signatureHash(entity.BotSignature{HashTree: []uint32{1, 2, 3}}),
		signatureHash(entity.BotSignature{HashTree: []uint32{1, 2, 3}}),
	)
}

func TestSubmitBotPersistsStableFingerprintHash(t *testing.T) {
	botRepo := botsWithRatings(nil)
	versionRepo := &fakeVersionRepo{}
	plagiarismSvc := NewPlagiarismService()
	uc := NewBotUsecase(botRepo, versionRepo, plagiarismSvc)

	bot, err := uc.SubmitBot(context.Background(), "user-1", "bot", "chassis", "move()")

	require.NoError(t, err)
	assert.NotNil(t, bot)
	require.Len(t, versionRepo.versions, 1)
	assert.Equal(t, bot.ID, versionRepo.versions[0].BotID)
	assert.Equal(t, signatureHash(plagiarismSvc.GenerateSignature("move()")), versionRepo.versions[0].Hash)
}
