package docs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntoDescription(t *testing.T) {
	// Test that the constant is defined and not empty
	assert.NotEmpty(t, intoDescription)

	// Test that it contains expected content
	assert.Contains(t, intoDescription, "CodeArena")
	assert.Contains(t, intoDescription, "IAM Service")
	assert.Contains(t, intoDescription, "Competition & Leagues")
	assert.Contains(t, intoDescription, "Simulation Engine")

	// Test that it contains resource links
	assert.Contains(t, intoDescription, "github.com/codearena-platform")
	assert.Contains(t, intoDescription, "codearena-core")

	// Test that it's properly formatted markdown
	assert.Contains(t, intoDescription, "# Documentation")
	assert.Contains(t, intoDescription, "## Resources")
	assert.Contains(t, intoDescription, "## Core Modules")

	// Test that links are properly formatted
	assert.Contains(t, intoDescription, "[codearena-platform](https://github.com/codearena-platform/codearena-platform)")
	assert.Contains(t, intoDescription, "[codearena-core](https://github.com/codearena-platform/codearena-core)")

	// Test that it contains ELO information
	assert.Contains(t, intoDescription, "ELO & Rankings")

	// Test that it's a reasonable length
	assert.Greater(t, len(intoDescription), 500, "Description should be substantial")
}

func TestIntoDescriptionStructure(t *testing.T) {
	lines := strings.Split(intoDescription, "\n")

	// Should have multiple lines
	assert.Greater(t, len(lines), 10, "Description should have multiple lines")

	// Should start with main heading
	assert.True(t, strings.HasPrefix(strings.TrimSpace(lines[0]), "# Documentation for CodeArena APIs"))

	// Should contain various markdown elements
	var hasLinks, hasList bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "> ") {
			hasList = true
		}
		if strings.Contains(trimmed, "](") {
			hasLinks = true
		}
	}

	assert.True(t, hasLinks, "Description should contain markdown links")
	assert.True(t, hasList, "Description should contain blockquotes/list-like items")
}

func TestIntoDescriptionContent(t *testing.T) {
	// Test specific content sections
	sections := []string{
		"Documentation for CodeArena APIs",
		"Resources",
		"Core Modules",
		"Competition & Leagues",
		"Simulation Engine",
		"ELO & Rankings",
	}

	for _, section := range sections {
		assert.Contains(t, intoDescription, section, "Description should contain section: %s", section)
	}

	// Test technical terms
	technicalTerms := []string{
		"arena", "simulations", "league", "tournament",
		"physics", "sensor", "hazards", "ELO",
	}

	for _, term := range technicalTerms {
		assert.Contains(t, strings.ToLower(intoDescription), strings.ToLower(term),
			"Description should contain technical term: %s", term)
	}
}
