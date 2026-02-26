package services

import (
	"context"
	"hash/fnv"
	"regexp"
	"strings"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

// Plagiarism Threshold: > 85% overlap triggers a block
const PlagiarismThreshold = 0.85

// PlagiarismService implements a Rolling Hash + Winnowing algorithm
// to safely detect source-code similarity ignoring spaces and variables formats.
type PlagiarismService struct {
	// In a real scenario, this would hold a repository to load all previously saved BotSignatures
	// signatureRepo repository.BotSignatureRepository
}

func NewPlagiarismService() *PlagiarismService {
	return &PlagiarismService{}
}

// CheckCode evaluates source code against a given set of existing signatures.
// If similarity > threshold, returns false (not accepted) and the PlagiarismReport.
func (s *PlagiarismService) CheckCode(ctx context.Context, sourceCode string, existingSignatures []entity.BotSignature) (bool, *entity.PlagiarismReport) {
	// 1. Generate signature for the new code
	newSignature := s.GenerateSignature(sourceCode)

	highestSimilarity := 0.0
	var matchedBotID string

	// 2. Compare against all existing
	for _, sig := range existingSignatures {
		similarity := s.calculateJaccardSimilarity(newSignature.HashTree, sig.HashTree)
		if similarity > highestSimilarity {
			highestSimilarity = similarity
			matchedBotID = sig.BotID
		}
	}

	report := &entity.PlagiarismReport{
		IsPlagiarized: highestSimilarity >= PlagiarismThreshold,
		Similarity:    highestSimilarity,
		MatchedBotID:  matchedBotID,
	}

	return !report.IsPlagiarized, report
}

// GenerateSignature cleans the code and generates a Winnowing fingerprint
func (s *PlagiarismService) GenerateSignature(sourceCode string) entity.BotSignature {
	cleaned := s.tokenizeAndClean(sourceCode)
	kGrams := s.generateKGrams(cleaned, 5) // k=5 chars
	hashes := s.hashKGrams(kGrams)
	fingerprint := s.winnowing(hashes, 4) // window size w=4

	return entity.BotSignature{
		HashTree: fingerprint,
	}
}

// tokenizeAndClean removes whitespaces, newlines and comments to focus purely on logic flow
func (s *PlagiarismService) tokenizeAndClean(code string) string {
	// Remove single-line comments //
	reSingle := regexp.MustCompile(`//.*`)
	code = reSingle.ReplaceAllString(code, "")

	// Remove multi-line comments /* */
	reMulti := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	code = reMulti.ReplaceAllString(code, "")

	// Remove all whitespaces and newlines
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "\n", "")
	code = strings.ReplaceAll(code, "\t", "")
	code = strings.ReplaceAll(code, "\r", "")

	return strings.ToLower(code)
}

func (s *PlagiarismService) generateKGrams(text string, k int) []string {
	var kGrams []string
	if len(text) < k {
		return []string{text}
	}
	for i := 0; i <= len(text)-k; i++ {
		kGrams = append(kGrams, text[i:i+k])
	}
	return kGrams
}

func (s *PlagiarismService) hashKGrams(kgrams []string) []uint32 {
	hashes := make([]uint32, len(kgrams))
	for i, gram := range kgrams {
		h := fnv.New32a()
		h.Write([]byte(gram))
		hashes[i] = h.Sum32()
	}
	return hashes
}

// winnowing creates a robust fingerprint by selecting the minimum hash in every window
func (s *PlagiarismService) winnowing(hashes []uint32, w int) []uint32 {
	var fingerprint []uint32
	if len(hashes) == 0 {
		return fingerprint
	}

	if len(hashes) <= w {
		min := hashes[0]
		for _, h := range hashes {
			if h < min {
				min = h
			}
		}
		return []uint32{min}
	}

	for i := 0; i <= len(hashes)-w; i++ {
		min := hashes[i]
		for j := 1; j < w; j++ {
			if hashes[i+j] < min {
				min = hashes[i+j]
			}
		}
		// Only add if it's not the same as the last one to avoid duplication clustering
		if len(fingerprint) == 0 || fingerprint[len(fingerprint)-1] != min {
			fingerprint = append(fingerprint, min)
		}
	}
	return fingerprint
}

// calculateJaccardSimilarity returns the ratio of intersection over union of two hash sets
func (s *PlagiarismService) calculateJaccardSimilarity(sigA, sigB []uint32) float64 {
	if len(sigA) == 0 && len(sigB) == 0 {
		return 1.0
	}
	if len(sigA) == 0 || len(sigB) == 0 {
		return 0.0
	}

	setA := make(map[uint32]bool)
	setB := make(map[uint32]bool)

	for _, h := range sigA {
		setA[h] = true
	}
	for _, h := range sigB {
		setB[h] = true
	}

	intersection := 0
	for h := range setA {
		if setB[h] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
